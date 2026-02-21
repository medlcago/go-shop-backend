package service

import (
	"context"
	"errors"
	"fmt"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/pkg/apperrors"
	"go-shop-backend/pkg/mapper"
	"go-shop-backend/pkg/utils"

	"github.com/google/uuid"
)

type productService struct {
	productRepo repository.ProductRepository
	urlBuilder  PublicURLBuilder
}

func NewProductService(
	productRepo repository.ProductRepository,
	urlBuilder PublicURLBuilder,
) *productService {
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

	response, err := mapper.MapOne[*models.Product, dto.ProductResponse](product)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to map product: %w", op, err)
	}

	p.attachImageURLs(ctx, response.Images, product.Images)

	return response, nil
}

func (p *productService) ListProducts(ctx context.Context, req dto.ListProductRequest) ([]*dto.ProductResponse, int64, error) {
	const op = "productService.ListProducts"

	products, total, err := p.productRepo.ListProducts(ctx, req)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}

	response, err := mapper.MapList[*models.Product, *dto.ProductResponse](products)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: failed to map products: %w", op, err)
	}

	for i := range response {
		p.attachImageURLs(ctx, response[i].Images, products[i].Images)
	}

	return response, total, nil
}

func (p *productService) CreateProduct(ctx context.Context, req dto.ProductCreateRequest) (*dto.ProductResponse, error) {
	const op = "productService.CreateProduct"

	product := &models.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		IsActive:    true,
	}

	if req.IsActive != nil {
		product.IsActive = *req.IsActive
	}

	product.Slug = utils.Slugify(product.Name)

	err := p.productRepo.Create(ctx, product)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	response, err := mapper.MapOne[*models.Product, dto.ProductResponse](product)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to map product: %w", op, err)
	}

	return response, nil
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

	if err := p.productRepo.Update(ctx, product); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	response, err := mapper.MapOne[*models.Product, dto.ProductResponse](product)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to map product: %w", op, err)
	}

	p.attachImageURLs(ctx, response.Images, product.Images)

	return response, nil
}

func (p *productService) Search(ctx context.Context, req dto.SearchProductRequest) ([]*dto.ProductResponse, int64, error) {
	const op = "productService.Search"

	products, total, err := p.productRepo.Search(ctx, req)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}

	response, err := mapper.MapList[*models.Product, *dto.ProductResponse](products)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: failed to map products: %w", op, err)
	}

	for i := range response {
		p.attachImageURLs(ctx, response[i].Images, products[i].Images)
	}

	return response, total, nil
}
