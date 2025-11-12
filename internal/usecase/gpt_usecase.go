package usecase

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/zuxt268/sales/internal/entity"
	"github.com/zuxt268/sales/internal/interfaces/adapter"
	"github.com/zuxt268/sales/internal/interfaces/dto/external"
	"github.com/zuxt268/sales/internal/interfaces/repository"
	"github.com/zuxt268/sales/internal/model"
	"github.com/zuxt268/sales/internal/util"
)

type GptUsecase interface {
	AnalyzeDomain(ctx context.Context, domainMessage *external.DomainMessage) error
	AnalyzeDomains(ctx context.Context) error
}

type gptUsecase struct {
	baseRepo     repository.BaseRepository
	domainRepo   repository.DomainRepository
	slackAdapter adapter.SlackAdapter
	gptRepo      adapter.GptAdapter
}

func NewGptUsecase(
	baseRepo repository.BaseRepository,
	domainRepo repository.DomainRepository,
	slackAdapter adapter.SlackAdapter,
	gptRepo adapter.GptAdapter,
) GptUsecase {
	return &gptUsecase{
		baseRepo:     baseRepo,
		domainRepo:   domainRepo,
		slackAdapter: slackAdapter,
		gptRepo:      gptRepo,
	}
}

func (u *gptUsecase) AnalyzeDomain(ctx context.Context, domainMessage *external.DomainMessage) error {
	slog.Info("analyzing domain", "domainMessage", domainMessage)

	return u.baseRepo.WithTransaction(ctx, func(ctx context.Context) error {
		domain, err := u.domainRepo.GetForUpdate(ctx, repository.DomainFilter{ID: &domainMessage.DomainId})
		if err != nil {
			if errors.Is(err, entity.ErrNotFound) {
				return nil
			}
			return err
		}
		if domain.Status != model.StatusCrawlCompInfo {
			return nil
		}
		if err := u.gptRepo.Analyze(ctx, domain); err != nil {
			return err
		}
		domain.MobilePhone, domain.LandlinePhone = entity.SplitPhone(domain.Phone)
		domain.Status = model.StatusDone
		if err := u.domainRepo.Save(ctx, domain); err != nil {
			return err
		}
		slog.Info("analyzed", "domain", domain)
		return nil
	})
}

func (u *gptUsecase) AnalyzeDomains(ctx context.Context) error {
	_ = u.slackAdapter.Send(ctx, "analyze 開始")

	domains, err := u.domainRepo.FindAll(ctx, repository.DomainFilter{
		Status: util.Pointer(model.StatusCrawlCompInfo),
	})
	if err != nil {
		return err
	}

	semaphore := make(chan struct{}, 20)
	var wg sync.WaitGroup

	for _, d := range domains {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(d *model.Domain) {
			defer wg.Done()
			defer func() { <-semaphore }()

			if err := u.gptRepo.Analyze(ctx, d); err != nil {
				slog.Error("gpt repo analyze error", "error", err)
				return
			}
			d.MobilePhone, d.LandlinePhone = entity.SplitPhone(d.Phone)
			d.Status = model.StatusDone
			if err := u.domainRepo.Save(ctx, d); err != nil {
				slog.Error("gpt repo save error", "error", err)
				return
			}
		}(d)
	}

	wg.Wait()
	_ = u.slackAdapter.Send(ctx, "analyze 終了")
	return nil
}
