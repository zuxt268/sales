package main

import (
	"context"
	"fmt"
	"os"

	"github.com/zuxt268/sales/internal/di"
	"github.com/zuxt268/sales/internal/domain"
	"github.com/zuxt268/sales/internal/infrastructure"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Usage: token <password>")
		os.Exit(1)
	}

	target := os.Args[1]

	// DB接続
	db := infrastructure.NewDatabase()

	// 依存性注入
	u := di.GetFetchUsecase(db)

	err := u.Fetch(context.Background(), domain.PostFetchRequest{Target: target})
	if err != nil {
		panic(err)
	}
}
