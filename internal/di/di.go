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
	viewDnsAdapter := adapter.NewViewDNSAdapter(config.Env.ViewDnsApiUrl)
	baseRepo := repository.NewBaseRepository(db)
	targetRepo := repository.NewTargetRepository(db)
	logRepo := repository.NewLogRepository(db)
	taskRepo := repository.NewTaskRepository(db)
	gptRepo := adapter.NewGptAdapter()
	slackAdapter := adapter.NewSlackAdapter()
	redisQueue := infrastructure.NewRedisQueue()
	taskQueueAdapter := adapter.NewTaskQueueAdapter(redisQueue)
	pubSubAdapter := adapter.NewPubSubAdapter(pubSubClient)

	fetchUsecase := usecase.NewFetchUsecase(viewDnsAdapter, slackAdapter, pubSubAdapter, domainRepo, targetRepo)
	domainUsecase := usecase.NewDomainUsecase(baseRepo, domainRepo)
	targetUsecase := usecase.NewTargetUsecase(baseRepo, targetRepo)
	logUsecase := usecase.NewLogUsecase(baseRepo, logRepo)
	gptUsecase := usecase.NewGptUsecase(baseRepo, domainRepo, slackAdapter, gptRepo)
	taskUsecase := usecase.NewTaskUsecase(baseRepo, taskRepo, taskQueueAdapter)
	sshAdapter := adapter.NewSSHAdapter()
	sheetAdapter := adapter.NewSheetAdapter(sheetClient, driveClient)
	deployUsecase := usecase.NewDeployUsecase(sshAdapter, logRepo, sheetAdapter)
	sheetUsecase := usecase.NewSheetUsecase(baseRepo, domainRepo, sheetAdapter, sshAdapter)

	return handler.NewApiHandler(
		fetchUsecase,
		domainUsecase,
		targetUsecase,
		logUsecase,
		gptUsecase,
		taskUsecase,
		deployUsecase,
		sheetUsecase,
	)
}
