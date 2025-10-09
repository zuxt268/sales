package main

import (
	"context"

	"github.com/zuxt268/sales/internal/di"
	"github.com/zuxt268/sales/internal/infrastructure"
)

func main() {
	// DB接続
	db := infrastructure.NewDatabase()

	// 依存性注入
	u := di.GetGptUsecase(db)

	err := u.AnalyzeDomains(context.Background())
	if err != nil {
		panic(err)
	}
}
