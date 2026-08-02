package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserPaymentMethod struct {
	ID     uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_payment_methods_user_method_unique,where:deleted_at IS NULL;index:idx_user_payment_methods_user"`

	Title                   string  `gorm:"type:varchar(255);not null"`
	Provider                string  `gorm:"type:varchar(255);not null;uniqueIndex:idx_user_payment_methods_user_method_unique,where:deleted_at IS NULL"`
	ProviderCustomerID      *string `gorm:"type:varchar(255)"`
	ProviderPaymentMethodID string  `gorm:"type:varchar(255);not null;uniqueIndex:idx_user_payment_methods_user_method_unique,where:deleted_at IS NULL"`

	Type string `gorm:"type:varchar(100);not null"`

	Brand    *string `gorm:"type:varchar(100)"`
	Last4    *string `gorm:"type:varchar(4)"`
	ExpMonth *int16  `gorm:"type:smallint"`
	ExpYear  *int16  `gorm:"type:smallint"`

	IsDefault bool `gorm:"not null;default:false"`

	LastUsedAt time.Time      `gorm:"type:timestamptz;default:now();not null"`
	CreatedAt  time.Time      `gorm:"type:timestamptz;default:now();not null"`
	UpdatedAt  time.Time      `gorm:"type:timestamptz;default:now();not null"`
	DeletedAt  gorm.DeletedAt `gorm:"type:timestamptz;index:idx_user_payment_methods_deleted_at"`
}
