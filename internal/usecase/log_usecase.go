package usecase

import (
	"context"

	"github.com/zuxt268/sales/internal/domain"
	"github.com/zuxt268/sales/internal/interfaces/repository"
)

type LogUsecase interface {
	GetLogs(ctx context.Context, req domain.GetLogsRequest) ([]domain.Log, error)
	CreateLogs(ctx context.Context, req domain.CreateLogRequest) (*domain.Log, error)
}

type logUsecase struct {
	baseRepo repository.BaseRepository
	logRepo  repository.LogRepository
}

func NewLogUsecase(
	baseRepo repository.BaseRepository,
	logRepo repository.LogRepository,
) LogUsecase {
	return &logUsecase{
		baseRepo: baseRepo,
		logRepo:  logRepo,
	}
}

func (u *logUsecase) GetLogs(ctx context.Context, req domain.GetLogsRequest) ([]domain.Log, error) {
	return u.logRepo.FindAll(ctx, repository.LogFilter{
		Category: req.Category,
		Limit:    req.Limit,
		Offset:   req.Offset,
	})
}

func (u *logUsecase) CreateLogs(ctx context.Context, req domain.CreateLogRequest) (*domain.Log, error) {
	log := &domain.Log{
		Name:     req.Name,
		Category: req.Category,
		Message:  req.Message,
	}

	err := u.logRepo.Create(ctx, log)
	if err != nil {
		return nil, err
	}

	return log, nil
}
