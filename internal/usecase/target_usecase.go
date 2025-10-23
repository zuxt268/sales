package usecase

import (
	"context"

	"github.com/zuxt268/sales/internal/domain"
	"github.com/zuxt268/sales/internal/interfaces/repository"
)

type TargetUsecase interface {
	GetTargets(ctx context.Context, req domain.GetTargetsRequest) ([]domain.Target, error)
	CreateTarget(ctx context.Context, req domain.CreateTargetRequest) (*domain.Target, error)
	UpdateTarget(ctx context.Context, id int, req domain.UpdateTargetRequest) (*domain.Target, error)
	DeleteTarget(ctx context.Context, id int) error
}

type targetUsecase struct {
	baseRepo   repository.BaseRepository
	targetRepo repository.TargetRepository
}

func NewTargetUsecase(
	baseRepo repository.BaseRepository,
	targetRepo repository.TargetRepository,
) TargetUsecase {
	return &targetUsecase{
		baseRepo:   baseRepo,
		targetRepo: targetRepo,
	}
}

func (u *targetUsecase) GetTargets(ctx context.Context, req domain.GetTargetsRequest) ([]domain.Target, error) {
	return u.targetRepo.FindAll(ctx, repository.TargetFilter{
		Limit:  req.Limit,
		Offset: req.Offset,
	})
}

func (u *targetUsecase) CreateTarget(ctx context.Context, req domain.CreateTargetRequest) (*domain.Target, error) {
	target := &domain.Target{
		IP:     req.IP,
		Name:   req.Name,
		Status: domain.TargetStatusInit,
	}

	err := u.targetRepo.Save(ctx, target)
	if err != nil {
		return nil, err
	}

	return target, nil
}

func (u *targetUsecase) UpdateTarget(ctx context.Context, id int, req domain.UpdateTargetRequest) (*domain.Target, error) {
	var target domain.Target
	err := u.baseRepo.WithTransaction(ctx, func(ctx context.Context) error {
		var err error
		target, err = u.targetRepo.GetForUpdate(ctx, repository.TargetFilter{
			ID: &id,
		})
		if err != nil {
			return err
		}

		if req.IP != nil {
			target.IP = *req.IP
		}
		if req.Name != nil {
			target.Name = *req.Name
		}
		if err := u.targetRepo.Save(ctx, &target); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &target, nil
}

func (u *targetUsecase) DeleteTarget(ctx context.Context, id int) error {
	return u.targetRepo.Delete(ctx, repository.TargetFilter{
		ID: &id,
	})
}
