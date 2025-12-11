package service

import (
	"context"
	"go-shop-backend/internal/dto"
)

type AuthService interface {
	Login(ctx context.Context, req dto.UserLoginRequest) (*dto.LoginResponse, error)
	Register(ctx context.Context, req dto.UserRegisterRequest) (*dto.RegisterResponse, error)
}
