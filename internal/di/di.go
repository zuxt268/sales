package di

import (
	"github.com/zuxt268/sales/internal/config"
	"github.com/zuxt268/sales/internal/interfaces/handler"
	"github.com/zuxt268/sales/internal/interfaces/repository"
	"github.com/zuxt268/sales/internal/usecase"

	"gorm.io/gorm"
)

func Initialize(db *gorm.DB) handler.ApiHandler {
	domainRepo := repository.NewDomainRepository(db)
	viewDnsRepo := repository.NewViewDNSRepository(config.Env.ViewDnsApiUrl)
	baseRepo := repository.NewBaseRepository(db)
	fetchUsecase := usecase.NewFetchUsecase(viewDnsRepo, domainRepo)
	pageUsecase := usecase.NewPageUsecase(baseRepo, domainRepo)
	return handler.NewApiHandler(fetchUsecase, pageUsecase)
}
