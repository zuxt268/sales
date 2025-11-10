package adapter

import (
	"fmt"

	"github.com/zuxt268/sales/internal/infrastructure"
)

type SheetAdapter interface {
	Output(sheetID string, sheetName string, rows [][]interface{}) error
	Input(sheetID string, sheetName string) ([][]interface{}, error)
}

type sheetAdapter struct {
	googleSheetsClient infrastructure.GoogleSheetsClient
}

func NewSheetAdapter(
	googleSheetsClient infrastructure.GoogleSheetsClient,
) SheetAdapter {
	return &sheetAdapter{
		googleSheetsClient: googleSheetsClient,
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
