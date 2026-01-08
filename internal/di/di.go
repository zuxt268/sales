package di

import (
	"github.com/zuxt268/sales/internal/config"
	"github.com/zuxt268/sales/internal/infrastructure"
	"github.com/zuxt268/sales/internal/interfaces/adapter"
	"github.com/zuxt268/sales/internal/interfaces/handler"
	"github.com/zuxt268/sales/internal/interfaces/repository"
	"github.com/zuxt268/sales/internal/usecase"

	"gorm.io/gorm"
)

func Initialize(
	db *gorm.DB,
	sheetClient infrastructure.GoogleSheetsClient,
	driveClient infrastructure.GoogleDriveClient,
	pubSubClient infrastructure.PubSubClient,
) handler.ApiHandler {
	domainRepo := repository.NewDomainRepository(db)
	homstaRepo := repository.NewHomstaRepository(db)
	viewDnsAdapter := adapter.NewViewDNSAdapter(config.Env.ViewDnsApiUrl)
	baseRepo := repository.NewBaseRepository(db)
	targetRepo := repository.NewTargetRepository(db)
	gptAdapter := adapter.NewGptAdapter()
	slackAdapter := adapter.NewSlackAdapter()
	pubSubAdapter := adapter.NewPubSubAdapter(pubSubClient)

	fetchUsecase := usecase.NewFetchUsecase(viewDnsAdapter, slackAdapter, pubSubAdapter, domainRepo, targetRepo)
	domainUsecase := usecase.NewDomainUsecase(baseRepo, domainRepo)
	targetUsecase := usecase.NewTargetUsecase(baseRepo, targetRepo)
	homstaUsecase := usecase.NewHomstaUsecase(baseRepo, homstaRepo, gptAdapter)
	gptUsecase := usecase.NewGptUsecase(baseRepo, domainRepo, slackAdapter, gptAdapter)
	sshAdapter := adapter.NewSSHAdapter()
	sheetAdapter := adapter.NewSheetAdapter(sheetClient, driveClient)
	deployUsecase := usecase.NewDeployUsecase(sshAdapter, sheetAdapter, homstaRepo)
	sheetUsecase := usecase.NewSheetUsecase(baseRepo, domainRepo, sheetAdapter, sshAdapter)
	growthUsecase := usecase.NewGrowthUsecase(
		baseRepo,
		domainRepo,
		targetRepo,
		pubSubAdapter,
		viewDnsAdapter,
		sheetAdapter,
		gptAdapter,
	)

	return handler.NewApiHandler(
		fetchUsecase,
		domainUsecase,
		targetUsecase,
		gptUsecase,
		deployUsecase,
		sheetUsecase,
		growthUsecase,
		homstaUsecase,
		slackAdapter,
	)
}
