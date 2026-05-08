package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Category struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ParentID    *uuid.UUID `gorm:"type:uuid;index:idx_categories_parent_id;check:parent_id IS NULL OR id <> parent_id"`
	Name        string     `gorm:"type:varchar(255);not null"`
	Slug        string     `gorm:"type:varchar(255);not null;uniqueIndex:idx_categories_slug_unique,where:deleted_at IS NULL"`
	Description string     `gorm:"type:text;not null"`
	IsActive    bool       `gorm:"type:boolean;not null;default:true"`
	SortOrder   int        `gorm:"type:integer;not null;default:0"`
	HasChildren bool       `gorm:"column:has_children;->"`

	CreatedAt time.Time      `gorm:"type:timestamptz;default:now();not null"`
	UpdatedAt time.Time      `gorm:"type:timestamptz;default:now();not null"`
	DeletedAt gorm.DeletedAt `gorm:"type:timestamptz;index:idx_categories_deleted_at"`

	Parent *Category `gorm:"foreignKey:ParentID;constraint:OnDelete:SET NULL"`

	Products []Product `gorm:"many2many:product_categories;constraint:OnDelete:CASCADE"`
}

func (c *Category) IsRoot() bool {
	return c.ParentID == nil
}
