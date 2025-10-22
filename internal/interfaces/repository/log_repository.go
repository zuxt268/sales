package repository

import (
	"context"

	"github.com/zuxt268/sales/internal/domain"

	"gorm.io/gorm"
)

type LogRepository interface {
	FindAll(ctx context.Context, filter LogFilter) ([]domain.Log, error)
	Create(ctx context.Context, log *domain.Log) error
}

type logRepository struct {
	db *gorm.DB
}

func NewLogRepository(db *gorm.DB) LogRepository {
	return &logRepository{
		db: db,
	}
}

func (r *logRepository) FindAll(ctx context.Context, filter LogFilter) ([]domain.Log, error) {
	var logs []domain.Log
	err := filter.Apply(r.getDb(ctx).WithContext(ctx)).Find(&logs).Error
	if err != nil {
		return nil, domain.WrapDatabase("failed to find logs", err)
	}
	return logs, nil
}

func (r *logRepository) Create(ctx context.Context, log *domain.Log) error {
	err := r.getDb(ctx).Create(log).Error
	if err != nil {
		return domain.WrapDatabase("failed to create log", err)
	}
	return nil
}

func (r *logRepository) getDb(ctx context.Context) *gorm.DB {
	if v, ok := ctx.Value(TxKey{}).(*gorm.DB); ok {
		return v.WithContext(ctx)
	}
	return r.db.WithContext(ctx)
}

type LogFilter struct {
	Category *string
	Limit    *int
	Offset   *int
}

func (f *LogFilter) Apply(db *gorm.DB) *gorm.DB {
	if f.Category != nil {
		db = db.Where("category = ?", f.Category)
	}
	if f.Limit != nil {
		db = db.Limit(*f.Limit)
		if f.Offset != nil {
			db = db.Offset(*f.Offset)
		}
	}
	return db.Order("created_at DESC")
}
