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
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	Role      string    `json:"role"`
}

type UserTokenResponse struct {
	*TokenResponse
	User *UserResponse `json:"user"`
}
