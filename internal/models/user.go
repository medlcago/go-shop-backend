package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRole string

const (
	UserRoleAdmin    UserRole = "admin"
	UserRoleSeller   UserRole = "seller"
	UserRoleCustomer UserRole = "customer"
)

type User struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email        string         `gorm:"type:citext;not null;uniqueIndex:idx_users_email_unique"`
	PasswordHash string         `gorm:"type:varchar(255);not null"`
	FullName     *string        `gorm:"type:varchar(255)"`
	Phone        *string        `gorm:"type:varchar(50)"`
	CreatedAt    time.Time      `gorm:"type:timestamptz;default:now();not null"`
	UpdatedAt    time.Time      `gorm:"type:timestamptz;default:now();not null"`
	DeletedAt    gorm.DeletedAt `gorm:"type:timestamptz;index:idx_users_deleted_at"`
	Role         UserRole       `gorm:"type:varchar(50);default:'customer';not null"`
}
