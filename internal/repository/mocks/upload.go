package mocks

import (
	"context"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"

	"github.com/stretchr/testify/mock"
)

var _ repository.UploadRepository = (*UploadRepositoryMock)(nil)

type UploadRepositoryMock struct {
	mock.Mock
}

func (u *UploadRepositoryMock) Create(ctx context.Context, req *models.Upload) error {
	args := u.Called(ctx, req)
	return args.Error(0)
}

func (u *UploadRepositoryMock) Exists(ctx context.Context, objectKey string) (bool, error) {
	args := u.Called(ctx, objectKey)
	return args.Bool(0), args.Error(1)
}
