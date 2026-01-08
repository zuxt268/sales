package usecase

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/zuxt268/sales/internal/entity"
	"github.com/zuxt268/sales/internal/interfaces/adapter"
	"github.com/zuxt268/sales/internal/interfaces/dto/request"
	"github.com/zuxt268/sales/internal/interfaces/repository"
	"github.com/zuxt268/sales/internal/model"
	"github.com/zuxt268/sales/internal/util"
)

type HomstaUsecase interface {
	CreateHomsta(ctx context.Context, req request.Homsta) error
	GetHomstas(ctx context.Context, limit, offset *int) ([]*model.Homsta, error)
	GetHomsta(ctx context.Context, name string) (*model.Homsta, error)
	AnalyzeIndustry(ctx context.Context) error
}

type homstaUsecase struct {
	baseRepo   repository.BaseRepository
	homstaRepo repository.HomstaRepository
	gptAdapter adapter.GptAdapter
}

func NewHomstaUsecase(
	baseRepo repository.BaseRepository,
	homstaRepo repository.HomstaRepository,
	gptAdapter adapter.GptAdapter,
) HomstaUsecase {
	return &homstaUsecase{
		baseRepo:   baseRepo,
		homstaRepo: homstaRepo,
		gptAdapter: gptAdapter,
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
	u.Path = path.Join(u.Path, "service")

	resp, err := http.Get(u.String())
	if err != nil {
		return "", err
	}
	if resp.StatusCode == 404 {
		resp.Body.Close()
		resp, err = http.Get(u.String())
		if err != nil {
			return "", err
		}
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

func (u *homstaUsecase) AnalyzeIndustry(ctx context.Context) error {
	domains, err := u.homstaRepo.FindAll(ctx, repository.HomstaFilter{
		Industry:       util.Pointer(""),
		NotDomainEmpty: util.Pointer(true),
	})
	if err != nil {
		return err
	}
	for _, domain := range domains {
		text, err := getCompInfo(domain.SiteURL)
		if err != nil {
			fmt.Println(domain.SiteURL, err)
		}
		industry, err := u.gptAdapter.AnalyzeSiteIndustry(ctx, text)
		if err != nil {
			return err
		}
		domain.Industry = industry
		if err := u.homstaRepo.Save(ctx, domain); err != nil {
			return err
		}
		fmt.Println(domain.Domain, domain.Industry)
	}
	return nil
}
