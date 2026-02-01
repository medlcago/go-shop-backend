package mocks

import (
	"context"
	"go-shop-backend/internal/dto"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type UserServiceMock struct {
	mock.Mock
}

func (u *UserServiceMock) GetUserByID(ctx context.Context, userID uuid.UUID) (*dto.UserResponse, error) {
	args := u.Called(ctx, userID)
	return args.Get(0).(*dto.UserResponse), args.Error(1)
}
