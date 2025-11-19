package main

import (
	"github.com/zuxt268/sales/internal/config"
	"github.com/zuxt268/sales/internal/infrastructure"
	"github.com/zuxt268/sales/internal/interfaces/adapter"
)

func main() {
	credPath := config.Env.GoogleServiceAccountPath
	sheetClient := infrastructure.NewGoogleSheetsClient(credPath)

	driveClient := infrastructure.NewGoogleDriveClient(credPath)
	a := adapter.NewSheetAdapter(sheetClient, driveClient)

	err := a.ShareDrive("")
	if err != nil {
		panic(err)
	}
}
