package usecase

import (
	"context"
	"unicode/utf8"

	"github.com/zuxt268/sales/internal/domain"
	"github.com/zuxt268/sales/internal/interfaces/repository"
)

type PageUsecase interface {
	GetDomains(ctx context.Context, req domain.GetDomainsRequest) ([]domain.Domain, error)
	UpdateDomain(ctx context.Context, id int, req domain.UpdateDomainRequest) (*domain.Domain, error)
	DeleteDomain(ctx context.Context, id int) error
	GetTargets(ctx context.Context, req domain.GetTargetsRequest) ([]domain.Target, error)
	CreateTarget(ctx context.Context, req domain.CreateTargetRequest) (*domain.Target, error)
	UpdateTarget(ctx context.Context, req domain.UpdateTargetRequest) (*domain.Target, error)
	DeleteTarget(ctx context.Context, id int) error
	GetLogs(ctx context.Context, req domain.GetLogsRequest) ([]domain.Log, error)
	CreateLogs(ctx context.Context, req domain.CreateLogRequest) (*domain.Log, error)
}

type pageUsecase struct {
	baseRepo   repository.BaseRepository
	domainRepo repository.DomainRepository
	targetRepo repository.TargetRepository
	logRepo    repository.LogRepository
}

func NewPageUsecase(
	baseRepo repository.BaseRepository,
	domainRepo repository.DomainRepository,
	targetRepo repository.TargetRepository,
	logRepo repository.LogRepository,
) PageUsecase {
	return &pageUsecase{
		baseRepo:   baseRepo,
		domainRepo: domainRepo,
		targetRepo: targetRepo,
		logRepo:    logRepo,
	}
}

func (p *pageUsecase) GetDomains(ctx context.Context, req domain.GetDomainsRequest) ([]domain.Domain, error) {
	return p.domainRepo.FindAll(ctx, repository.DomainFilter{
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

func (p *pageUsecase) GetTargets(ctx context.Context, req domain.GetTargetsRequest) ([]domain.Target, error) {
	return p.targetRepo.FindAll(ctx, repository.TargetFilter{
		Limit:  req.Limit,
		Offset: req.Offset,
	})
}

func (p *pageUsecase) CreateTarget(ctx context.Context, req domain.CreateTargetRequest) (*domain.Target, error) {
	target := &domain.Target{
		IP:     req.IP,
		Name:   req.Name,
		Status: domain.TargetStatusInit,
	}

	err := p.targetRepo.Save(ctx, target)
	if err != nil {
		return nil, err
	}

	return target, nil
}

func (p *pageUsecase) UpdateTarget(ctx context.Context, req domain.UpdateTargetRequest) (*domain.Target, error) {
	var target domain.Target
	err := p.baseRepo.WithTransaction(ctx, func(ctx context.Context) error {
		var err error
		ip := req.IP
		target, err = p.targetRepo.GetForUpdate(ctx, repository.TargetFilter{
			IP: &ip,
		})
		if err != nil {
			return err
		}

		target.Name = req.Name

		if err := p.targetRepo.Save(ctx, &target); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &target, nil
}

func (p *pageUsecase) DeleteTarget(ctx context.Context, id int) error {
	return p.targetRepo.Delete(ctx, repository.TargetFilter{
		ID: &id,
	})
}

func (p *pageUsecase) GetLogs(ctx context.Context, req domain.GetLogsRequest) ([]domain.Log, error) {
	return p.logRepo.FindAll(ctx, repository.LogFilter{
		Category: req.Category,
		Limit:    req.Limit,
		Offset:   req.Offset,
	})
}

func (p *pageUsecase) CreateLogs(ctx context.Context, req domain.CreateLogRequest) (*domain.Log, error) {
	log := &domain.Log{
		Name:     req.Name,
		Category: req.Category,
		Message:  req.Message,
	}

	err := p.logRepo.Create(ctx, log)
	if err != nil {
		return nil, err
	}

	return log, nil
}
