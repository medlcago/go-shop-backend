package models

import (
	"errors"
	"go-shop-backend/pkg/apperror"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	errNegativeQty = errors.New("qty < 0")
)

type Product struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name        string         `gorm:"type:varchar(255);not null"`
	Description *string        `gorm:"type:text"`
	Price       int64          `gorm:"not null"`
	Slug        string         `gorm:"type:varchar(255);not null"`
	Stock       int            `gorm:"default:0;check:stock >= 0"`
	Reserved    int            `gorm:"default:0;check:reserved >= 0"`
	IsActive    bool           `gorm:"not null"`
	CreatedAt   time.Time      `gorm:"type:timestamptz;default:now();not null"`
	UpdatedAt   time.Time      `gorm:"type:timestamptz;default:now();not null"`
	DeletedAt   gorm.DeletedAt `gorm:"type:timestamptz;index:idx_products_deleted_at"`

	Categories []*Category `gorm:"many2many:product_categories;constraint:OnDelete:CASCADE"`
	Images     []*Upload   `gorm:"polymorphic:Entity;polymorphicId:EntityID;polymorphicValue:products"`
}

func (p *Product) CanBeAdded(qty int) error {
	if !p.IsActive {
		return apperror.ErrProductNotActive
	}

	if p.Available() < qty {
		return apperror.ErrInsufficientStock
	}

	return nil
}

func (p *Product) Available() int {
	return p.Stock - p.Reserved
}

func (p *Product) Reserve(qty int) error {
	if qty < 0 {
		return errNegativeQty
	}

	if !p.IsActive {
		return apperror.ErrProductNotActive
	}

	if p.Available() < qty {
		return apperror.ErrInsufficientStock
	}

	p.Reserved += qty
	return nil
}

func (p *Product) Release(qty int) error {
	if qty < 0 {
		return errNegativeQty
	}

	if p.Reserved < qty {
		return apperror.ErrInconsistentStock
	}

	p.Reserved -= qty
	return nil
}

func (p *Product) Deduct(qty int) error {
	if qty < 0 {
		return errNegativeQty
	}

	if p.Stock < qty || p.Reserved < qty {
		return apperror.ErrInconsistentStock
	}

	p.Stock -= qty
	p.Reserved -= qty
	return nil
}
