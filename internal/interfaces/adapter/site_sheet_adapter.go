package adapter

import (
	"fmt"
	"time"

	"github.com/zuxt268/sales/internal/config"
	"github.com/zuxt268/sales/internal/infrastructure"
	"github.com/zuxt268/sales/internal/model"
	"github.com/zuxt268/sales/internal/util"
)

type SheetAdapter interface {
	Output(sheetID string, sheetName string, rows [][]interface{}) error
	Input(sheetID string, sheetName string) ([][]interface{}, error)
	BackupToGoogleDrive(sheetID string, driveFolderID string) error
	BackupDomainsToGoogleDrive(domainsByTarget map[string][]*model.Domain, driveFolderID string) error
	ClearAllSheets(sheetID string) error
	ShareDrive(driveFolderID string) error
}

type sheetAdapter struct {
	googleSheetsClient infrastructure.GoogleSheetsClient
	googleDriveClient  infrastructure.GoogleDriveClient
}

func NewSheetAdapter(
	googleSheetsClient infrastructure.GoogleSheetsClient,
	googleDriveClient infrastructure.GoogleDriveClient,
) SheetAdapter {
	return &sheetAdapter{
		googleSheetsClient: googleSheetsClient,
		googleDriveClient:  googleDriveClient,
	}
}

func (s *sheetAdapter) Output(sheetID, sheetName string, rows [][]interface{}) error {
	err := s.googleSheetsClient.ClearRange(sheetID, fmt.Sprintf("%s!A:Z", sheetName))
	if err != nil {
		return err
	}

	// シート名としてrivalを使用し、A1から書き込み
	_, err = s.googleSheetsClient.WriteToSheetOrCreate(
		sheetID,
		sheetName, // シート名
		"A1",      // セル範囲の開始位置
		rows,
		"USER_ENTERED", // ユーザー入力と同じ処理
	)
	if err != nil {
		return fmt.Errorf("failed to write to sheet %s: %w", sheetName, err)
	}
	return nil
}

func (s *sheetAdapter) Input(sheetID string, sheetName string) ([][]interface{}, error) {
	return s.googleSheetsClient.ReadRange(sheetID, fmt.Sprintf("%s!A:Z", sheetName))
}

// BackupToGoogleDrive backs up all sheets in a spreadsheet to Google Drive as CSV files
func (s *sheetAdapter) BackupToGoogleDrive(sheetID string, driveFolderID string) error {
	spreadsheet, err := s.googleSheetsClient.GetSpreadsheet(sheetID)
	if err != nil {
		return fmt.Errorf("failed to get spreadsheet: %w", err)
	}

	folderName := fmt.Sprintf("競合サイト_%s", time.Now().Format("20060102_150405"))
	backupFolderID, err := s.googleDriveClient.CreateFolder(folderName, driveFolderID)
	if err != nil {
		return fmt.Errorf("failed to create backup folder: %w", err)
	}

	// Backup each sheet
	for _, sheet := range spreadsheet.Sheets {
		sheetTitle := sheet.Properties.Title

		// Read sheet data
		data, err := s.googleSheetsClient.ReadRange(sheetID, fmt.Sprintf("%s!A:Z", sheetTitle))
		if err != nil {
			return fmt.Errorf("failed to read sheet %s: %w", sheetTitle, err)
		}

		// Skip empty sheets
		if len(data) == 0 {
			continue
		}

		// Convert to CSV
		csvReader, err := util.ConvertToCSVReader(data)
		if err != nil {
			return fmt.Errorf("failed to convert sheet %s to CSV: %w", sheetTitle, err)
		}

		// Upload to Google Drive
		fileName := fmt.Sprintf("%s.csv", sheetTitle)
		if err := s.googleDriveClient.UploadCSV(fileName, csvReader, backupFolderID); err != nil {
			return fmt.Errorf("failed to upload CSV for sheet %s: %w", sheetTitle, err)
		}
	}

	return nil
}

// BackupDomainsToGoogleDrive backs up domains grouped by target to Google Drive as CSV files
func (s *sheetAdapter) BackupDomainsToGoogleDrive(domainsByTarget map[string][]*model.Domain, driveFolderID string) error {
	// Create backup folder with timestamp
	timestamp := time.Now().Format("20060102_150405")
	backupFolderName := fmt.Sprintf("競合サイト_%s", timestamp)
	backupFolderID, err := s.googleDriveClient.CreateFolder(backupFolderName, driveFolderID)
	if err != nil {
		return fmt.Errorf("failed to create backup folder: %w", err)
	}

	// Backup each target as a separate CSV file
	for target, domains := range domainsByTarget {
		// Skip if no domains
		if len(domains) == 0 {
			continue
		}

		// Convert to CSV
		csvReader, err := util.ConvertDomainsToCSVReader(domains)
		if err != nil {
			return fmt.Errorf("failed to convert domains for target %s to CSV: %w", target, err)
		}

		// Upload to Google Drive
		fileName := fmt.Sprintf("%s.csv", target)
		if err := s.googleDriveClient.UploadCSV(fileName, csvReader, backupFolderID); err != nil {
			return fmt.Errorf("failed to upload CSV for target %s: %w", target, err)
		}
	}

	// Share folder with specified email if configured (after all files are uploaded)
	if shareEmail := config.Env.GoogleDriveShareEmail; shareEmail != "" {
		if err := s.googleDriveClient.ShareFolder(backupFolderID, shareEmail); err != nil {
			return fmt.Errorf("failed to share backup folder: %w", err)
		}
	}

	return nil
}

// ClearAllSheets clears all data in all sheets of a spreadsheet
func (s *sheetAdapter) ClearAllSheets(sheetID string) error {
	// Get spreadsheet info
	spreadsheet, err := s.googleSheetsClient.GetSpreadsheet(sheetID)
	if err != nil {
		return fmt.Errorf("failed to get spreadsheet: %w", err)
	}

	// Clear each sheet
	for _, sheet := range spreadsheet.Sheets {
		sheetTitle := sheet.Properties.Title
		clearRange := fmt.Sprintf("%s!A:Z", sheetTitle)
		if err := s.googleSheetsClient.ClearRange(sheetID, clearRange); err != nil {
			return fmt.Errorf("failed to clear sheet %s: %w", sheetTitle, err)
		}
	}

	return nil
}

func (s *sheetAdapter) ShareDrive(driveFolderID string) error {
	return s.googleDriveClient.ShareFolder(driveFolderID, config.Env.GoogleDriveShareEmail)
}
