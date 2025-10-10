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
	gptRepo := repository.NewGptRepository()
	slackAdapter := adapter.NewSlackAdapter()
	fetchUsecase := usecase.NewFetchUsecase(viewDnsAdapter, slackAdapter, domainRepo)
	pageUsecase := usecase.NewPageUsecase(baseRepo, domainRepo)
	gptUsecase := usecase.NewGptUsecase(slackAdapter, domainRepo, gptRepo)
	return handler.NewApiHandler(fetchUsecase, pageUsecase, gptUsecase)
}

func GetGptUsecase(db *gorm.DB) usecase.GptUsecase {
	gptRepo := repository.NewGptRepository()
	domainRepo := repository.NewDomainRepository(db)
	slackAdapter := adapter.NewSlackAdapter()
	return usecase.NewGptUsecase(slackAdapter, domainRepo, gptRepo)
}
