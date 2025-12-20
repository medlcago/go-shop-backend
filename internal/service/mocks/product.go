package mocks

import (
	"context"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

var _ service.ProductService = (*ProductServiceMock)(nil)

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
