package usecase

import (
	"context"
	"strings"

	"github.com/zuxt268/sales/internal/interfaces/dto/request"
	"github.com/zuxt268/sales/internal/interfaces/repository"
	"github.com/zuxt268/sales/internal/model"
)

type HomstaUsecase interface {
	CreateHomsta(ctx context.Context, req request.Homsta) error
	GetHomstas(ctx context.Context, limit, offset *int) ([]*model.Homsta, error)
	GetHomsta(ctx context.Context, name string) (*model.Homsta, error)
}

type homstaUsecase struct {
	baseRepo   repository.BaseRepository
	homstaRepo repository.HomstaRepository
}

func NewHomstaUsecase(
	baseRepo repository.BaseRepository,
	homstaRepo repository.HomstaRepository,
) HomstaUsecase {
	return &homstaUsecase{
		baseRepo:   baseRepo,
		homstaRepo: homstaRepo,
	}
}

func (u *homstaUsecase) CreateHomsta(ctx context.Context, req request.Homsta) error {

	exists, err := u.homstaRepo.Exists(ctx, repository.HomstaFilter{
		Path: &req.Path,
	})
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	dbUsage, dbName := getDb(req.DbUsage)

	homsta := &model.Homsta{
		Domain:      getDomain(req.Path),
		BlogName:    req.BlogName,
		Path:        req.Path,
		SiteURL:     req.SiteUrl,
		Description: req.Description,
		Users:       req.Users,
		DBName:      dbName,
		DBUsage:     dbUsage,
		DiscUsage:   req.DiscUsage,
	}

	return u.homstaRepo.Save(ctx, homsta)
}

func getDomain(path string) string {
	paths := strings.Split(path, "/")
	for i := len(paths); i > 0; i-- {
		if paths[i] == "public_html" {
			continue
		}
		return paths[i]
	}
	return ""
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
