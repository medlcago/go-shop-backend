package dto

import "time"

type UserLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserRegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,max=60"`
}

type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type LoginResponse struct {
	*TokenResponse
	User *UserResponse `json:"user"`
}

type RegisterResponse struct {
	*TokenResponse
	User *UserResponse `json:"user"`
}
