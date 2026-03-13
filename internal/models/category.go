package models

import (
	"github.com/google/uuid"
)

type Category struct {
	ID       uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name     string     `gorm:"type:varchar(255);not null"`
	Slug     string     `gorm:"type:varchar(255);not null"`
	ParentID *uuid.UUID `gorm:"type:uuid;index:idx_categories_parent_id"`
	Parent   *Category  `gorm:"foreignKey:ParentID;constraint:OnDelete:CASCADE"`

	Products []Product `gorm:"many2many:product_categories;constraint:OnDelete:CASCADE"`
}
