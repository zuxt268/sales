package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/zuxt268/sales/internal/entity"
	"github.com/zuxt268/sales/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type HomstaRepository interface {
	Exists(ctx context.Context, f HomstaFilter) (bool, error)
	Get(ctx context.Context, f HomstaFilter) (*model.Homsta, error)
	GetForUpdate(ctx context.Context, f HomstaFilter) (*model.Homsta, error)
	FindAll(ctx context.Context, f HomstaFilter) ([]*model.Homsta, error)
	Save(ctx context.Context, homsta *model.Homsta) error
	BulkInsert(ctx context.Context, homstas []*model.Homsta) error
	Delete(ctx context.Context, f HomstaFilter) error
	Count(ctx context.Context, f HomstaFilter) (int64, error)
}

type homstaRepository struct {
	db *gorm.DB
}

func NewHomstaRepository(db *gorm.DB) HomstaRepository {
	return &homstaRepository{
		db: db,
	}
}

func (r *homstaRepository) Exists(ctx context.Context, f HomstaFilter) (bool, error) {
	var homstas []model.Homsta
	err := f.Apply(r.getDb(ctx).WithContext(ctx)).Find(&homstas).Error
	if err != nil {
		return false, fmt.Errorf("failed to get homsta: %w", err)
	}
	return len(homstas) > 0, nil
}

func (r *homstaRepository) Get(ctx context.Context, f HomstaFilter) (*model.Homsta, error) {
	h := &model.Homsta{}
	err := f.Apply(r.getDb(ctx).WithContext(ctx)).First(h).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return h, entity.WrapNotFound("homsta")
		}
		return nil, fmt.Errorf("failed to get homsta: %w", err)
	}
	return h, nil
}

func (r *homstaRepository) GetForUpdate(ctx context.Context, f HomstaFilter) (*model.Homsta, error) {
	h := &model.Homsta{}
	err := f.Apply(r.getDb(ctx).WithContext(ctx)).Clauses(clause.Locking{Strength: "UPDATE"}).First(h).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return h, entity.WrapNotFound("homsta")
		}
		return nil, fmt.Errorf("failed to get homsta for update: %w", err)
	}
	return h, nil
}

func (r *homstaRepository) FindAll(ctx context.Context, f HomstaFilter) ([]*model.Homsta, error) {
	var hs []*model.Homsta
	err := f.Apply(r.getDb(ctx).WithContext(ctx)).Find(&hs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get homstas: %w", err)
	}
	return hs, nil
}

func (r *homstaRepository) Count(ctx context.Context, f HomstaFilter) (int64, error) {
	var count int64
	f.Limit = nil
	f.Offset = nil
	err := f.Apply(r.getDb(ctx).WithContext(ctx)).Model(&model.Homsta{}).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count homstas: %w", err)
	}
	return count, nil
}

func (r *homstaRepository) Save(ctx context.Context, h *model.Homsta) error {
	err := r.getDb(ctx).Save(h).Error
	if err != nil {
		return fmt.Errorf("failed to save homsta: %w", err)
	}
	return nil
}

func (r *homstaRepository) BulkInsert(ctx context.Context, homstas []*model.Homsta) error {
	err := r.getDb(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		DoNothing: true,
	}).WithContext(ctx).CreateInBatches(homstas, 100).Error
	if err != nil {
		return fmt.Errorf("failed to insert homstas: %w", err)
	}
	return nil
}

func (r *homstaRepository) Delete(ctx context.Context, f HomstaFilter) error {
	err := f.Apply(r.db.WithContext(ctx)).Delete(&model.Homsta{}).Error
	if err != nil {
		return fmt.Errorf("failed to delete homsta: %w", err)
	}
	return nil
}

func (r *homstaRepository) getDb(ctx context.Context) *gorm.DB {
	if v, ok := ctx.Value(TxKey{}).(*gorm.DB); ok {
		return v.WithContext(ctx)
	}
	return r.db.WithContext(ctx)
}

type HomstaFilter struct {
	Name        *string
	PartialName *string
	Path        *string
	DBName      *string
	Industry    *string
	Limit       *int
	Offset      *int
}

func (h *HomstaFilter) Apply(db *gorm.DB) *gorm.DB {
	if h.Name != nil {
		db = db.Where("name = ?", *h.Name)
	}
	if h.PartialName != nil {
		db = db.Where("name like ?", "%"+*h.PartialName+"%")
	}
	if h.Path != nil {
		db = db.Where("path = ?", *h.Path)
	}
	if h.DBName != nil {
		db = db.Where("db_name = ?", *h.DBName)
	}
	if h.Industry != nil {
		db = db.Where("industry = ?", *h.Industry)
	}
	if h.Limit != nil {
		db = db.Limit(*h.Limit)
		if h.Offset != nil {
			db = db.Offset(*h.Offset)
		}
	}
	return db
}