package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Product struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name        string         `gorm:"type:varchar(255);not null"`
	Description *string        `gorm:"type:text"`
	Price       int64          `gorm:"not null"`
	Slug        string         `gorm:"type:varchar(255);not null"`
	Stock       int            `gorm:"default:0;check:stock >= 0"`
	Reserved    int            `gorm:"default:0;check:reserved >= 0"`
	IsActive    bool           `gorm:"default:true;not null"`
	CreatedAt   time.Time      `gorm:"type:timestamptz;default:now();not null"`
	UpdatedAt   time.Time      `gorm:"type:timestamptz;default:now();not null"`
	DeletedAt   gorm.DeletedAt `gorm:"type:timestamptz;index:idx_products_deleted_at"`

	Categories []Category `gorm:"many2many:product_categories;constraint:OnDelete:CASCADE"`
	Images     []Upload   `gorm:"polymorphic:Entity"`
}

func (p Product) Available() int {
	return p.Stock - p.Reserved
}
