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

// ConvertDomainsToCSVReader converts []*model.Domain to CSV format as io.Reader
func ConvertDomainsToCSVReader(domains []*model.Domain) (io.Reader, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header
	header := []string{
		"ID", "Name", "Target", "CanView", "IsJapan", "IsSend", "Title",
		"OwnerID", "Address", "Phone", "MobilePhone", "LandlinePhone",
		"Industry", "President", "Company", "Prefecture", "IsSSL",
		"PageNum", "Status", "UpdatedAt", "CreatedAt",
	}
	if err := writer.Write(header); err != nil {
		return nil, fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data rows
	for _, d := range domains {
		row := []string{
			fmt.Sprintf("%d", d.ID),
			d.Name,
			d.Target,
			fmt.Sprintf("%t", d.CanView),
			fmt.Sprintf("%t", d.IsJapan),
			fmt.Sprintf("%t", d.IsSend),
			d.Title,
			d.OwnerID,
			d.Address,
			d.Phone,
			d.MobilePhone,
			d.LandlinePhone,
			d.Industry,
			d.President,
			d.Company,
			d.Prefecture,
			fmt.Sprintf("%t", d.IsSSL),
			fmt.Sprintf("%d", d.PageNum),
			string(d.Status),
			d.UpdatedAt.Format("2006-01-02 15:04:05"),
			d.CreatedAt.Format("2006-01-02 15:04:05"),
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
