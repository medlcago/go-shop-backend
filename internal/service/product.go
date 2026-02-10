package service

import (
	"context"
	"errors"
	"fmt"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/pkg/apperrors"
	"go-shop-backend/pkg/utils"

	"github.com/google/uuid"
)

type productService struct {
	productRepo repository.ProductRepository
	urlBuilder  PublicURLBuilder
}

func NewProductService(productRepo repository.ProductRepository, urlBuilder PublicURLBuilder) ProductService {
	return &productService{
		productRepo: productRepo,
		urlBuilder:  urlBuilder,
	}
}

func (p *productService) attachImageURLs(
	ctx context.Context,
	respImages []dto.UploadResponse,
	modelImages []models.Upload,
) {
	for i := range respImages {
		respImages[i].URL = p.urlBuilder.PublicURL(ctx, modelImages[i].ObjectKey)
	}
}

func (p *productService) GetProductByID(ctx context.Context, productID uuid.UUID) (*dto.ProductResponse, error) {
	const op = "productService.GetProductByID"

	product, err := p.productRepo.GetByID(ctx, productID, true)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return nil, apperrors.ErrProductNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var resp dto.ProductResponse
	if err := utils.Copy(&resp, product); err != nil {
		return nil, fmt.Errorf("%s: failed to copy product: %w", op, err)
	}

	p.attachImageURLs(ctx, resp.Images, product.Images)

	return &resp, nil
}

func (p *productService) ListProducts(ctx context.Context, req dto.ListProductRequest) ([]*dto.ProductResponse, int64, error) {
	const op = "productService.ListProducts"

	products, total, err := p.productRepo.ListProducts(ctx, req)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}

	resp := make([]*dto.ProductResponse, len(products))
	if err := utils.Copy(&resp, products); err != nil {
		return nil, 0, fmt.Errorf("%s: failed to copy products: %w", op, err)
	}

	for i := range resp {
		p.attachImageURLs(ctx, resp[i].Images, products[i].Images)
	}

	return resp, total, nil
}

func (p *productService) CreateProduct(ctx context.Context, req dto.ProductCreateRequest) (*dto.ProductResponse, error) {
	const op = "productService.CreateProduct"

	var product models.Product
	if err := utils.Copy(&product, req); err != nil {
		return nil, fmt.Errorf("%s: failed to copy product: %w", op, err)
	}
	product.IsActive = true

	if req.IsActive != nil {
		product.IsActive = *req.IsActive
	}

	product.Slug = utils.Slugify(product.Name)

	err := p.productRepo.CreateProduct(ctx, &product)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var resp dto.ProductResponse
	if err := utils.Copy(&resp, product); err != nil {
		return nil, fmt.Errorf("%s: failed to copy product: %w", op, err)
	}

	return &resp, nil
}

func (p *productService) UpdateProduct(ctx context.Context, productID uuid.UUID, req dto.ProductUpdateRequest) (*dto.ProductResponse, error) {
	const op = "productService.UpdateProduct"

	product, err := p.productRepo.GetByID(ctx, productID, false)
	if err != nil {
		return nil, apperrors.ErrProductNotFound
	}

	if req.Name != nil {
		product.Name = *req.Name
		product.Slug = utils.Slugify(product.Name)
	}

	if req.Description != nil {
		product.Description = req.Description
	}

	if req.Price != nil {
		product.Price = *req.Price
	}

	if req.Stock != nil {
		product.Stock = *req.Stock
	}

	if req.IsActive != nil {
		product.IsActive = *req.IsActive
	}

	if err := p.productRepo.UpdateProduct(ctx, product); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var resp dto.ProductResponse
	if err := utils.Copy(&resp, product); err != nil {
		return nil, fmt.Errorf("%s: failed to copy product: %w", op, err)
	}

	p.attachImageURLs(ctx, resp.Images, product.Images)

	return &resp, nil
}

func (p *productService) Search(ctx context.Context, req dto.SearchProductRequest) ([]*dto.ProductResponse, int64, error) {
	const op = "productService.Search"

	products, total, err := p.productRepo.Search(ctx, req)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}

	resp := make([]*dto.ProductResponse, len(products))
	if err := utils.Copy(&resp, products); err != nil {
		return nil, 0, fmt.Errorf("%s: failed to copy products: %w", op, err)
	}

	for i := range resp {
		p.attachImageURLs(ctx, resp[i].Images, products[i].Images)
	}

	return resp, total, nil
}
