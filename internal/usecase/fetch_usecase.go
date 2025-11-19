package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"sync"

	"github.com/zuxt268/sales/internal/config"
	"github.com/zuxt268/sales/internal/interfaces/adapter"
	"github.com/zuxt268/sales/internal/interfaces/dto/external"
	"github.com/zuxt268/sales/internal/interfaces/repository"
	"github.com/zuxt268/sales/internal/model"
	"github.com/zuxt268/sales/internal/util"
)

type FetchUsecase interface {
	Polling(ctx context.Context)
	Fetch(ctx context.Context)
}

type fetchUsecase struct {
	viewDnsAdapter adapter.ViewDNSAdapter
	slackAdapter   adapter.SlackAdapter
	pubSubAdapter  adapter.PubSubAdapter
	domainRepo     repository.DomainRepository
	targetRepo     repository.TargetRepository
}

func NewFetchUsecase(
	viewDnsAdapter adapter.ViewDNSAdapter,
	slackAdapter adapter.SlackAdapter,
	pubSubAdapter adapter.PubSubAdapter,
	domainRepo repository.DomainRepository,
	targetRepo repository.TargetRepository,
) FetchUsecase {
	return &fetchUsecase{
		viewDnsAdapter: viewDnsAdapter,
		slackAdapter:   slackAdapter,
		pubSubAdapter:  pubSubAdapter,
		domainRepo:     domainRepo,
		targetRepo:     targetRepo,
	}
}

func (u *fetchUsecase) Polling(ctx context.Context) {
	slog.Info("Polling is invoked")

	domains, err := u.domainRepo.FindAll(ctx, repository.DomainFilter{
		Status: util.Pointer(model.StatusInitialize),
	})
	if err != nil {
		slog.Error("Error fetching domains", slog.Any("error", err))
		return
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, 20) // 同時10件まで

	for _, domain := range domains {
		wg.Add(1)
		go func(d *model.Domain) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			if err := u.handleDomain(ctx, d); err != nil {
				slog.Error("failed to process domain",
					slog.Int("domain_id", d.ID),
					slog.Any("error", err),
				)
			}
		}(domain)
	}

	wg.Wait()
	slog.Info("Polling finished")
}

func (u *fetchUsecase) handleDomain(ctx context.Context, domain *model.Domain) error {
	if err := u.pubSubAdapter.PushDomain(ctx, &external.DomainMessage{
		DomainId: domain.ID,
	}); err != nil {
		return fmt.Errorf("pubsub publish failed: %w", err)
	}

	domain.Status = model.StatusCheckView
	if err := u.domainRepo.Save(ctx, domain); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}
	return nil
}

func (u *fetchUsecase) Fetch(ctx context.Context) {
	slog.Info("fetch is invoked")

	targets, err := u.targetRepo.FindAll(ctx, repository.TargetFilter{
		NotName: util.Pointer("WIX"),
	})
	if err != nil {
		slog.Error("failed to fetch target", "error", err)
		return
	}

	for _, target := range targets {
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
			slog.Info("insert domains", "domain_count", len(domains), "name", target.Name)

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
	}

	slog.Info("fetch success")
}
