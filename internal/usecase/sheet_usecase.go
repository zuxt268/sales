package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/zuxt268/sales/internal/config"
	"github.com/zuxt268/sales/internal/interfaces/adapter"
	"github.com/zuxt268/sales/internal/interfaces/dto/external"
	"github.com/zuxt268/sales/internal/interfaces/repository"
	"github.com/zuxt268/sales/internal/model"
	"github.com/zuxt268/sales/internal/util"
)

type SheetUsecase interface {
	RivalSheetOutput(ctx context.Context) error
	Assort(ctx context.Context)
}

type sheetUsecase struct {
	baseRepo     repository.BaseRepository
	domainRepo   repository.DomainRepository
	sheetAdapter adapter.SheetAdapter
	sshAdapter   adapter.SSHAdapter
}

func NewSheetUsecase(
	baseRepo repository.BaseRepository,
	domainRepo repository.DomainRepository,
	sheetAdapter adapter.SheetAdapter,
	sshAdapter adapter.SSHAdapter,
) SheetUsecase {
	return &sheetUsecase{
		baseRepo:     baseRepo,
		domainRepo:   domainRepo,
		sheetAdapter: sheetAdapter,
		sshAdapter:   sshAdapter,
	}
}

func (s *sheetUsecase) RivalSheetOutput(ctx context.Context) error {
	// ステータスが"done"のドメインを全て取得
	domains, err := s.domainRepo.FindAll(ctx, repository.DomainFilter{
		Status: util.Pointer(model.StatusDone),
	})
	if err != nil {
		return err
	}

	// ターゲットごとにドメインをグループ化
	results := make(map[string][]*model.Domain)
	for _, d := range domains {
		results[d.Target] = append(results[d.Target], d)
	}

	rivalSheetID := config.Env.SheetID

	// 各ターゲットごとにスプレッドシートに出力
	var errors []error
	for target, domains := range results {
		rows := external.GetRows(domains)
		if err := s.sheetAdapter.Output(rivalSheetID, target, rows); err != nil {
			// エラーを収集して処理を継続（全ターゲットを処理）
			errors = append(errors, err)
		}
	}

	// エラーがあれば最初のエラーを返す
	if len(errors) > 0 {
		return errors[0]
	}

	return nil
}

// DomainInfo assortの出力構造体
type DomainInfo struct {
	Domain          string `json:"domain"`
	IsWordPress     bool   `json:"is_wordpress"`
	WordPressCount  int    `json:"wordpress_count"`
	IsHomsta        bool   `json:"is_homsta"`
	HasPublicHTML   bool   `json:"has_public_html"`
	UnderPublicHTML bool   `json:"under_public_html"`
	IsMultisite     bool   `json:"is_multisite"`
}

func (u *sheetUsecase) Assort(ctx context.Context) {
	slog.Info("Assort 処理開始")

	serverIDs := strings.Split(config.Env.ServerIDs, ",")
	for _, serverID := range serverIDs {
		sshConf, err := config.GetSSHConfig(serverID)
		if err != nil {
			slog.Error("SSH設定取得失敗", "error", err.Error())
			return
		}

		// コマンド実行（--json で出力をJSON配列として受け取る）
		cmd := "walk assort --json"

		stdout, err := u.sshAdapter.RunOutput(sshConf, cmd)
		if err != nil {
			slog.Error("walk assort 実行失敗", "error", err.Error())
			return
		}

		// JSONデコード
		var results []DomainInfo
		if err := json.Unmarshal([]byte(stdout), &results); err != nil {
			slog.Error("JSONデコード失敗", "error", err.Error())
			return
		}

		slog.Info("Assort結果取得完了", "count", len(results))

		rows := make([][]interface{}, 0, len(results)+1)
		rows = append(rows, []interface{}{
			"ドメイン",
			"WordPressがある",
			"WordPressの数",
			"ホムスタ案件である",
			"public_htmlがある",
			"public_htmlにwordpressがある",
			"マルチサイトである",
		})

		// 結果をログまたはDBに保存
		for _, r := range results {
			msg := fmt.Sprintf("[Assort] domain=%s wp=%v num=%d homsta=%v pub=%v underPub=%v multi=%v",
				r.Domain, r.IsWordPress, r.WordPressCount, r.IsHomsta, r.HasPublicHTML, r.UnderPublicHTML, r.IsMultisite)
			slog.Info(msg)
			if strings.HasPrefix(r.Domain, ".") {
				continue
			}
			rows = append(rows, []interface{}{
				r.Domain,
				r.IsWordPress,
				r.WordPressCount,
				r.IsHomsta,
				r.HasPublicHTML,
				r.UnderPublicHTML,
				r.IsMultisite,
			})
		}

		if err := u.sheetAdapter.Output(config.Env.SiteSheetID, serverID, rows); err != nil {
			slog.Error(err.Error())
		}
	}

	slog.Info("Assort処理完了")
}
