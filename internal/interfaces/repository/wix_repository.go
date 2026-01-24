package repository

import (
	"context"
	"fmt"

	"github.com/zuxt268/sales/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WixRepository interface {
	UpdateByName(ctx context.Context, wix *model.Wix) error
	BulkInsert(ctx context.Context, wixes []*model.Wix) error
	FindAll(ctx context.Context, f WixFilter) ([]*model.Wix, error)
	Count(ctx context.Context, f WixFilter) (int64, error)
	DeleteByName(ctx context.Context, name string) error
}

type wixRepository struct {
	db *gorm.DB
}

func NewWixRepository(db *gorm.DB) WixRepository {
	return &wixRepository{
		db: db,
	}
}

func (r *wixRepository) UpdateByName(ctx context.Context, wix *model.Wix) error {
	err := r.getDb(ctx).Model(&model.Wix{}).Where("name = ?", wix.Name).Update("owner_id", wix.OwnerID).Error
	if err != nil {
		return fmt.Errorf("failed to update wix: %w", err)
	}
	return nil
}

func (r *wixRepository) BulkInsert(ctx context.Context, wixes []*model.Wix) error {
	err := r.getDb(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		DoNothing: true,
	}).WithContext(ctx).CreateInBatches(wixes, 100).Error
	if err != nil {
		return fmt.Errorf("failed to insert wixes: %w", err)
	}
	return nil
}

func (r *wixRepository) FindAll(ctx context.Context, f WixFilter) ([]*model.Wix, error) {
	var wixes []*model.Wix
	err := f.Apply(r.getDb(ctx).WithContext(ctx)).Find(&wixes).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get wixes: %w", err)
	}
	return wixes, nil
}

func (r *wixRepository) Count(ctx context.Context, f WixFilter) (int64, error) {
	var count int64
	f.Limit = nil
	f.Offset = nil
	err := f.Apply(r.getDb(ctx).WithContext(ctx)).Model(&model.Wix{}).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count wixes: %w", err)
	}
	return count, nil
}

func (r *wixRepository) DeleteByName(ctx context.Context, name string) error {
	err := r.getDb(ctx).Where("name = ?", name).Delete(&model.Wix{}).Error
	if err != nil {
		return fmt.Errorf("failed to delete wix: %w", err)
	}
	return nil
}

func (r *wixRepository) getDb(ctx context.Context) *gorm.DB {
	if v, ok := ctx.Value(TxKey{}).(*gorm.DB); ok {
		return v.WithContext(ctx)
	}
	return r.db.WithContext(ctx)
}

type WixFilter struct {
	Name    *string
	OwnerID *string
	Limit   *int
	Offset  *int
}

func (f *WixFilter) Apply(db *gorm.DB) *gorm.DB {
	if f.Name != nil {
		db = db.Where("name = ?", *f.Name)
	}
	if f.OwnerID != nil {
		db = db.Where("owner_id = ?", *f.OwnerID)
	}
	if f.Limit != nil {
		db = db.Limit(*f.Limit)
		if f.Offset != nil {
			db = db.Offset(*f.Offset)
		}
	}
	return db
}