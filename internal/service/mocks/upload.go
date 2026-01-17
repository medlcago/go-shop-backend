package mocks

import (
	"context"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/service"

	"github.com/stretchr/testify/mock"
)

var _ service.UploadService = (*UploadServiceMock)(nil)

type UploadServiceMock struct {
	mock.Mock
}

func (u *UploadServiceMock) SignURL(ctx context.Context, req dto.SignURLRequest) (*dto.SignURLResponse, error) {
	args := u.Called(ctx, req)
	return args.Get(0).(*dto.SignURLResponse), args.Error(1)
}
