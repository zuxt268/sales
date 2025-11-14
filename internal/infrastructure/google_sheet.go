package infrastructure

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type googleSheetsClient struct {
	ctx     context.Context
	service *sheets.Service
}

type GoogleSheetsClient interface {
	WriteToSheetOrCreate(spreadsheetID, sheetTitle, cellRange string, values [][]interface{}, valueInputOption string) (bool, error)
	ClearRange(spreadsheetID, clearRange string) error
	ReadRange(spreadsheetID, readRange string) ([][]interface{}, error)
	GetSpreadsheet(spreadsheetID string) (*sheets.Spreadsheet, error)
}

// NewGoogleSheetsClient サービスアカウントを使用してGoogle Sheets APIクライアントを初期化
func NewGoogleSheetsClient(credPath string) GoogleSheetsClient {
	ctx := context.Background()

	// サービスアカウントキーファイルを読み込み
	b, err := os.ReadFile(credPath)
	if err != nil {
		log.Fatalf("unable to read service account file: %s", err.Error())
	}

	// サービスアカウントからクライアントを作成
	creds, err := google.CredentialsFromJSON(ctx, b, sheets.SpreadsheetsScope)
	if err != nil {
		log.Fatalf("unable to parse service account credentials: %s", err.Error())
	}

	// Sheets APIサービスを作成
	service, err := sheets.NewService(ctx, option.WithCredentials(creds))
	if err != nil {
		log.Fatalf("unable to create sheets service: %s", err.Error())
	}

	slog.Info("Google Sheets client initialized with service account",
		"credentials_path", credPath,
	)

	return &googleSheetsClient{
		service: service,
		ctx:     ctx,
	}
}

// GetSpreadsheet スプレッドシート情報を取得
func (c *googleSheetsClient) GetSpreadsheet(spreadsheetID string) (*sheets.Spreadsheet, error) {
	spreadsheet, err := c.service.Spreadsheets.Get(spreadsheetID).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to get spreadsheet: %w", err)
	}

	slog.Info("Got spreadsheet",
		"spreadsheet_id", spreadsheetID,
		"title", spreadsheet.Properties.Title,
		"sheets_count", len(spreadsheet.Sheets),
	)

	return spreadsheet, nil
}

// ReadRange 指定範囲のデータを読み取り
// range例: "Sheet1!A1:D10" or "Sheet1" (シート全体)
func (c *googleSheetsClient) ReadRange(spreadsheetID, readRange string) ([][]interface{}, error) {
	resp, err := c.service.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to read range: %w", err)
	}

	slog.Info("Read range from spreadsheet",
		"spreadsheet_id", spreadsheetID,
		"range", readRange,
		"rows", len(resp.Values),
	)

	return resp.Values, nil
}

// ReadMultipleRanges 複数範囲のデータを一度に読み取り
func (c *googleSheetsClient) ReadMultipleRanges(spreadsheetID string, ranges []string) (map[string][][]interface{}, error) {
	resp, err := c.service.Spreadsheets.Values.BatchGet(spreadsheetID).Ranges(ranges...).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to batch read ranges: %w", err)
	}

	result := make(map[string][][]interface{})
	for _, vr := range resp.ValueRanges {
		result[vr.Range] = vr.Values
	}

	slog.Info("Batch read ranges from spreadsheet",
		"spreadsheet_id", spreadsheetID,
		"ranges_count", len(ranges),
	)

	return result, nil
}

// WriteRange 指定範囲にデータを書き込み
// valueInputOption: "RAW" (そのまま) or "USER_ENTERED" (ユーザー入力と同じ処理)
func (c *googleSheetsClient) WriteRange(spreadsheetID, writeRange string, values [][]interface{}, valueInputOption string) error {
	if valueInputOption == "" {
		valueInputOption = "USER_ENTERED"
	}

	valueRange := &sheets.ValueRange{
		Values: values,
	}

	_, err := c.service.Spreadsheets.Values.Update(spreadsheetID, writeRange, valueRange).
		ValueInputOption(valueInputOption).
		Do()

	if err != nil {
		return fmt.Errorf("unable to write range: %w", err)
	}

	slog.Info("Wrote range to spreadsheet",
		"spreadsheet_id", spreadsheetID,
		"range", writeRange,
		"rows", len(values),
	)

	return nil
}

// AppendRows 行を末尾に追加
func (c *googleSheetsClient) AppendRows(spreadsheetID, appendRange string, values [][]interface{}, valueInputOption string) error {
	if valueInputOption == "" {
		valueInputOption = "USER_ENTERED"
	}

	valueRange := &sheets.ValueRange{
		Values: values,
	}

	_, err := c.service.Spreadsheets.Values.Append(spreadsheetID, appendRange, valueRange).
		ValueInputOption(valueInputOption).
		Do()

	if err != nil {
		return fmt.Errorf("unable to append rows: %w", err)
	}

	slog.Info("Appended rows to spreadsheet",
		"spreadsheet_id", spreadsheetID,
		"range", appendRange,
		"rows", len(values),
	)

	return nil
}

// BatchUpdate 複数範囲のデータを一度に更新
func (c *googleSheetsClient) BatchUpdate(spreadsheetID string, data map[string][][]interface{}, valueInputOption string) error {
	if valueInputOption == "" {
		valueInputOption = "USER_ENTERED"
	}

	var valueRanges []*sheets.ValueRange
	for rangeStr, values := range data {
		valueRanges = append(valueRanges, &sheets.ValueRange{
			Range:  rangeStr,
			Values: values,
		})
	}

	batchUpdateRequest := &sheets.BatchUpdateValuesRequest{
		ValueInputOption: valueInputOption,
		Data:             valueRanges,
	}

	_, err := c.service.Spreadsheets.Values.BatchUpdate(spreadsheetID, batchUpdateRequest).Do()
	if err != nil {
		return fmt.Errorf("unable to batch update: %w", err)
	}

	slog.Info("Batch updated spreadsheet",
		"spreadsheet_id", spreadsheetID,
		"ranges_count", len(data),
	)

	return nil
}

// ClearRange 指定範囲のデータをクリア
func (c *googleSheetsClient) ClearRange(spreadsheetID, clearRange string) error {
	_, err := c.service.Spreadsheets.Values.Clear(spreadsheetID, clearRange, &sheets.ClearValuesRequest{}).Do()
	if err != nil {
		return fmt.Errorf("unable to clear range: %w", err)
	}
	slog.Info("Cleared range in spreadsheet",
		"spreadsheet_id", spreadsheetID,
		"range", clearRange,
	)
	return nil
}

// AddSheet 新しいシートを追加
func (c *googleSheetsClient) AddSheet(spreadsheetID, sheetTitle string) (*sheets.SheetProperties, error) {
	requests := []*sheets.Request{
		{
			AddSheet: &sheets.AddSheetRequest{
				Properties: &sheets.SheetProperties{
					Title: sheetTitle,
				},
			},
		},
	}

	batchUpdateRequest := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: requests,
	}

	resp, err := c.service.Spreadsheets.BatchUpdate(spreadsheetID, batchUpdateRequest).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to add sheet: %w", err)
	}

	addedSheet := resp.Replies[0].AddSheet

	slog.Info("Added sheet to spreadsheet",
		"spreadsheet_id", spreadsheetID,
		"sheet_title", sheetTitle,
		"sheet_id", addedSheet.Properties.SheetId,
	)

	return addedSheet.Properties, nil
}

// DeleteSheet シートを削除
func (c *googleSheetsClient) DeleteSheet(spreadsheetID string, sheetID int64) error {
	requests := []*sheets.Request{
		{
			DeleteSheet: &sheets.DeleteSheetRequest{
				SheetId: sheetID,
			},
		},
	}

	batchUpdateRequest := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: requests,
	}

	_, err := c.service.Spreadsheets.BatchUpdate(spreadsheetID, batchUpdateRequest).Do()
	if err != nil {
		return fmt.Errorf("unable to delete sheet: %w", err)
	}

	slog.Info("Deleted sheet from spreadsheet",
		"spreadsheet_id", spreadsheetID,
		"sheet_id", sheetID,
	)

	return nil
}

// UpdateSheetProperties シートのプロパティを更新（名前変更など）
func (c *googleSheetsClient) UpdateSheetProperties(spreadsheetID string, sheetID int64, newTitle string) error {
	requests := []*sheets.Request{
		{
			UpdateSheetProperties: &sheets.UpdateSheetPropertiesRequest{
				Properties: &sheets.SheetProperties{
					SheetId: sheetID,
					Title:   newTitle,
				},
				Fields: "title",
			},
		},
	}

	batchUpdateRequest := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: requests,
	}

	_, err := c.service.Spreadsheets.BatchUpdate(spreadsheetID, batchUpdateRequest).Do()
	if err != nil {
		return fmt.Errorf("unable to update sheet properties: %w", err)
	}

	slog.Info("Updated sheet properties",
		"spreadsheet_id", spreadsheetID,
		"sheet_id", sheetID,
		"new_title", newTitle,
	)

	return nil
}

// CopySheet シートをコピー
func (c *googleSheetsClient) CopySheet(spreadsheetID string, sourceSheetID int64, destinationSpreadsheetID string) (*sheets.SheetProperties, error) {
	copyRequest := &sheets.CopySheetToAnotherSpreadsheetRequest{
		DestinationSpreadsheetId: destinationSpreadsheetID,
	}

	resp, err := c.service.Spreadsheets.Sheets.CopyTo(spreadsheetID, sourceSheetID, copyRequest).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to copy sheet: %w", err)
	}

	slog.Info("Copied sheet",
		"source_spreadsheet_id", spreadsheetID,
		"source_sheet_id", sourceSheetID,
		"dest_spreadsheet_id", destinationSpreadsheetID,
		"new_sheet_id", resp.SheetId,
	)

	return resp, nil
}

// InsertRows 行を挿入
func (c *googleSheetsClient) InsertRows(spreadsheetID string, sheetID int64, startIndex, endIndex int64) error {
	requests := []*sheets.Request{
		{
			InsertDimension: &sheets.InsertDimensionRequest{
				Range: &sheets.DimensionRange{
					SheetId:    sheetID,
					Dimension:  "ROWS",
					StartIndex: startIndex,
					EndIndex:   endIndex,
				},
			},
		},
	}

	batchUpdateRequest := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: requests,
	}

	_, err := c.service.Spreadsheets.BatchUpdate(spreadsheetID, batchUpdateRequest).Do()
	if err != nil {
		return fmt.Errorf("unable to insert rows: %w", err)
	}

	slog.Info("Inserted rows",
		"spreadsheet_id", spreadsheetID,
		"sheet_id", sheetID,
		"start_index", startIndex,
		"end_index", endIndex,
	)

	return nil
}

// DeleteRows 行を削除
func (c *googleSheetsClient) DeleteRows(spreadsheetID string, sheetID int64, startIndex, endIndex int64) error {
	requests := []*sheets.Request{
		{
			DeleteDimension: &sheets.DeleteDimensionRequest{
				Range: &sheets.DimensionRange{
					SheetId:    sheetID,
					Dimension:  "ROWS",
					StartIndex: startIndex,
					EndIndex:   endIndex,
				},
			},
		},
	}

	batchUpdateRequest := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: requests,
	}

	_, err := c.service.Spreadsheets.BatchUpdate(spreadsheetID, batchUpdateRequest).Do()
	if err != nil {
		return fmt.Errorf("unable to delete rows: %w", err)
	}

	slog.Info("Deleted rows",
		"spreadsheet_id", spreadsheetID,
		"sheet_id", sheetID,
		"start_index", startIndex,
		"end_index", endIndex,
	)

	return nil
}

// InsertColumns 列を挿入
func (c *googleSheetsClient) InsertColumns(spreadsheetID string, sheetID int64, startIndex, endIndex int64) error {
	requests := []*sheets.Request{
		{
			InsertDimension: &sheets.InsertDimensionRequest{
				Range: &sheets.DimensionRange{
					SheetId:    sheetID,
					Dimension:  "COLUMNS",
					StartIndex: startIndex,
					EndIndex:   endIndex,
				},
			},
		},
	}

	batchUpdateRequest := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: requests,
	}

	_, err := c.service.Spreadsheets.BatchUpdate(spreadsheetID, batchUpdateRequest).Do()
	if err != nil {
		return fmt.Errorf("unable to insert columns: %w", err)
	}

	slog.Info("Inserted columns",
		"spreadsheet_id", spreadsheetID,
		"sheet_id", sheetID,
		"start_index", startIndex,
		"end_index", endIndex,
	)

	return nil
}

// DeleteColumns 列を削除
func (c *googleSheetsClient) DeleteColumns(spreadsheetID string, sheetID int64, startIndex, endIndex int64) error {
	requests := []*sheets.Request{
		{
			DeleteDimension: &sheets.DeleteDimensionRequest{
				Range: &sheets.DimensionRange{
					SheetId:    sheetID,
					Dimension:  "COLUMNS",
					StartIndex: startIndex,
					EndIndex:   endIndex,
				},
			},
		},
	}

	batchUpdateRequest := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: requests,
	}

	_, err := c.service.Spreadsheets.BatchUpdate(spreadsheetID, batchUpdateRequest).Do()
	if err != nil {
		return fmt.Errorf("unable to delete columns: %w", err)
	}

	slog.Info("Deleted columns",
		"spreadsheet_id", spreadsheetID,
		"sheet_id", sheetID,
		"start_index", startIndex,
		"end_index", endIndex,
	)

	return nil
}

// FindSheet タイトルでシートを検索
func (c *googleSheetsClient) FindSheet(spreadsheetID, sheetTitle string) (*sheets.Sheet, error) {
	spreadsheet, err := c.GetSpreadsheet(spreadsheetID)
	if err != nil {
		return nil, err
	}

	for _, sheet := range spreadsheet.Sheets {
		if sheet.Properties.Title == sheetTitle {
			return sheet, nil
		}
	}

	return nil, fmt.Errorf("sheet not found: %s", sheetTitle)
}

// SheetExists シートが存在するかチェック
func (c *googleSheetsClient) SheetExists(spreadsheetID, sheetTitle string) (bool, error) {
	_, err := c.FindSheet(spreadsheetID, sheetTitle)
	if err != nil {
		if err.Error() == fmt.Sprintf("sheet not found: %s", sheetTitle) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetOrCreateSheet シートが存在すれば取得、なければ作成
func (c *googleSheetsClient) GetOrCreateSheet(spreadsheetID, sheetTitle string) (*sheets.Sheet, bool, error) {
	// シートを検索
	sheet, err := c.FindSheet(spreadsheetID, sheetTitle)
	if err == nil {
		// シートが存在する
		slog.Info("Sheet already exists",
			"spreadsheet_id", spreadsheetID,
			"sheet_title", sheetTitle,
		)
		return sheet, false, nil
	}

	// シートが存在しない場合は作成
	if err.Error() == fmt.Sprintf("sheet not found: %s", sheetTitle) {
		newSheet, err := c.AddSheet(spreadsheetID, sheetTitle)
		if err != nil {
			return nil, false, fmt.Errorf("unable to create sheet: %w", err)
		}

		// Sheet型に変換
		fullSheet := &sheets.Sheet{
			Properties: newSheet,
		}

		slog.Info("Created new sheet",
			"spreadsheet_id", spreadsheetID,
			"sheet_title", sheetTitle,
			"sheet_id", newSheet.SheetId,
		)
		return fullSheet, true, nil
	}

	// その他のエラー
	return nil, false, err
}

// WriteToSheetOrCreate シートがなければ作成してからデータを書き込み
func (c *googleSheetsClient) WriteToSheetOrCreate(spreadsheetID, sheetTitle, cellRange string, values [][]interface{}, valueInputOption string) (bool, error) {
	// シートを取得または作成
	_, created, err := c.GetOrCreateSheet(spreadsheetID, sheetTitle)
	if err != nil {
		return false, err
	}

	// 書き込み範囲を構築
	writeRange := fmt.Sprintf("%s!%s", sheetTitle, cellRange)

	// データを書き込み
	err = c.WriteRange(spreadsheetID, writeRange, values, valueInputOption)
	if err != nil {
		return created, fmt.Errorf("unable to write data: %w", err)
	}

	slog.Info("Wrote data to sheet",
		"spreadsheet_id", spreadsheetID,
		"sheet_title", sheetTitle,
		"range", cellRange,
		"created", created,
	)

	return created, nil
}
