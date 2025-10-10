package usecase

import (
	"context"

	"github.com/zuxt268/sales/internal/domain"
	"github.com/zuxt268/sales/internal/interfaces/adapter"
	"github.com/zuxt268/sales/internal/interfaces/repository"
	"github.com/zuxt268/sales/internal/util"
)

type GptUsecase interface {
	AnalyzeDomains(ctx context.Context) error
}

type gptUsecase struct {
	slackAdapter adapter.SlackAdapter
	domainRepo   repository.DomainRepository
	gptRepo      repository.GptRepository
}

func NewGptUsecase(
	slackAdapter adapter.SlackAdapter,
	domainRepo repository.DomainRepository,
	gptRepo repository.GptRepository,
) GptUsecase {
	return &gptUsecase{
		slackAdapter: slackAdapter,
		domainRepo:   domainRepo,
		gptRepo:      gptRepo,
	}
}

func (u *gptUsecase) AnalyzeDomains(ctx context.Context) error {
	_ = u.slackAdapter.Send(ctx, "analyze 開始")
	domains, err := u.domainRepo.FindAll(ctx, repository.DomainFilter{
		Status: util.Pointer(domain.StatusCrawlCompInfo),
	})
	if err != nil {
		return err
	}
	for _, d := range domains {
		if err := u.gptRepo.Analyze(ctx, &d); err != nil {
			return err
		}
		if err := u.domainRepo.Save(ctx, &d); err != nil {
			return err
		}
	}
	_ = u.slackAdapter.Send(ctx, "analyze 終了")
	return nil
}
