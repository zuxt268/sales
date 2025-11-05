package infrastructure

import (
	"fmt"
	"testing"

	"github.com/joho/godotenv"
	"github.com/zuxt268/sales/internal/config"
)

func TestSpreadSheet(t *testing.T) {
	_ = godotenv.Load("../../../.env")

	client := NewGoogleSheetsClient()
	data, err := client.ReadRange(config.Env.SheetID, "シート1")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(len(data))
	for i, row := range data {
		fmt.Printf("Row %d: %+v\n", i, row)
	}
}
