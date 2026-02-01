package mocks

import (
	"context"
	"go-shop-backend/internal/dto"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type ProductServiceMock struct {
	mock.Mock
}

func (p *ProductServiceMock) GetProductByID(ctx context.Context, productID uuid.UUID) (*dto.ProductResponse, error) {
	args := p.Called(ctx, productID)
	return args.Get(0).(*dto.ProductResponse), args.Error(1)
}

func (p *ProductServiceMock) ListProducts(ctx context.Context, req dto.ListProductRequest) ([]*dto.ProductResponse, int64, error) {
	args := p.Called(ctx, req)
	return args.Get(0).([]*dto.ProductResponse), int64(args.Int(1)), args.Error(2)
}

func (p *ProductServiceMock) CreateProduct(ctx context.Context, req dto.ProductCreateRequest) (*dto.ProductResponse, error) {
	args := p.Called(ctx, req)
	return args.Get(0).(*dto.ProductResponse), args.Error(1)
}

func (p *ProductServiceMock) UpdateProduct(ctx context.Context, productID uuid.UUID, req dto.ProductUpdateRequest) (*dto.ProductResponse, error) {
	args := p.Called(ctx, productID, req)
	return args.Get(0).(*dto.ProductResponse), args.Error(1)
}
