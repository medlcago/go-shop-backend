package mocks

import (
	"context"
	"go-shop-backend/internal/dto"

	"github.com/stretchr/testify/mock"
)

type UploadServiceMock struct {
	mock.Mock
}

func (u *UploadServiceMock) SignURL(ctx context.Context, req dto.SignURLRequest) (*dto.SignURLResponse, error) {
	args := u.Called(ctx, req)
	return args.Get(0).(*dto.SignURLResponse), args.Error(1)
}

func (u *UploadServiceMock) Save(ctx context.Context, req dto.UploadRequest) (*dto.UploadResponse, error) {
	args := u.Called(ctx, req)
	return args.Get(0).(*dto.UploadResponse), args.Error(1)
}

func (u *UploadServiceMock) PublicURL(ctx context.Context, objectKey string) string {
	args := u.Called(ctx, objectKey)
	return args.String(0)
}
