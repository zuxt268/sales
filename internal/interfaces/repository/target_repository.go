package repository

import (
	"context"
	"errors"

	"github.com/zuxt268/sales/internal/domain"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TargetRepository interface {
	Exists(ctx context.Context, f TargetFilter) (bool, error)
	Get(ctx context.Context, f TargetFilter) (domain.Target, error)
	GetForUpdate(ctx context.Context, f TargetFilter) (domain.Target, error)
	FindAll(ctx context.Context, f TargetFilter) ([]domain.Target, error)
	Save(ctx context.Context, target *domain.Target) error
	BulkInsert(ctx context.Context, targets []*domain.Target) error
	Delete(ctx context.Context, f TargetFilter) error
}

type targetRepository struct {
	db *gorm.DB
}

func NewTargetRepository(db *gorm.DB) TargetRepository {
	return &targetRepository{
		db: db,
	}
}

func (r *targetRepository) Exists(ctx context.Context, f TargetFilter) (bool, error) {
	var targets []domain.Target
	err := f.Apply(r.getDb(ctx).WithContext(ctx)).Find(&targets).Error
	if err != nil {
		return false, domain.WrapDatabase("failed to get target", err)
	}
	return len(targets) > 0, nil
}

func (r *targetRepository) Get(ctx context.Context, f TargetFilter) (domain.Target, error) {
	t := domain.Target{}
	err := f.Apply(r.getDb(ctx).WithContext(ctx)).First(&t).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return t, domain.WrapNotFound("target")
		}
		return t, domain.WrapDatabase("failed to get target", err)
	}
	return t, nil
}

func (r *targetRepository) GetForUpdate(ctx context.Context, f TargetFilter) (domain.Target, error) {
	t := domain.Target{}
	err := f.Apply(r.getDb(ctx).WithContext(ctx)).Clauses(clause.Locking{Strength: "UPDATE"}).First(&t).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return t, domain.WrapNotFound("target")
		}
		return t, domain.WrapDatabase("failed to get target for update", err)
	}
	return t, nil
}

func (r *targetRepository) FindAll(ctx context.Context, f TargetFilter) ([]domain.Target, error) {
	var ts []domain.Target
	err := f.Apply(r.getDb(ctx).WithContext(ctx)).Find(&ts).Error
	if err != nil {
		return nil, domain.WrapDatabase("failed to find targets", err)
	}
	return ts, nil
}

func (r *targetRepository) Save(ctx context.Context, t *domain.Target) error {
	err := r.getDb(ctx).Save(t).Error
	if err != nil {
		return domain.WrapDatabase("failed to save target", err)
	}
	return nil
}

func (r *targetRepository) BulkInsert(ctx context.Context, targets []*domain.Target) error {
	err := r.getDb(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "ip"}},
		DoNothing: true,
	}).WithContext(ctx).CreateInBatches(targets, 100).Error
	if err != nil {
		return domain.WrapDatabase("failed to bulk insert targets", err)
	}
	return nil
}

func (r *targetRepository) Delete(ctx context.Context, f TargetFilter) error {
	err := f.Apply(r.db.WithContext(ctx)).Delete(&domain.Target{}).Error
	if err != nil {
		return domain.WrapDatabase("failed to delete target", err)
	}
	return nil
}

func (r *targetRepository) getDb(ctx context.Context) *gorm.DB {
	if v, ok := ctx.Value(TxKey{}).(*gorm.DB); ok {
		return v.WithContext(ctx)
	}
	return r.db.WithContext(ctx)
}

type TargetFilter struct {
	ID     *int
	IP     *string
	Name   *string
	Status *domain.TargetStatus
	Limit  *int
	Offset *int
}

func (f *TargetFilter) Apply(db *gorm.DB) *gorm.DB {
	if f.ID != nil {
		db = db.Where("id = ?", *f.ID)
	}
	if f.IP != nil {
		db = db.Where("ip = ?", *f.IP)
	}
	if f.Name != nil {
		db = db.Where("name = ?", *f.Name)
	}
	if f.Status != nil {
		db = db.Where("status = ?", *f.Status)
	}
	if f.Limit != nil {
		db = db.Limit(*f.Limit)
		if f.Offset != nil {
			db = db.Offset(*f.Offset)
		}
	}
	return db
}
