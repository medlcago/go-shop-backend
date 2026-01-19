package models

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name        string    `gorm:"type:varchar(255);not null"`
	Description *string   `gorm:"type:text"`
	Price       float64   `gorm:"type:numeric(10,2);not null"`
	Slug        string    `gorm:"type:varchar(255);not null"`
	Stock       int       `gorm:"default:0;check:stock >= 0"`
	IsActive    bool      `gorm:"default:true;not null"`
	CreatedAt   time.Time `gorm:"type:timestamptz;default:now();not null"`
	UpdatedAt   time.Time `gorm:"type:timestamptz;default:now();not null"`

	Categories []Category `gorm:"many2many:product_categories;constraint:OnDelete:CASCADE"`
	Images     []Upload   `gorm:"polymorphic:Entity"`
}
