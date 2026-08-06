package models

import (
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PasskeyCredential struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`

	UserID uuid.UUID `gorm:"type:uuid;not null;index:idx_passkey_credentials_user_id"`

	CredentialID []byte              `gorm:"not null;uniqueIndex:idx_passkey_credentials_credential_id_unique,where:deleted_at IS NULL"`
	Credential   webauthn.Credential `gorm:"not null;serializer:json"`

	Name *string `gorm:"type:varchar(255)"`

	CreatedAt  time.Time      `gorm:"type:timestamptz;default:now();not null"`
	UpdatedAt  time.Time      `gorm:"type:timestamptz;default:now();not null"`
	DeletedAt  gorm.DeletedAt `gorm:"type:timestamptz;index:idx_passkey_credentials_deleted_at"`
	LastUsedAt *time.Time     `gorm:"type:timestamptz"`
}
