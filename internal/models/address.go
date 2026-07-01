package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Address struct {
	ID     uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID uuid.UUID `gorm:"type:uuid;not null;index:idx_addresses_user_id"`

	Name    string `gorm:"size:100;not null"`
	Street  string `gorm:"size:255;not null"`
	House   string `gorm:"size:50;not null"`
	City    string `gorm:"size:255;not null"`
	Country string `gorm:"size:100;not null"`

	Floor     *string `gorm:"size:50;default:null"`
	Entrance  *string `gorm:"size:50;default:null"`
	Apartment *string `gorm:"size:50;default:null"`
	Comment   *string `gorm:"type:text;default:null"`

	IsDefault bool `gorm:"not null;default:false;index:idx_addresses_is_default"`

	CreatedAt time.Time      `gorm:"type:timestamptz;default:now();not null"`
	UpdatedAt time.Time      `gorm:"type:timestamptz;default:now();not null"`
	DeletedAt gorm.DeletedAt `gorm:"type:timestamptz;index:idx_addresses_deleted_at"`
}
