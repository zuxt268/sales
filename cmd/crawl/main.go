package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/zuxt268/sales/internal/infrastructure"
	"github.com/zuxt268/sales/internal/interfaces/repository"
	"github.com/zuxt268/sales/internal/model"
	"github.com/zuxt268/sales/internal/util"
)

func main() {
	db := infrastructure.NewDatabase()
	wixRepo := repository.NewWixRepository(db)

	ctx := context.Background()

	for {
		wixes, err := wixRepo.FindAll(ctx, repository.WixFilter{
			OwnerID: util.Pointer(""),
			Limit:   util.Pointer(1000),
		})
		if err != nil {
			panic(err)
		}
		if len(wixes) == 0 {
			break
		}

		var wg sync.WaitGroup
		sem := make(chan struct{}, 20)

		for _, wix := range wixes {
			wg.Add(1)
			sem <- struct{}{}
			go func(w *model.Wix) {
				defer wg.Done()
				defer func() { <-sem }()

				ownerID := page(w.Name)
				if ownerID == "" {
					_ = wixRepo.DeleteByName(ctx, w.Name)
					return
				}
				w.OwnerID = ownerID
				_ = wixRepo.UpdateByName(ctx, w)
				fmt.Printf("日本のWixサイト: %s\n", w.Name)
			}(wix)
		}
		wg.Wait()
	}
}

var (
	hiraganaKatakana = regexp.MustCompile(`[\x{3040}-\x{309F}\x{30A0}-\x{30FF}]`)
	japaneseParticle = regexp.MustCompile(`(の|に|を|が)`)
	langJa           = regexp.MustCompile(`(?i)<html[^>]*lang=["']ja`)
	ownerIdPattern   = regexp.MustCompile(`"ownerId":"([\w-]{36})"`)
)

func page(domain string) string {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get("https://" + domain)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	text := string(body)

	// 日本語チェック
	isJapanese := (hiraganaKatakana.MatchString(text) && japaneseParticle.MatchString(text)) ||
		langJa.MatchString(text)
	if !isJapanese {
		return ""
	}

	// ownerId抽出
	match := ownerIdPattern.FindStringSubmatch(text)
	if len(match) < 2 {
		return ""
	}

	return match[1]
}
