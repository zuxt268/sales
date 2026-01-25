package main

import (
	"context"
	"fmt"
	"time"

	"github.com/zuxt268/sales/internal/config"
	"github.com/zuxt268/sales/internal/infrastructure"
	"github.com/zuxt268/sales/internal/interfaces/repository"
	"github.com/zuxt268/sales/internal/util"
)

func main() {
	db := infrastructure.NewDatabase()
	wixRepo := repository.NewWixRepository(db)
	ctx := context.Background()
	wixes, err := wixRepo.FindAll(ctx, repository.WixFilter{
		HasOwnerID: util.Pointer(true),
	})
	if err != nil {
		panic(err)
	}
	data := make([][]interface{}, 0, len(wixes))
	for _, wix := range wixes {
		data = append(data, []interface{}{
			wix.Name,
			wix.OwnerID,
		})
	}

	fmt.Println(len(data))

	test := [][]interface{}{
		{
			"test", "test",
		},
	}

	csvReader, err := util.ConvertToCSVReader(test)
	if err != nil {
		panic(err)
	}

	now := time.Now().Format("20060102_150405")
	fileName := fmt.Sprintf("%s.csv", now)

	credPath := config.Env.GoogleServiceAccountPath
	googleDriveClient := infrastructure.NewGoogleDriveClient(credPath)
	backupFolderName := fmt.Sprintf("競合サイト_%s", now)
	driveFolderID := config.Env.GoogleDriveBackupFolderID
	backupFolderID, err := googleDriveClient.CreateFolder(backupFolderName, driveFolderID)
	if err != nil {
		panic(err)
	}
	if err := googleDriveClient.UploadCSV(fileName, csvReader, backupFolderID); err != nil {
		panic(err)
	}
}
