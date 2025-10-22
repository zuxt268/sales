package di

import (
	"github.com/zuxt268/sales/internal/config"
	"github.com/zuxt268/sales/internal/interfaces/adapter"
	"github.com/zuxt268/sales/internal/interfaces/handler"
	"github.com/zuxt268/sales/internal/interfaces/repository"
	"github.com/zuxt268/sales/internal/usecase"

	"gorm.io/gorm"
)

func Initialize(db *gorm.DB) handler.ApiHandler {
	domainRepo := repository.NewDomainRepository(db)
	viewDnsAdapter := adapter.NewViewDNSAdapter(config.Env.ViewDnsApiUrl)
	baseRepo := repository.NewBaseRepository(db)
	targetRepo := repository.NewTargetRepository(db)
	gptRepo := repository.NewGptRepository()
	slackAdapter := adapter.NewSlackAdapter()
	fetchUsecase := usecase.NewFetchUsecase(targetRepo, viewDnsAdapter, slackAdapter, domainRepo)
	pageUsecase := usecase.NewPageUsecase(baseRepo, domainRepo, targetRepo)
	gptUsecase := usecase.NewGptUsecase(slackAdapter, domainRepo, gptRepo)
	return handler.NewApiHandler(fetchUsecase, pageUsecase, gptUsecase)
}

func GetGptUsecase(db *gorm.DB) usecase.GptUsecase {
	gptRepo := repository.NewGptRepository()
	domainRepo := repository.NewDomainRepository(db)
	slackAdapter := adapter.NewSlackAdapter()
	return usecase.NewGptUsecase(slackAdapter, domainRepo, gptRepo)
}

func GetFetchUsecase(db *gorm.DB) usecase.FetchUsecase {
	targetRepo := repository.NewTargetRepository(db)
	domainRepo := repository.NewDomainRepository(db)
	viewDnsAdapter := adapter.NewViewDNSAdapter(config.Env.ViewDnsApiUrl)
	slackAdapter := adapter.NewSlackAdapter()
	return usecase.NewFetchUsecase(targetRepo, viewDnsAdapter, slackAdapter, domainRepo)
}
