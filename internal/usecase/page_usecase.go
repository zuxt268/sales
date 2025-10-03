package usecase

import (
	"context"

	"github.com/zuxt268/sales/internal/domain"
	"github.com/zuxt268/sales/internal/interfaces/repository"
)

type PageUsecase interface {
	GetDomains(ctx context.Context, req domain.GetDomainsRequest) ([]domain.Domain, error)
	UpdateDomain(ctx context.Context, id int, req domain.UpdateDomainRequest) (*domain.Domain, error)
	DeleteDomain(ctx context.Context, id int) error
}

type pageUsecase struct {
	baseRepo   repository.BaseRepository
	domainRepo repository.DomainRepository
}

func NewPageUsecase(
	baseRepo repository.BaseRepository,
	domainRepo repository.DomainRepository,
) PageUsecase {
	return &pageUsecase{
		baseRepo:   baseRepo,
		domainRepo: domainRepo,
	}
}

func (p *pageUsecase) GetDomains(ctx context.Context, req domain.GetDomainsRequest) ([]domain.Domain, error) {
	return p.domainRepo.FindAll(ctx, repository.DomainFilter{
		PartialName: req.Name,
		CanView:     req.CanView,
		IsSend:      req.IsSend,
		OwnerID:     req.OwnerID,
		Industry:    req.Industry,
		IsSSL:       req.IsSSL,
		Status:      req.Status,
		Limit:       req.Limit,
		Offset:      req.Offset,
	})
}

func (p *pageUsecase) UpdateDomain(ctx context.Context, id int, req domain.UpdateDomainRequest) (*domain.Domain, error) {
	var target domain.Domain
	err := p.baseRepo.WithTransaction(ctx, func(ctx context.Context) error {
		var err error
		target, err = p.domainRepo.GetForUpdate(ctx, repository.DomainFilter{
			ID: &id,
		})
		if err != nil {
			return err
		}
		target.Status = domain.Status(req.Status)
		if req.IsSend != nil {
			target.IsSend = *req.IsSend
		}
		if req.CanView != nil {
			target.CanView = *req.CanView
		}
		if req.Title != nil {
			target.Title = *req.Title
		}
		if req.OwnerID != nil {
			target.OwnerID = *req.OwnerID
		}
		if req.Address != nil {
			target.Address = *req.Address
		}
		if req.Phone != nil {
			target.Phone = *req.Phone
		}
		if req.Industry != nil {
			target.Industry = *req.Industry
		}
		if req.IsSSL != nil {
			target.IsSSL = *req.IsSSL
		}
		if req.RawPage != nil {
			target.RawPage = *req.RawPage
		}
		if req.PageNum != nil {
			target.PageNum = *req.PageNum
		}
		if err := p.domainRepo.Save(ctx, &target); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &target, nil
}

func (p *pageUsecase) DeleteDomain(ctx context.Context, id int) error {
	return p.domainRepo.Delete(ctx, repository.DomainFilter{
		ID: &id,
	})
}
