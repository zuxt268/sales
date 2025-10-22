package usecase

import (
	"context"
	"fmt"
	"sync"

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

	semaphore := make(chan struct{}, 20)
	var wg sync.WaitGroup

	for _, d := range domains {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(d domain.Domain) {
			defer wg.Done()
			defer func() { <-semaphore }()

			if err := u.gptRepo.Analyze(ctx, &d); err != nil {
				fmt.Println(err)
				return
			}
			d.SetPhone()
			d.Status = domain.StatusDone
			if err := u.domainRepo.Save(ctx, &d); err != nil {
				fmt.Println(err)
				return
			}
		}(d)
	}

	wg.Wait()
	_ = u.slackAdapter.Send(ctx, "analyze 終了")
	return nil
}
