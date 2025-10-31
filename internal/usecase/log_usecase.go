package usecase

import (
	"context"

	"github.com/zuxt268/sales/internal/interfaces/dto/request"
	"github.com/zuxt268/sales/internal/interfaces/dto/response"
	"github.com/zuxt268/sales/internal/interfaces/repository"
	"github.com/zuxt268/sales/internal/model"
)

type LogUsecase interface {
	GetLogs(ctx context.Context, req request.GetLogs) (*response.Logs, error)
	CreateLogs(ctx context.Context, req request.CreateLog) (*response.Log, error)
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

func (u *logUsecase) GetLogs(ctx context.Context, req request.GetLogs) (*response.Logs, error) {
	filter := repository.LogFilter{
		Category: &req.Category,
		Limit:    req.Limit,
		Offset:   req.Offset,
	}
	logs, err := u.logRepo.FindAll(ctx, filter)
	if err != nil {
		return nil, err
	}
	total, err := u.logRepo.Count(ctx, filter)
	if err != nil {
		return nil, err
	}
	return response.GetLogs(logs, total), nil
}

func (u *logUsecase) CreateLogs(ctx context.Context, req request.CreateLog) (*response.Log, error) {
	l := &model.Log{
		Category: req.Category,
		Name:     req.Name,
		Message:  req.Message,
	}
	err := u.logRepo.Create(ctx, l)
	if err != nil {
		return nil, err
	}
	return response.GetLog(l), nil
}
