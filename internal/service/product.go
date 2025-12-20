package service

import (
	"context"
	"errors"
	"fmt"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/repository"
	"go-shop-backend/pkg/apperrors"
	"go-shop-backend/pkg/utils"

	"github.com/google/uuid"
)

type productService struct {
	productRepo repository.ProductRepository
}

func NewProductService(productRepo repository.ProductRepository) ProductService {
	return &productService{
		productRepo: productRepo,
	}
}

func (p *productService) GetProductByID(ctx context.Context, productID uuid.UUID) (*dto.ProductResponse, error) {
	const op = "productService.GetProductByID"

	product, err := p.productRepo.GetByID(ctx, productID)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return nil, apperrors.ErrProductNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var res dto.ProductResponse
	if err := utils.Copy(&res, product); err != nil {
		return nil, fmt.Errorf("%s: failed to copy product: %w", op, err)
	}

	return &res, nil
}

func (p *productService) ListProducts(ctx context.Context, req dto.ListProductRequest) ([]*dto.ProductResponse, int64, error) {
	const op = "productService.ListProducts"

	products, total, err := p.productRepo.ListProducts(ctx, req)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}

	var resp []*dto.ProductResponse

	if err := utils.Copy(&resp, products); err != nil {
		return nil, 0, fmt.Errorf("%s: failed to copy products: %w", op, err)
	}

	return resp, total, nil
}
