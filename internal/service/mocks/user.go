package mocks

import (
	"context"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/service"

	"github.com/stretchr/testify/mock"
)

var _ service.UserService = (*UserServiceMock)(nil)

type UserServiceMock struct {
	mock.Mock
}

func (u *UserServiceMock) GetUserByID(ctx context.Context, userID string) (*dto.UserResponse, error) {
	args := u.Called(ctx, userID)
	return args.Get(0).(*dto.UserResponse), args.Error(1)
}
