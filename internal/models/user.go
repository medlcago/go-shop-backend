package models

import (
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
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
	ID               uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email            string         `gorm:"type:citext;not null;uniqueIndex:idx_users_email_unique"`
	PasswordHash     string         `gorm:"type:varchar(255);not null"`
	FullName         *string        `gorm:"type:varchar(255)"`
	Phone            *string        `gorm:"type:varchar(50)"`
	CreatedAt        time.Time      `gorm:"type:timestamptz;default:now();not null"`
	UpdatedAt        time.Time      `gorm:"type:timestamptz;default:now();not null"`
	DeletedAt        gorm.DeletedAt `gorm:"type:timestamptz;index:idx_users_deleted_at"`
	EmailConfirmedAt *time.Time     `gorm:"type:timestamptz"`
	Role             UserRole       `gorm:"type:varchar(50);default:'customer';not null"`

	TwoFAEnabled     bool       `gorm:"default:false;not null"`
	TwoFASecret      *string    `gorm:"varchar(255)"`
	TwoFAConfirmedAt *time.Time `gorm:"type:timestamptz"`

	Passkeys []PasskeyCredential `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`

	Addresses []Address `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

func (u *User) EmailConfirmed() bool {
	return u.EmailConfirmedAt != nil
}

func (u *User) WebAuthnID() []byte {
	return u.ID[:]
}

func (u *User) WebAuthnName() string {
	return u.Email
}

func (u *User) WebAuthnDisplayName() string {
	if u.FullName != nil {
		return *u.FullName
	}

	return u.Email
}

func (u *User) WebAuthnCredentials() []webauthn.Credential {
	credentials := make([]webauthn.Credential, 0, len(u.Passkeys))

	for _, passkey := range u.Passkeys {
		credentials = append(credentials, passkey.Credential)
	}

	return credentials
}
