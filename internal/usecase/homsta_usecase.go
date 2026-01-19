package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/zuxt268/sales/internal/config"
	"github.com/zuxt268/sales/internal/entity"
	"github.com/zuxt268/sales/internal/interfaces/adapter"
	"github.com/zuxt268/sales/internal/interfaces/dto/request"
	"github.com/zuxt268/sales/internal/interfaces/repository"
	"github.com/zuxt268/sales/internal/model"
	"github.com/zuxt268/sales/internal/util"
	"golang.org/x/exp/slog"
)

type HomstaUsecase interface {
	CreateHomsta(ctx context.Context, req request.Homsta) error
	GetHomstas(ctx context.Context, limit, offset *int) ([]*model.Homsta, error)
	GetHomsta(ctx context.Context, name string) (*model.Homsta, error)
	AnalyzeIndustry(ctx context.Context) error
	Output(ctx context.Context) error
	FetchDomainDetails(ctx context.Context) error
	FetchDomains(ctx context.Context) ([]string, error)
}

type homstaUsecase struct {
	baseRepo     repository.BaseRepository
	homstaRepo   repository.HomstaRepository
	sshAdapter   adapter.SSHAdapter
	gptAdapter   adapter.GptAdapter
	sheetAdapter adapter.SheetAdapter
	slackAdapter adapter.SlackAdapter
}

func NewHomstaUsecase(
	baseRepo repository.BaseRepository,
	homstaRepo repository.HomstaRepository,
	sshAdapter adapter.SSHAdapter,
	gptAdapter adapter.GptAdapter,
	sheetAdapter adapter.SheetAdapter,
	slackAdapter adapter.SlackAdapter,
) HomstaUsecase {
	return &homstaUsecase{
		baseRepo:     baseRepo,
		homstaRepo:   homstaRepo,
		sshAdapter:   sshAdapter,
		gptAdapter:   gptAdapter,
		sheetAdapter: sheetAdapter,
		slackAdapter: slackAdapter,
	}
}

func (u *homstaUsecase) CreateHomsta(ctx context.Context, req request.Homsta) error {

	dbUsage, dbName := getDb(req.DbUsage)
	homsta := &model.Homsta{
		Domain:      getDomain(req.SiteUrl),
		BlogName:    req.BlogName,
		Path:        req.Path,
		SiteURL:     req.SiteUrl,
		Description: req.Description,
		Users:       req.Users,
		DBName:      dbName,
		DBUsage:     dbUsage,
		DiscUsage:   req.DiscUsage,
	}
	exists, err := u.homstaRepo.Get(ctx, repository.HomstaFilter{Path: &req.Path})
	if err != nil && !errors.Is(err, entity.ErrNotFound) {
		return err
	}
	if err == nil {
		homsta.ID = exists.ID
		homsta.Industry = exists.Industry
	}
	return u.homstaRepo.Save(ctx, homsta)
}

func getDomain(siteUrl string) string {
	urlStr, err := url.Parse(siteUrl)
	if err != nil {
		return ""
	}
	return urlStr.Host
}

func getDb(dbUsage string) (name, usage string) {
	dbInfo := strings.Split(dbUsage, ":")
	if len(dbInfo) != 2 {
		return "", ""
	}
	return strings.ReplaceAll(dbInfo[0], " ", ""),
		strings.ReplaceAll(dbInfo[1], " ", "")
}

func (u *homstaUsecase) GetHomstas(ctx context.Context, limit, offset *int) ([]*model.Homsta, error) {
	filter := repository.HomstaFilter{
		Limit:  limit,
		Offset: offset,
	}
	return u.homstaRepo.FindAll(ctx, filter)
}

func (u *homstaUsecase) GetHomsta(ctx context.Context, name string) (*model.Homsta, error) {
	filter := repository.HomstaFilter{
		Name: &name,
	}
	return u.homstaRepo.Get(ctx, filter)
}

func getCompInfo(siteUrl string) (string, error) {
	u, err := url.Parse(siteUrl)
	if err != nil {
		return "", err
	}

	targetUrl := u.String()
	if get(targetUrl + "/service") {
		targetUrl = targetUrl + "/service"
	} else if get(siteUrl) {
		// nothing to do
	} else {
		return "", errors.New("site is unavailable")
	}

	resp, err := http.Get(targetUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	body := doc.Find("body")
	body.Find("script,style,link,noscript").Remove()

	text := strings.Join(strings.Fields(body.Text()), " ")
	return text, nil
}

func get(u string) bool {
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return false
	}
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return false
	}
	return true
}

func (u *homstaUsecase) AnalyzeIndustry(ctx context.Context) error {
	domains, err := u.homstaRepo.FindAll(ctx, repository.HomstaFilter{
		Industry:       util.Pointer(""),
		NotDomainEmpty: util.Pointer(true),
	})
	if err != nil {
		return err
	}
	fmt.Println("対象ドメイン", len(domains))
	for _, domain := range domains {
		text, err := getCompInfo(domain.SiteURL)
		if err != nil {
			fmt.Println(domain.SiteURL, err)
			continue
		}
		text = fmt.Sprintf("サイト名: %s, ディスクリプション: %s", domain.BlogName, domain.Description) + text
		industry, err := u.gptAdapter.AnalyzeSiteIndustry(ctx, text)
		if err != nil {
			fmt.Println(domain.SiteURL, err)
			continue
		}
		domain.Industry = industry
		if err := u.homstaRepo.Save(ctx, domain); err != nil {
			return err
		}
		fmt.Println(domain.Domain, domain.Industry)
	}
	return nil
}

func (u *homstaUsecase) Output(ctx context.Context) error {
	domains, err := u.homstaRepo.FindAll(ctx, repository.HomstaFilter{
		NotDomainEmpty: util.Pointer(true),
		OrderBy:        []string{"industry"},
	})
	if err != nil {
		return err
	}

	rows := make([][]interface{}, 0, len(domains)+1)
	rows = append(rows, []interface{}{
		"ドメイン",
		"サーバーディレクトリ",
		"URL",
		"サイト名",
		"ディスクリプション",
		"業種",
		"データベース名",
		"データベース使用量(MB)",
		"ディスク使用量(MB)",
		"ユーザー",
	})
	for _, d := range domains {
		rows = append(rows, []interface{}{
			d.Domain,
			d.Path,
			d.SiteURL,
			d.BlogName,
			d.Description,
			d.Industry,
			d.DBName,
			d.GetDbUsage(),
			d.GetDiscUsage(),
			d.Users,
		})
	}

	if err := u.sheetAdapter.Output(config.Env.SiteSheetID, "サイト一覧", rows); err != nil {
		return err
	}
	return nil
}

func (u *homstaUsecase) FetchDomainDetails(ctx context.Context) error {
	serverIDs := strings.Split(config.Env.ServerIDs, ",")

	type result struct {
		out string
		err error
	}

	ch := make(chan result, len(serverIDs))
	var wg sync.WaitGroup
	for _, serverID := range serverIDs {
		wg.Add(1)

		go func() {
			defer wg.Done()
			sshConf, err := config.GetSSHConfig(serverID)
			if err != nil {
				slog.Error("SSH設定取得失敗", "serverID", serverID, "error", err.Error())
				ch <- result{err: err}
				return
			}

			cmd := "walk fetchDomainDetails"
			stdout, err := u.sshAdapter.RunOutput(sshConf, cmd)
			if err != nil {
				ch <- result{err: err}
				return
			}
			ch <- result{out: stdout}
		}()
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	existsPaths := make([]string, 0, 4048)
	var firstErr error

	for r := range ch {
		if r.err != nil {
			if firstErr == nil {
				firstErr = r.err
			}
			continue
		}

		var partial []entity.DomainDetails
		if err := json.Unmarshal([]byte(r.out), &partial); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		fmt.Println("partial", len(partial))

		for _, d := range partial {
			dbName, dbUsage := getDb(d.DBUsage)
			homsta := &model.Homsta{
				Domain:      getDomain(d.SiteUrl),
				BlogName:    d.BlogName,
				Path:        d.Path,
				SiteURL:     d.SiteUrl,
				Description: d.Description,
				Users:       d.Users,
				DBName:      dbName,
				DBUsage:     dbUsage,
				DiscUsage:   d.DiscUsage,
			}
			existsPaths = append(existsPaths, homsta.Path)
			exists, err := u.homstaRepo.FindAll(ctx, repository.HomstaFilter{
				Path: &d.Path,
			})
			if err != nil {
				return err
			}
			updated := false
			if len(exists) != 0 {
				exist := exists[0]

				if exist.Domain == homsta.Domain &&
					exist.DBName == dbName &&
					exist.DBUsage == dbUsage &&
					exist.Path == d.Path &&
					exist.Description == d.Description &&
					exist.DiscUsage == d.DiscUsage &&
					exist.Users == d.Users &&
					exist.SiteURL == d.SiteUrl &&
					exist.BlogName == d.BlogName {
					continue
				}
				updated = true
				homsta.ID = exist.ID
				homsta.Industry = exist.Industry
				homsta.CreatedAt = exist.CreatedAt
			}

			err = u.homstaRepo.Save(ctx, homsta)
			if err != nil {
				fmt.Println(err)
				continue
			}
			if updated {
				fmt.Println("updated", homsta.Path)
			} else {
				fmt.Println("created", homsta.Path)
			}
		}
	}
	if firstErr != nil {
		return firstErr
	}

	domains, err := u.homstaRepo.FindAll(ctx, repository.HomstaFilter{})
	if err != nil {
		return err
	}
	for _, d := range domains {
		if !slices.Contains(existsPaths, d.Path) {
			if err := u.homstaRepo.Delete(ctx, repository.HomstaFilter{
				Path: &d.Path,
			}); err != nil {
				fmt.Println(err)
			}
		}
	}

	return nil
}

func (u *homstaUsecase) FetchDomains(ctx context.Context) ([]string, error) {

	var domains []string
	serverIDs := strings.Split(config.Env.ServerIDs, ",")
	for _, serverID := range serverIDs {
		sshConf, err := config.GetSSHConfig(serverID)
		if err != nil {
			slog.Error("SSH設定取得失敗", "error", err.Error())
			return nil, err
		}

		cmd := "walk fetchDomains"

		stdout, err := u.sshAdapter.RunOutput(sshConf, cmd)
		if err != nil {
			return nil, err
		}

		var result []string
		if err := json.Unmarshal([]byte(stdout), &result); err != nil {
			return nil, err
		}
		domains = append(domains, result...)
	}
	return domains, nil
}
