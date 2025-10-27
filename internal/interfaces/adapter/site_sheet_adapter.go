package adapter

import (
	"context"
	"fmt"

	"github.com/zuxt268/sales/internal/domain"
	"github.com/zuxt268/sales/internal/infrastructure"
)

type SiteSheetAdapter interface {
	Output(ctx context.Context, rival string, results []domain.Domain) error
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

func (s *siteSheetAdapter) Output(ctx context.Context, rival string, results []domain.Domain) error {
	// ヘッダー行を追加
	cells := make([][]interface{}, 0, len(results)+1)
	cells = append(cells, []interface{}{
		"名前",
		"会社名",
		"携帯電話",
	})

	// データ行を追加
	for _, result := range results {
		row := []interface{}{
			result.Name,
			result.Company,
			result.MobilePhone,
		}
		cells = append(cells, row)
	}

	// シート名としてrivalを使用し、A1から書き込み
	_, err := s.googleSheetsClient.WriteToSheetOrCreate(
		s.sheetID,
		rival,          // シート名
		"A1",           // セル範囲の開始位置
		cells,
		"USER_ENTERED", // ユーザー入力と同じ処理
	)
	if err != nil {
		return fmt.Errorf("failed to write to sheet %s: %w", rival, err)
	}

	return nil
}
