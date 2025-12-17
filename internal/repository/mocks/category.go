package mocks

import (
	"context"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"

	"github.com/stretchr/testify/mock"
)

var _ repository.CategoryRepository = (*CategoryRepositoryMock)(nil)

type CategoryRepositoryMock struct {
	mock.Mock
}

func (c *CategoryRepositoryMock) ListCategories(ctx context.Context, req dto.ListCategoryRequest) ([]*models.ProductCategory, int64, error) {
	args := c.Called(ctx, req)
	return args.Get(0).([]*models.ProductCategory), int64(args.Int(1)), args.Error(2)
}
