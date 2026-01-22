package usecase

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/zuxt268/sales/internal/config"
	"github.com/zuxt268/sales/internal/entity"
	"github.com/zuxt268/sales/internal/interfaces/adapter"
	"github.com/zuxt268/sales/internal/interfaces/dto/external"
	"github.com/zuxt268/sales/internal/interfaces/repository"
	"github.com/zuxt268/sales/internal/model"
	"github.com/zuxt268/sales/internal/util"
	"golang.org/x/exp/slog"
	"golang.org/x/sync/errgroup"
)

type GrowthUsecase interface {
	Fetch(ctx context.Context) error
	Polling(ctx context.Context) error
	Analyze(ctx context.Context, domainMessage *external.DomainMessage) error
	Output(ctx context.Context) error
	FetchWix(ctx context.Context) error
}

type growthUsecase struct {
	baseRepo       repository.BaseRepository
	domainRepo     repository.DomainRepository
	targetRepo     repository.TargetRepository
	pubSubAdapter  adapter.PubSubAdapter
	viewDnsAdapter adapter.ViewDNSAdapter
	sheetAdapter   adapter.SheetAdapter
	gptAdapter     adapter.GptAdapter
}

func NewGrowthUsecase(
	baseRepo repository.BaseRepository,
	domainRepo repository.DomainRepository,
	targetRepo repository.TargetRepository,
	pubSubAdapter adapter.PubSubAdapter,
	viewDnsAdapter adapter.ViewDNSAdapter,
	sheetAdapter adapter.SheetAdapter,
	gptAdapter adapter.GptAdapter,
) GrowthUsecase {
	return &growthUsecase{
		baseRepo:       baseRepo,
		domainRepo:     domainRepo,
		targetRepo:     targetRepo,
		pubSubAdapter:  pubSubAdapter,
		viewDnsAdapter: viewDnsAdapter,
		sheetAdapter:   sheetAdapter,
		gptAdapter:     gptAdapter,
	}
}

const (
	totalCallsPerRun     = 100   // プランに合わせて調整。1日上限より少なく。
	nonWixRatio          = 0.8   // 80% 非WIX, 20% WIX など
	pollingBatchSize     = 300   // Pollingで一度に処理するドメイン数
	maxConcurrentPolling = 20    // Polling時の最大並行処理数
	reverseIPPageSize    = 10000 // viewdns の 1 ページあたり件数
)

// sortTargetsByLastFetched sorts targets by LastFetchedAt (NULL first, then oldest first)
func sortTargetsByLastFetched(targets []*model.Target) {
	sort.Slice(targets, func(i, j int) bool {
		ti, tj := targets[i], targets[j]
		if ti.LastFetchedAt == nil && tj.LastFetchedAt != nil {
			return true
		}
		if ti.LastFetchedAt != nil && tj.LastFetchedAt == nil {
			return false
		}
		if ti.LastFetchedAt == nil && tj.LastFetchedAt == nil {
			return ti.ID < tj.ID
		}
		return ti.LastFetchedAt.Before(*tj.LastFetchedAt)
	})
}

func (u *growthUsecase) Fetch(ctx context.Context) error {

	return u.FetchWix(ctx)

	//// 1. ターゲット全取得（ソートは Filter でもいいし、Go 側でもよい）
	//targets, err := u.targetRepo.FindAll(ctx, repository.TargetFilter{
	//	// OrderLastFetchedAsc は無くてもOK。Go側で sort.Slice しているので。
	//})
	//if err != nil {
	//	return err
	//}
	//
	//var nonWixTargets, wixTargets []*model.Target
	//for _, t := range targets {
	//	if t.Name == "WIX" {
	//		wixTargets = append(wixTargets, t)
	//	} else {
	//		nonWixTargets = append(nonWixTargets, t)
	//	}
	//}
	//
	//// 2. 最近叩いていないものから優先（LastFetchedAt が NULL のものを先）
	//sortTargetsByLastFetched(nonWixTargets)
	//sortTargetsByLastFetched(wixTargets)
	//
	//// 3. レート制限に合わせて「非WIX枠」「WIX枠」を決める
	//nonWixQuota := int(float64(totalCallsPerRun) * nonWixRatio)
	//wixQuota := totalCallsPerRun - nonWixQuota
	//
	//// 4. 非WIXから優先して 1IP = 最大1ページずつ進める
	//for _, t := range nonWixTargets {
	//	if nonWixQuota <= 0 {
	//		break
	//	}
	//	if err := u.fetchOnePage(ctx, t); err != nil {
	//		return err
	//	}
	//	nonWixQuota--
	//}
	//
	//// 5. 余った枠で WIX も 1IP = 最大1ページずつ進める
	//for _, t := range wixTargets {
	//	if wixQuota <= 0 {
	//		break
	//	}
	//	if err := u.fetchOnePage(ctx, t); err != nil {
	//		return err
	//	}
	//	wixQuota--
	//}
	//
	//return nil
}

// 1つの target(IP) について、現在のページを1つだけ進める
// 戻り値 bool は「実際に API を叩いて進捗があったかどうか」です。
func (u *growthUsecase) fetchOnePage(ctx context.Context, target *model.Target) error {
	// 次に叩くページ番号（current_page が 0 以下なら 1 から）
	page := target.CurrentPage
	if page <= 0 {
		page = 1
	}

	// ViewDNS ReverseIP API 呼び出し
	resp, err := u.viewDnsAdapter.GetReverseIP(ctx, &external.ReverseIpRequest{
		Host:   target.IP,
		ApiKey: config.Env.ApiKey,
		Page:   page,
	})
	if err != nil {
		return fmt.Errorf("get reverse ip (ip=%s, page=%d): %w", target.IP, page, err)
	}

	// トランザクション内でドメイン保存とターゲット更新を行う
	if err := u.baseRepo.WithTransaction(ctx, func(ctx context.Context) error {
		// ドメイン保存（domains 側に UNIQUE 制約を貼っておいて、BulkInsert 内部で
		// INSERT IGNORE / ON DUPLICATE KEY UPDATE を使う前提）
		if len(resp.Response.Domains) > 0 {
			for _, d := range resp.Response.Domains {
				exists, err := u.domainRepo.Exists(ctx, repository.DomainFilter{
					Name: &d.Name,
				})
				if err != nil {
					return err
				}
				if exists {
					continue
				}
				if err := u.domainRepo.Save(ctx, &model.Domain{
					Name:   d.Name,
					Target: target.Name,
					Status: model.StatusInitialize,
				}); err != nil {
					return err
				}
			}
		}

		// このタイミングでの DomainCount から「現在の maxPage」を計算（DB には保存しない）
		count, err := strconv.Atoi(resp.Response.DomainCount)
		if err != nil {
			return fmt.Errorf("failed to parse domain count (ip=%s): %w", target.IP, err)
		}
		maxPageNow := (count + reverseIPPageSize - 1) / reverseIPPageSize

		now := time.Now()

		// --- ページ進行ロジック ---
		if target.Name == "WIX" {
			// WIX は「最初からやり直さない」方針：
			// ・とにかく CurrentPage を前に進めていく
			// ・末尾まで来ている場合は maxPageNow に張り付く
			//   → DomainCount が増えて maxPageNow が増えたら、そのうち新しいページに届く
			if maxPageNow <= 0 {
				// ドメインが 0 件のときは次回も page=1 からでOK
				target.CurrentPage = 1
			} else if page >= maxPageNow {
				target.CurrentPage = maxPageNow
			} else {
				target.CurrentPage = page + 1
			}

		} else {
			// 非WIX は「なるべく1周取り切る」モード
			if maxPageNow <= 0 {
				// 0件なら一応「1周完了」とみなして CurrentPage を 1 にリセット
				target.CurrentPage = 1
				t := now
				target.LastFullScanAt = &t
			} else if page >= maxPageNow {
				// 1〜maxPageNow まで一通り読んだので周回完了
				target.CurrentPage = 1
				t := now
				target.LastFullScanAt = &t
			} else {
				target.CurrentPage = page + 1
			}
		}

		target.LastFetchedAt = &now
		if err := u.targetRepo.Save(ctx, target); err != nil {
			return fmt.Errorf("save target (ip=%s): %w", target.IP, err)
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (u *growthUsecase) Polling(ctx context.Context) error {
	domains, err := u.domainRepo.FindAll(ctx, repository.DomainFilter{
		Status: util.Pointer(model.StatusInitialize),
		Limit:  util.Pointer(pollingBatchSize),
	})
	if err != nil {
		return fmt.Errorf("fetch domains: %w", err)
	}

	if len(domains) == 0 {
		slog.Info("no domains to poll")
		return nil
	}

	slog.Info("starting polling", "domain_count", len(domains))

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(maxConcurrentPolling)

	for _, domain := range domains {
		d := domain // ループ変数のキャプチャ
		g.Go(func() error {
			if err := u.handleDomain(ctx, d); err != nil {
				slog.Error("failed to handle domain", "domain_id", d.ID, "error", err)
				return fmt.Errorf("failed to handle domain (domain_id=%d): %w", d.ID, err)
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	slog.Info("polling completed", "domain_count", len(domains))
	return nil
}

func (u *growthUsecase) handleDomain(ctx context.Context, domain *model.Domain) error {
	if err := u.pubSubAdapter.PushDomain(ctx, &external.DomainMessage{
		DomainId: domain.ID,
	}); err != nil {
		return fmt.Errorf("pubsub publish failed: %w", err)
	}

	domain.Status = model.StatusCheckView
	if err := u.domainRepo.Save(ctx, domain); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}
	return nil
}

func (u *growthUsecase) Analyze(ctx context.Context, domainMessage *external.DomainMessage) error {
	slog.Info("analyzing domain", "domainMessage", domainMessage)

	return u.baseRepo.WithTransaction(ctx, func(ctx context.Context) error {
		domain, err := u.domainRepo.GetForUpdate(ctx, repository.DomainFilter{ID: &domainMessage.DomainId})
		if err != nil {
			return err
		}
		if domain.Status != model.StatusCrawlCompInfo {
			return nil
		}
		if err := u.gptAdapter.Analyze(ctx, domain); err != nil {
			return err
		}
		domain.MobilePhone, domain.LandlinePhone = entity.SplitPhone(domain.Phone)
		domain.Status = model.StatusPendingOutput
		if err := u.domainRepo.Save(ctx, domain); err != nil {
			return err
		}
		slog.Info("analyzed", "domain", domain)
		return nil
	})
}

func (u *growthUsecase) Output(ctx context.Context) error {

	domains, err := u.domainRepo.FindAll(ctx, repository.DomainFilter{
		Status: util.Pointer(model.StatusPendingOutput),
	})
	if err != nil {
		return fmt.Errorf("failed to get domains: %w", err)
	}
	if len(domains) == 0 {
		return nil
	}

	domainsByTarget := make(map[string][]*model.Domain)
	for _, d := range domains {
		domainsByTarget[d.Target] = append(domainsByTarget[d.Target], d)
	}

	driveFolderID := config.Env.GoogleDriveBackupFolderID
	if err := u.sheetAdapter.BackupDomainsToGoogleDrive(domainsByTarget, driveFolderID); err != nil {
		return fmt.Errorf("failed to backup domains to Google Drive: %w", err)
	}

	if updateErr := u.domainRepo.BulkUpdateStatus(ctx, model.StatusPendingOutput, model.StatusDone); updateErr != nil {
		return fmt.Errorf("failed to bulk update domain status: %w", updateErr)
	}
	return nil
}

func (u *growthUsecase) FetchWix(ctx context.Context) error {
	// 1. ターゲット全取得（ソートは Filter でもいいし、Go 側でもよい）
	targets, err := u.targetRepo.FindAll(ctx, repository.TargetFilter{
		Name: util.Pointer("WIX"),
	})
	if err != nil {
		return err
	}

	sortTargetsByLastFetched(targets)

	wixQuota := 500
	if len(targets) == 0 {
		return nil
	}

	for wixQuota >= 0 {
		for _, t := range targets {
			if err := u.fetchOnePage(ctx, t); err != nil {
				return err
			}
			wixQuota = wixQuota - 1
		}
	}
	return nil
}
