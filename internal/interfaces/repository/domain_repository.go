package repository

import (
	"context"
	"errors"

	"github.com/zuxt268/sales/internal/domain"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DomainRepository interface {
	Exists(ctx context.Context, f DomainFilter) (bool, error)
	Get(ctx context.Context, f DomainFilter) (domain.Domain, error)
	GetForUpdate(ctx context.Context, f DomainFilter) (domain.Domain, error)
	FindAll(ctx context.Context, f DomainFilter) ([]domain.Domain, error)
	Save(ctx context.Context, domain *domain.Domain) error
	BulkInsert(ctx context.Context, domains []*domain.Domain) error
	Delete(ctx context.Context, f DomainFilter) error
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
	var domains []domain.Domain
	err := f.Apply(r.getDb(ctx).WithContext(ctx)).Find(&domains).Error
	if err != nil {
		return false, domain.WrapDatabase("failed to get domain", err)
	}
	return len(domains) > 0, nil
}

func (r *domainRepository) Get(ctx context.Context, f DomainFilter) (domain.Domain, error) {
	d := domain.Domain{}
	err := f.Apply(r.getDb(ctx).WithContext(ctx)).First(&d).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return d, domain.WrapNotFound("domain")
		}
		return d, domain.WrapDatabase("failed to get domain", err)
	}
	return d, nil
}

func (r *domainRepository) GetForUpdate(ctx context.Context, f DomainFilter) (domain.Domain, error) {
	d := domain.Domain{}
	err := f.Apply(r.getDb(ctx).WithContext(ctx)).Clauses(clause.Locking{Strength: "UPDATE"}).First(&d).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return d, domain.WrapNotFound("domain")
		}
		return d, domain.WrapDatabase("failed to get domain for update", err)
	}
	return d, nil
}

func (r *domainRepository) FindAll(ctx context.Context, f DomainFilter) ([]domain.Domain, error) {
	var ds []domain.Domain
	err := f.Apply(r.getDb(ctx).WithContext(ctx)).Find(&ds).Error
	if err != nil {
		return nil, domain.WrapDatabase("failed to find domains", err)
	}
	return ds, nil
}

func (r *domainRepository) Save(ctx context.Context, d *domain.Domain) error {
	err := r.getDb(ctx).Save(d).Error
	if err != nil {
		return domain.WrapDatabase("failed to save domain", err)
	}
	return nil
}

func (r *domainRepository) BulkInsert(ctx context.Context, domains []*domain.Domain) error {
	err := r.getDb(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		DoNothing: true,
	}).WithContext(ctx).CreateInBatches(domains, 100).Error
	if err != nil {
		return domain.WrapDatabase("failed to bulk insert domains", err)
	}
	return nil
}

func (r *domainRepository) Delete(ctx context.Context, f DomainFilter) error {
	err := f.Apply(r.db.WithContext(ctx)).Delete(&domain.Domain{}).Error
	if err != nil {
		return domain.WrapDatabase("failed to delete domain", err)
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
	CanView     *bool
	IsJapan     *bool
	IsSend      *bool
	OwnerID     *string
	Industry    *string
	IsSSL       *bool
	Status      *domain.Status
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
