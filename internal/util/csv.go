package util

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"

	"github.com/zuxt268/sales/internal/model"
)

// ConvertToCSVReader converts [][]interface{} to CSV format as io.Reader
func ConvertToCSVReader(data [][]interface{}) (io.Reader, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Convert [][]interface{} to [][]string
	for _, row := range data {
		stringRow := make([]string, len(row))
		for i, cell := range row {
			if cell == nil {
				stringRow[i] = ""
			} else {
				stringRow[i] = fmt.Sprintf("%v", cell)
			}
		}
		if err := writer.Write(stringRow); err != nil {
			return nil, fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("CSV writer error: %w", err)
	}

	return &buf, nil
}

// ConvertToCsv converts []*model.Domain to CSV format as io.Reader
func ConvertToCsv(domains []*model.Domain) (io.Reader, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header
	header := []string{
		"ドメイン",
		"サイトタイトル",
		"ownerId",
		"携帯電話",
		"固定電話",
		"業種",
		"代表者",
		"企業名",
		"都道府県",
	}
	if err := writer.Write(header); err != nil {
		return nil, fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data rows
	for _, d := range domains {
		row := []string{
			d.Name,
			d.Title,
			d.OwnerID,
			d.MobilePhone,
			d.LandlinePhone,
			d.Industry,
			d.President,
			d.Company,
			d.Prefecture,
		}
		if err := writer.Write(row); err != nil {
			return nil, fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("CSV writer error: %w", err)
	}

	return &buf, nil
}
