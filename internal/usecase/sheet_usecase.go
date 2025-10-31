package usecase

import (
	"context"

	"github.com/zuxt268/sales/internal/interfaces/adapter"
	"github.com/zuxt268/sales/internal/interfaces/repository"
	"github.com/zuxt268/sales/internal/model"
	"github.com/zuxt268/sales/internal/util"
)

type SheetUsecase interface {
	Output(ctx context.Context) error
}

type sheetUsecase struct {
	baseRepo         repository.BaseRepository
	domainRepo       repository.DomainRepository
	siteSheetAdapter adapter.SiteSheetAdapter
}

func NewSheetUsecase(
	baseRepo repository.BaseRepository,
	domainRepo repository.DomainRepository,
	siteSheetAdapter adapter.SiteSheetAdapter,
) SheetUsecase {
	return &sheetUsecase{
		baseRepo:         baseRepo,
		domainRepo:       domainRepo,
		siteSheetAdapter: siteSheetAdapter,
	}
}

func (s *sheetUsecase) Output(ctx context.Context) error {
	// ステータスが"done"のドメインを全て取得
	domains, err := s.domainRepo.FindAll(ctx, repository.DomainFilter{
		Status: util.Pointer(model.StatusDone),
	})
	if err != nil {
		return err
	}

	// ターゲットごとにドメインをグループ化
	results := make(map[string][]model.Domain)
	for _, d := range domains {
		results[d.Target] = append(results[d.Target], d)
	}

	// 各ターゲットごとにスプレッドシートに出力
	var errors []error
	for target, domains := range results {
		if err := s.siteSheetAdapter.Output(ctx, target, domains); err != nil {
			// エラーを収集して処理を継続（全ターゲットを処理）
			errors = append(errors, err)
		}
	}

	// エラーがあれば最初のエラーを返す
	if len(errors) > 0 {
		return errors[0]
	}

	return nil
}
