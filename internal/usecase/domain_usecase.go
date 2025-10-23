package usecase

import (
	"context"
	"unicode/utf8"

	"github.com/zuxt268/sales/internal/domain"
	"github.com/zuxt268/sales/internal/interfaces/repository"
)

type DomainUsecase interface {
	GetDomains(ctx context.Context, req domain.GetDomainsRequest) ([]domain.Domain, error)
	UpdateDomain(ctx context.Context, id int, req domain.UpdateDomainRequest) (*domain.Domain, error)
	DeleteDomain(ctx context.Context, id int) error
}

type domainUsecase struct {
	baseRepo   repository.BaseRepository
	domainRepo repository.DomainRepository
}

func NewDomainUsecase(
	baseRepo repository.BaseRepository,
	domainRepo repository.DomainRepository,
) DomainUsecase {
	return &domainUsecase{
		baseRepo:   baseRepo,
		domainRepo: domainRepo,
	}
}

func (u *domainUsecase) GetDomains(ctx context.Context, req domain.GetDomainsRequest) ([]domain.Domain, error) {
	return u.domainRepo.FindAll(ctx, repository.DomainFilter{
		PartialName: req.Name,
		CanView:     req.CanView,
		IsJapan:     req.IsJapan,
		IsSend:      req.IsSend,
		OwnerID:     req.OwnerID,
		Industry:    req.Industry,
		IsSSL:       req.IsSSL,
		Status:      req.Status,
		Limit:       req.Limit,
		Offset:      req.Offset,
	})
}

func (u *domainUsecase) UpdateDomain(ctx context.Context, id int, req domain.UpdateDomainRequest) (*domain.Domain, error) {
	var target domain.Domain
	err := u.baseRepo.WithTransaction(ctx, func(ctx context.Context) error {
		var err error
		target, err = u.domainRepo.GetForUpdate(ctx, repository.DomainFilter{
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
		if req.IsJapan != nil {
			target.IsJapan = *req.IsJapan
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
		if req.MobilePhone != nil {
			target.MobilePhone = *req.MobilePhone
		}
		if req.LandlinePhone != nil {
			target.LandlinePhone = *req.LandlinePhone
		}
		if req.Industry != nil {
			target.Industry = *req.Industry
		}
		if req.IsSSL != nil {
			target.IsSSL = *req.IsSSL
		}
		if req.RawPage != nil {
			rawPage := *req.RawPage
			if utf8.RuneCountInString(rawPage) > 8000 {
				runes := []rune(rawPage)
				rawPage = string(runes[:8000])
			}
			target.RawPage = rawPage
		}
		if req.PageNum != nil {
			target.PageNum = *req.PageNum
		}
		if err := u.domainRepo.Save(ctx, &target); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &target, nil
}

func (u *domainUsecase) DeleteDomain(ctx context.Context, id int) error {
	return u.domainRepo.Delete(ctx, repository.DomainFilter{
		ID: &id,
	})
}
