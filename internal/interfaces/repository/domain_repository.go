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

type DomainRepository interface {
	Exists(ctx context.Context, f DomainFilter) (bool, error)
	Get(ctx context.Context, f DomainFilter) (*model.Domain, error)
	GetForUpdate(ctx context.Context, f DomainFilter) (*model.Domain, error)
	FindAll(ctx context.Context, f DomainFilter) ([]*model.Domain, error)
	Save(ctx context.Context, domain *model.Domain) error
	BulkInsert(ctx context.Context, domains []*model.Domain) error
	Delete(ctx context.Context, f DomainFilter) error
	Count(ctx context.Context, f DomainFilter) (int64, error)
}

type domainRepository struct {
	db *gorm.DB
}

func NewDomainRepository(db *gorm.DB) DomainRepository {
	return &domainRepository{
		db: db,
	}
}

func (r *domainRepository) Exists(ctx context.Context, f DomainFilter) (bool, error) {
	var domains []model.Domain
	err := f.Apply(r.getDb(ctx).WithContext(ctx)).Find(&domains).Error
	if err != nil {
		return false, fmt.Errorf("failed to get domain: %w", err)
	}
	return len(domains) > 0, nil
}

func (r *domainRepository) Get(ctx context.Context, f DomainFilter) (*model.Domain, error) {
	d := &model.Domain{}
	err := f.Apply(r.getDb(ctx).WithContext(ctx)).First(d).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return d, entity.WrapNotFound("domain")
		}
		return nil, fmt.Errorf("failed to get domain: %w", err)
	}
	return d, nil
}

func (r *domainRepository) GetForUpdate(ctx context.Context, f DomainFilter) (*model.Domain, error) {
	d := &model.Domain{}
	err := f.Apply(r.getDb(ctx).WithContext(ctx)).Clauses(clause.Locking{Strength: "UPDATE"}).First(d).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return d, entity.WrapNotFound("domain")
		}
		return nil, fmt.Errorf("failed to get domain for update: %w", err)
	}
	return d, nil
}

func (r *domainRepository) FindAll(ctx context.Context, f DomainFilter) ([]*model.Domain, error) {
	var ds []*model.Domain
	err := f.Apply(r.getDb(ctx).WithContext(ctx)).Find(&ds).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get domains: %w", err)
	}
	return ds, nil
}

func (r *domainRepository) Count(ctx context.Context, f DomainFilter) (int64, error) {
	var count int64
	f.Limit = nil
	f.Offset = nil
	err := f.Apply(r.getDb(ctx).WithContext(ctx)).Model(&model.Domain{}).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count domains: %w", err)
	}
	return count, nil
}

func (r *domainRepository) Save(ctx context.Context, d *model.Domain) error {
	err := r.getDb(ctx).Save(d).Error
	if err != nil {
		return fmt.Errorf("failed to save domain: %w", err)
	}
	return nil
}

func (r *domainRepository) BulkInsert(ctx context.Context, domains []*model.Domain) error {
	err := r.getDb(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		DoNothing: true,
	}).WithContext(ctx).CreateInBatches(domains, 100).Error
	if err != nil {
		return fmt.Errorf("failed to insert domains: %w", err)
	}
	return nil
}

func (r *domainRepository) Delete(ctx context.Context, f DomainFilter) error {
	err := f.Apply(r.db.WithContext(ctx)).Delete(&model.Domain{}).Error
	if err != nil {
		return fmt.Errorf("failed to delete domain: %w", err)
	}
	return nil
}

func (r *domainRepository) getDb(ctx context.Context) *gorm.DB {
	if v, ok := ctx.Value(TxKey{}).(*gorm.DB); ok {
		return v.WithContext(ctx)
	}
	return r.db.WithContext(ctx)
}

type DomainFilter struct {
	ID          *int
	PartialName *string
	Name        *string
	Target      *string
	CanView     *bool
	IsJapan     *bool
	IsSend      *bool
	OwnerID     *string
	Industry    *string
	IsSSL       *bool
	Status      *model.Status
	Limit       *int
	Offset      *int
}

func (d *DomainFilter) Apply(db *gorm.DB) *gorm.DB {
	if d.ID != nil {
		db = db.Where("id = ?", *d.ID)
	}
	if d.PartialName != nil {
		db = db.Where("name like ?", "%"+*d.PartialName+"%")
	}
	if d.Name != nil {
		db = db.Where("name = ?", *d.Name)
	}
	if d.Target != nil {
		db = db.Where("target = ?", *d.Target)
	}
	if d.CanView != nil {
		db = db.Where("can_view = ?", *d.CanView)
	}
	if d.IsJapan != nil {
		db = db.Where("is_japan = ?", *d.IsJapan)
	}
	if d.OwnerID != nil {
		db = db.Where("owner_id = ?", *d.OwnerID)
	}
	if d.Industry != nil {
		db = db.Where("industry = ?", *d.Industry)
	}
	if d.IsSend != nil {
		db = db.Where("is_send = ?", *d.IsSend)
	}
	if d.IsSSL != nil {
		db = db.Where("is_ssl = ?", *d.IsSSL)
	}
	if d.Status != nil {
		db = db.Where("status = ?", *d.Status)
	}
	if d.Limit != nil {
		db = db.Limit(*d.Limit)
		if d.Offset != nil {
			db = db.Offset(*d.Offset)
		}
	}
	return db
}
