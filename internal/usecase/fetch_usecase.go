package usecase

import (
	"context"
	"strconv"

	"github.com/zuxt268/sales/internal/config"
	"github.com/zuxt268/sales/internal/domain"
	"github.com/zuxt268/sales/internal/external"
	"github.com/zuxt268/sales/internal/interfaces/adapter"
	"github.com/zuxt268/sales/internal/interfaces/repository"
)

type FetchUsecase interface {
	Fetch(ctx context.Context, req domain.PostFetchRequest) error
}

type fetchUsecase struct {
	viewDnsAdapter adapter.ViewDNSAdapter
	slackAdapter   adapter.SlackAdapter
	domainRepo     repository.DomainRepository
	targetRepo     repository.TargetRepository
}

func NewFetchUsecase(
	viewDnsAdapter adapter.ViewDNSAdapter,
	slackAdapter adapter.SlackAdapter,
	domainRepo repository.DomainRepository,
	targetRepo repository.TargetRepository,
) FetchUsecase {
	return &fetchUsecase{
		viewDnsAdapter: viewDnsAdapter,
		slackAdapter:   slackAdapter,
		domainRepo:     domainRepo,
		targetRepo:     targetRepo,
	}
}

func (u *fetchUsecase) Fetch(ctx context.Context, req domain.PostFetchRequest) error {

	target, err := u.targetRepo.GetForUpdate(ctx, repository.TargetFilter{IP: &req.Target})
	if err != nil {
		return err
	}

	page := 1
	maxPage := 0
	for {
		resp, err := u.viewDnsAdapter.GetReverseIP(ctx, &external.ReverseIpRequest{
			Host:   target.IP,
			ApiKey: config.Env.ApiKey,
			Page:   page,
		})
		if err != nil {
			return err
		}

		domains := make([]*domain.Domain, 0, len(resp.Response.Domains))
		for _, d := range resp.Response.Domains {
			domains = append(domains, &domain.Domain{
				Name:   d.Name,
				Target: target.Name,
				Status: domain.StatusInitialize,
			})
		}

		err = u.domainRepo.BulkInsert(ctx, domains)
		if err != nil {
			return err
		}

		if maxPage == 0 {
			domainCount := resp.Response.DomainCount
			count, err := strconv.Atoi(domainCount)
			if err != nil {
				return err
			}
			maxPage = (count + 9999) / 10000 // ceil 計算
		}
		if page >= maxPage {
			break
		}
		page++
	}

	target.Status = domain.TargetStatusFetched
	err = u.targetRepo.Save(ctx, &target)
	if err != nil {
		return err
	}

	return nil
}
