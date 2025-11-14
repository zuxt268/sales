package infrastructure

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

type GoogleDriveClient interface {
	CreateFolder(folderName string, parentFolderID string) (string, error)
	UploadCSV(fileName string, csvData io.Reader, folderID string) error
}

type googleDriveClient struct {
	ctx     context.Context
	service *drive.Service
}

func NewGoogleDriveClient(credPath string) GoogleDriveClient {
	ctx := context.Background()

	// サービスアカウントキーファイルを読み込み
	b, err := os.ReadFile(credPath)
	if err != nil {
		log.Fatalf("unable to read service account file: %s", err.Error())
	}

	// サービスアカウントからクライアントを作成
	creds, err := google.CredentialsFromJSON(ctx, b, drive.DriveFileScope)
	if err != nil {
		log.Fatalf("unable to parse service account credentials: %s", err.Error())
	}

	// Drive APIサービスを作成
	service, err := drive.NewService(ctx, option.WithCredentials(creds))
	if err != nil {
		log.Fatalf("unable to create drive service: %s", err.Error())
	}

	slog.Info("Google Drive client initialized with service account",
		"credentials_path", credPath,
	)

	return &googleDriveClient{
		service: service,
		ctx:     ctx,
	}
}

// CreateFolder creates a new folder in Google Drive
func (c *googleDriveClient) CreateFolder(folderName string, parentFolderID string) (string, error) {
	file := &drive.File{
		Name:     folderName,
		MimeType: "application/vnd.google-apps.folder",
	}

	if parentFolderID != "" {
		file.Parents = []string{parentFolderID}
	}

	// Use SupportsAllDrives(true) to support shared drives
	createdFile, err := c.service.Files.Create(file).SupportsAllDrives(true).Do()
	if err != nil {
		return "", fmt.Errorf("unable to create folder: %w", err)
	}

	slog.Info("Created folder in Google Drive",
		"folder_name", folderName,
		"folder_id", createdFile.Id,
		"parent_folder_id", parentFolderID,
	)

	return createdFile.Id, nil
}

// UploadCSV uploads CSV data directly from memory to Google Drive
func (c *googleDriveClient) UploadCSV(fileName string, csvData io.Reader, folderID string) error {
	file := &drive.File{
		Name:     fileName,
		MimeType: "text/csv",
	}

	if folderID != "" {
		file.Parents = []string{folderID}
	}

	// Use SupportsAllDrives(true) to support shared drives
	_, err := c.service.Files.Create(file).Media(csvData).SupportsAllDrives(true).Do()
	if err != nil {
		return fmt.Errorf("unable to upload CSV: %w", err)
	}

	slog.Info("Uploaded CSV to Google Drive",
		"file_name", fileName,
		"folder_id", folderID,
	)

	return nil
}
