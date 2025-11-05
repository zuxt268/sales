package adapter

import (
	"context"
	"fmt"

	"github.com/zuxt268/sales/internal/infrastructure"
	"github.com/zuxt268/sales/internal/interfaces/dto"
)

type SiteSheetAdapter interface {
	Output(ctx context.Context, rival string, rows []dto.Row) error
}

type siteSheetAdapter struct {
	sheetID            string
	googleSheetsClient *infrastructure.GoogleSheetsClient
}

func NewSiteSheetAdapter(
	sheetID string,
	googleSheetsClient *infrastructure.GoogleSheetsClient,
) SiteSheetAdapter {
	return &siteSheetAdapter{
		sheetID:            sheetID,
		googleSheetsClient: googleSheetsClient,
	}
}

func (s *siteSheetAdapter) Output(ctx context.Context, rival string, rows []dto.Row) error {
	// ヘッダー行を追加
	cells := make([][]interface{}, 0, len(rows)+1)
	cells = append(cells, dto.Header)

	// データ行を追加
	for _, row := range rows {
		cells = append(cells, row.Columns)
	}

	// シート名としてrivalを使用し、A1から書き込み
	_, err := s.googleSheetsClient.WriteToSheetOrCreate(
		s.sheetID,
		rival, // シート名
		"A1",  // セル範囲の開始位置
		cells,
		"USER_ENTERED", // ユーザー入力と同じ処理
	)
	if err != nil {
		return fmt.Errorf("failed to write to sheet %s: %w", rival, err)
	}

	return nil
}
