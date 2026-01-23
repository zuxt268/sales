package main

import (
	"context"
	"fmt"
	"math"
	"strconv"

	"github.com/zuxt268/sales/internal/config"
	"github.com/zuxt268/sales/internal/infrastructure"
	"github.com/zuxt268/sales/internal/interfaces/adapter"
	"github.com/zuxt268/sales/internal/interfaces/dto/external"
	"github.com/zuxt268/sales/internal/interfaces/repository"
	"github.com/zuxt268/sales/internal/model"
)

func main() {
	db := infrastructure.NewDatabase()
	wixRepo := repository.NewWixRepository(db)
	viewDnsAdapter := adapter.NewViewDNSAdapter(config.Env.ViewDnsApiUrl)
	//targets := []string{"185.230.63.186", "185.230.63.171", "185.230.63.107"}
	targets := []string{"162.43.119.81", "23.236.62.147", "34.149.87.45"}

	for _, target := range targets {
		total := math.MaxInt
		fetch := 0
		page := 1
		for total > fetch {
			resp, err := viewDnsAdapter.GetReverseIP(context.Background(), &external.ReverseIpRequest{
				Host:   target,
				ApiKey: config.Env.ApiKey,
				Page:   page,
			})
			if err != nil {
				panic(err)
			}
			num, _ := strconv.Atoi(resp.Response.DomainCount)
			if total != num {
				total = num
			}
			fetch += len(resp.Response.Domains)

			// データをインサートする
			wixes := make([]*model.Wix, 0, len(resp.Response.Domains))
			for _, domain := range resp.Response.Domains {
				wixes = append(wixes, &model.Wix{
					Name:    domain.Name,
					OwnerID: "",
				})
			}
			if err := wixRepo.BulkInsert(context.Background(), wixes); err != nil {
				panic(err)
			}
			fmt.Printf("Inserted %d wixes for target %s (page %d)\n", len(wixes), target, page)
			page++
		}
	}
}
