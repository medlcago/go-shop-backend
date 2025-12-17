package mocks

import (
	"context"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/service"

	"github.com/stretchr/testify/mock"
)

var _ service.CategoryService = (*CategoryServiceMock)(nil)

type CategoryServiceMock struct {
	mock.Mock
}

func (c *CategoryServiceMock) ListCategories(ctx context.Context, req dto.ListCategoryRequest) ([]*dto.ProductCategoryResponse, int64, error) {
	args := c.Called(ctx, req)
	return args.Get(0).([]*dto.ProductCategoryResponse), int64(args.Int(1)), args.Error(2)
}
