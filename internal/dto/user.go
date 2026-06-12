package dto

import (
	"time"

	"github.com/google/uuid"
)

type UserLoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
	Code     string `json:"code" validate:"omitempty,required"`
}

type UserRegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,max=60"`
}

type UserResponse struct {
	ID             uuid.UUID `json:"id"`
	Email          string    `json:"email"`
	CreatedAt      time.Time `json:"created_at"`
	Role           string    `json:"role"`
	TwoFAEnabled   bool      `json:"two_fa_enabled"`
	EmailConfirmed bool      `json:"email_confirmed"`
}

type UserTokenResponse struct {
	*TokenResponse
	User *UserResponse `json:"user,omitempty"`
}

type Setup2FAResponse struct {
	Secret string `json:"secret"`
	QRCode string `json:"qr_code"`
}

type Confirm2FARequest struct {
	Password string `json:"password" validate:"required"`
	Code     string `json:"code" validate:"required"`
}

type Disable2FARequest struct {
	Password string `json:"password" validate:"required"`
	Code     string `json:"code" validate:"required"`
}

type SendEmailConfirmationResponse struct {
	ExpiresIn int `json:"expires_in"` // seconds
}

type ConfirmEmailRequest struct {
	Code string `json:"code" validate:"required"`
}

type ConfirmEmailResponse struct {
	OK               bool   `json:"ok"`
	EmailConfirmedAt string `json:"email_confirmed_at"` // ISO 8601
}

type ChangePasswordRequest struct {
	Password    string `json:"password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=6,max=60"`
	Code        string `json:"code" validate:"omitempty,required"`
}
