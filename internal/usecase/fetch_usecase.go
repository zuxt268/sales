package usecase

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/zuxt268/sales/internal/config"
	"github.com/zuxt268/sales/internal/external"
	"github.com/zuxt268/sales/internal/interfaces/adapter"
	"github.com/zuxt268/sales/internal/interfaces/repository"
	"github.com/zuxt268/sales/internal/model"
)

type FetchUsecase interface {
	Fetch(ctx context.Context, req model.PostFetchRequest)
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

func (u *fetchUsecase) Fetch(ctx context.Context, req model.PostFetchRequest) {
	slog.Info("fetch is invoked")

	target, err := u.targetRepo.GetForUpdate(ctx, repository.TargetFilter{IP: &req.Target})
	if err != nil {
		slog.Error("failed to fetch target", "error", err)
		return
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
			slog.Error("failed get reverse ip", "error", err)
			return
		}

		domains := make([]*model.Domain, 0, len(resp.Response.Domains))
		for _, d := range resp.Response.Domains {
			domains = append(domains, &model.Domain{
				Name:   d.Name,
				Target: target.Name,
				Status: model.StatusInitialize,
			})
		}

		err = u.domainRepo.BulkInsert(ctx, domains)
		if err != nil {
			slog.Error("failed to insert domains", "error", err)
			return
		}

		if maxPage == 0 {
			domainCount := resp.Response.DomainCount
			count, err := strconv.Atoi(domainCount)
			if err != nil {
				slog.Error("failed to parse domain count", "error", err)
				return
			}
			maxPage = (count + 9999) / 10000 // ceil 計算
		}
		if page >= maxPage {
			break
		}
		page++
	}

	target.Status = model.TargetStatusFetched
	err = u.targetRepo.Save(ctx, target)
	if err != nil {
		slog.Error("failed to save target", "error", err)
		return
	}

	slog.Info("fetch success")
}
