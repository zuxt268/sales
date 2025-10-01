package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

type BaseRepository interface {
	WithTransaction(ctx context.Context, f func(ctx context.Context) error) (err error)
}

type baseRepository struct {
	db *gorm.DB
}

func NewBaseRepository(db *gorm.DB) BaseRepository {
	return &baseRepository{db: db}
}

func (b *baseRepository) WithTransaction(ctx context.Context, f func(ctx context.Context) error) (err error) {
	if _, ok := ctx.Value(TxKey{}).(*gorm.DB); ok {
		return fmt.Errorf("transaction already exists")
	}
	tx := b.db.Begin()

	defer func() {
		if rec := recover(); rec != nil {
			tx.Rollback()
			err = fmt.Errorf("panic recovered in transaction: %v", rec)
		} else if err != nil {
			tx.Rollback()
		} else {
			if commitErr := tx.Commit(); commitErr != nil {
				tx.Rollback()
				err = fmt.Errorf("failed to commit transaction: %s", commitErr.Error)
			}
		}
	}()

	ctxWithTx := context.WithValue(ctx, TxKey{}, tx)
	return f(ctxWithTx)
}

type TxKey struct{}
