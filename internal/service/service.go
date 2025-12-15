package service

import (
	"context"
	"go-shop-backend/internal/dto"
)

type AuthService interface {
	Login(ctx context.Context, req dto.UserLoginRequest) (*dto.UserTokenResponse, error)
	Register(ctx context.Context, req dto.UserRegisterRequest) (*dto.UserTokenResponse, error)
}

type UserService interface {
	GetUserByID(ctx context.Context, userID string) (*dto.UserResponse, error)
}
