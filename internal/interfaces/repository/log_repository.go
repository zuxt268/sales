package repository

import (
	"context"

	"github.com/zuxt268/sales/internal/model"

	"gorm.io/gorm"
)

type LogRepository interface {
	FindAll(ctx context.Context, filter LogFilter) ([]*model.Log, error)
	Create(ctx context.Context, log *model.Log) error
	Count(ctx context.Context, filter LogFilter) (int64, error)
}

type logRepository struct {
	db *gorm.DB
}

func NewLogRepository(db *gorm.DB) LogRepository {
	return &logRepository{
		db: db,
	}
}

func (r *logRepository) FindAll(ctx context.Context, filter LogFilter) ([]*model.Log, error) {
	var logs []*model.Log
	err := filter.Apply(r.getDb(ctx).WithContext(ctx)).Find(&logs).Error
	if err != nil {
		return nil, model.WrapDatabase("failed to find logs", err)
	}
	return logs, nil
}

func (r *logRepository) Create(ctx context.Context, log *model.Log) error {
	err := r.getDb(ctx).Create(log).Error
	if err != nil {
		return model.WrapDatabase("failed to create log", err)
	}
	return nil
}

func (r *logRepository) Count(ctx context.Context, filter LogFilter) (int64, error) {
	var count int64
	filter.Limit = nil
	filter.Offset = nil
	if err := filter.Apply(r.getDb(ctx).WithContext(ctx)).Model(model.Log{}).Count(&count).Error; err != nil {
		return 0, model.WrapDatabase("failed to count logs", err)
	}
	return count, nil
}

func (r *logRepository) getDb(ctx context.Context) *gorm.DB {
	if v, ok := ctx.Value(TxKey{}).(*gorm.DB); ok {
		return v.WithContext(ctx)
	}
	return r.db.WithContext(ctx)
}

type LogFilter struct {
	Category *string
	Name     *string
	Limit    *int
	Offset   *int
}

func (f *LogFilter) Apply(db *gorm.DB) *gorm.DB {
	if f.Category != nil {
		db = db.Where("category = ?", f.Category)
	}
	if f.Name != nil {
		db = db.Where("name = ?", *f.Name)
	}
	if f.Limit != nil {
		db = db.Limit(*f.Limit)
		if f.Offset != nil {
			db = db.Offset(*f.Offset)
		}
	}
	return db.Order("created_at DESC")
}
