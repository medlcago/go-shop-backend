package mocks

import (
	"context"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

var _ repository.ProductRepository = (*ProductRepositoryMock)(nil)

type ProductRepositoryMock struct {
	mock.Mock
}

func (p *ProductRepositoryMock) GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	args := p.Called(ctx, id)
	return args.Get(0).(*models.Product), args.Error(1)
}

func (p *ProductRepositoryMock) ListProducts(ctx context.Context, req dto.ListProductRequest) ([]*models.Product, int64, error) {
	args := p.Called(ctx, req)
	return args.Get(0).([]*models.Product), int64(args.Int(1)), args.Error(2)
}

func (p *ProductRepositoryMock) ListProductsByCategory(ctx context.Context, req dto.ListProductRequest) ([]*models.Product, int64, error) {
	args := p.Called(ctx, req)
	return args.Get(0).([]*models.Product), int64(args.Int(1)), args.Error(2)
}

func (p *ProductRepositoryMock) CreateProduct(ctx context.Context, product *models.Product) error {
	args := p.Called(ctx, product)
	return args.Error(0)
}
