package dto

import (
	"time"

	"github.com/google/uuid"
)

type UserLoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type UserRegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,max=60"`
}

type UserResponse struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	CreatedAt    time.Time `json:"created_at"`
	Role         string    `json:"role"`
	TwoFAEnabled bool      `json:"two_fa_enabled"`
}

type UserTokenResponse struct {
	*TokenResponse
	User        *UserResponse `json:"user,omitempty"`
	Requires2FA bool          `json:"requires_2fa,omitempty"`
}

type Setup2FAResponse struct {
	Secret string `json:"secret"`
	QRCode string `json:"qr_code"`
}

type Confirm2FARequest struct {
	Password string `json:"password" validate:"required"`
	Code     string `json:"code" validate:"required"`
}

type Verify2FARequest struct {
	Token string `json:"token" validate:"required"`
	Code  string `json:"code" validate:"required"`
}

type Disable2FARequest struct {
	Password string `json:"password" validate:"required"`
	Code     string `json:"code" validate:"required"`
}
