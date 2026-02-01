package mocks

import (
	"context"
	"go-shop-backend/internal/dto"

	"github.com/stretchr/testify/mock"
)

type AuthServiceMock struct {
	mock.Mock
}

func (a *AuthServiceMock) Login(ctx context.Context, req dto.UserLoginRequest) (*dto.UserTokenResponse, error) {
	args := a.Called(ctx, req)
	return args.Get(0).(*dto.UserTokenResponse), args.Error(1)
}

func (a *AuthServiceMock) Register(ctx context.Context, req dto.UserRegisterRequest) (*dto.UserTokenResponse, error) {
	args := a.Called(ctx, req)
	return args.Get(0).(*dto.UserTokenResponse), args.Error(1)
}
