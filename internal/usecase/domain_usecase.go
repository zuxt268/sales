package usecase

import (
	"context"
	"unicode/utf8"

	"github.com/zuxt268/sales/internal/interfaces/dto/request"
	"github.com/zuxt268/sales/internal/interfaces/dto/response"
	"github.com/zuxt268/sales/internal/interfaces/repository"
	"github.com/zuxt268/sales/internal/model"
)

type DomainUsecase interface {
	GetDomains(ctx context.Context, req request.GetDomains) (*response.Domains, error)
	GetDomain(ctx context.Context, id int) (*response.Domain, error)
	UpdateDomain(ctx context.Context, id int, req request.UpdateDomain) (*response.Domain, error)
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

func (u *domainUsecase) GetDomain(ctx context.Context, id int) (*response.Domain, error) {
	d, err := u.domainRepo.Get(ctx, repository.DomainFilter{ID: &id})
	if err != nil {
		return nil, err
	}
	return response.GetDomain(d), nil
}

func (u *domainUsecase) GetDomains(ctx context.Context, req request.GetDomains) (*response.Domains, error) {
	filter := repository.DomainFilter{
		PartialName: req.Name,
		Target:      req.Target,
		CanView:     req.CanView,
		IsJapan:     req.IsJapan,
		IsSend:      req.IsSend,
		OwnerID:     req.OwnerID,
		Industry:    req.Industry,
		IsSSL:       req.IsSSL,
		Status:      req.Status,
		Limit:       req.Limit,
		Offset:      req.Offset,
	}
	domains, err := u.domainRepo.FindAll(ctx, filter)
	if err != nil {
		return nil, err
	}
	total, err := u.domainRepo.Count(ctx, filter)
	if err != nil {
		return nil, err
	}
	return response.GetDomains(domains, total), nil
}

func (u *domainUsecase) UpdateDomain(ctx context.Context, id int, req request.UpdateDomain) (*response.Domain, error) {
	var target model.Domain
	err := u.baseRepo.WithTransaction(ctx, func(ctx context.Context) error {
		var err error
		target, err = u.domainRepo.GetForUpdate(ctx, repository.DomainFilter{
			ID: &id,
		})
		if err != nil {
			return err
		}
		if req.Status != nil {
			target.Status = *req.Status
		}
		if req.IsSend != nil {
			target.IsSend = *req.IsSend
		}
		if req.Target != nil {
			target.Target = *req.Target
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
			address := *req.Address
			if utf8.RuneCountInString(address) > 300 {
				runes := []rune(address)
				address = string(runes[:300])
			}
			target.Address = address
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

	return response.GetDomain(target), nil
}

func (u *domainUsecase) DeleteDomain(ctx context.Context, id int) error {
	return u.domainRepo.Delete(ctx, repository.DomainFilter{
		ID: &id,
	})
}
