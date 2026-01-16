package mocks

import (
	"context"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

var _ repository.UserRepository = (*UserRepositoryMock)(nil)

type UserRepositoryMock struct {
	mock.Mock
}

func (u *UserRepositoryMock) CreateUser(ctx context.Context, user *models.User) error {
	args := u.Called(ctx, user)
	return args.Error(0)
}

func (u *UserRepositoryMock) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := u.Called(ctx, id)
	return args.Get(0).(*models.User), args.Error(1)
}

func (u *UserRepositoryMock) GetByIDUnscoped(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := u.Called(ctx, id)
	return args.Get(0).(*models.User), args.Error(1)
}

func (u *UserRepositoryMock) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := u.Called(ctx, email)
	return args.Get(0).(*models.User), args.Error(1)
}

func (u *UserRepositoryMock) GetByEmailUnscoped(ctx context.Context, email string) (*models.User, error) {
	args := u.Called(ctx, email)
	return args.Get(0).(*models.User), args.Error(1)
}
