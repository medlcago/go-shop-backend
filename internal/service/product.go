package service

import (
	"context"
	"errors"
	"fmt"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/internal/upload"
	"go-shop-backend/pkg/apperrors"
	"go-shop-backend/pkg/mapper"
	"go-shop-backend/pkg/utils"

	"github.com/google/uuid"
)

type productService struct {
	productRepo   repository.ProductRepository
	uploadManager upload.Manager
}

func NewProductService(
	productRepo repository.ProductRepository,
	uploadManager upload.Manager,
) *productService {
	return &productService{
		productRepo:   productRepo,
		uploadManager: uploadManager,
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

	response, err := p.mapProduct(product)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return response, nil
}

func (p *productService) ListProducts(ctx context.Context, req dto.ListProductRequest) ([]*dto.ProductResponse, int64, error) {
	const op = "productService.ListProducts"

	products, total, err := p.productRepo.ListProducts(ctx, req)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}

	response, err := p.mapProducts(products)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", op, err)
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

	response, err := p.mapProduct(product)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
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

	response, err := p.mapProduct(product)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return response, nil
}

func (p *productService) Search(ctx context.Context, req dto.SearchProductRequest) ([]*dto.ProductResponse, int64, error) {
	const op = "productService.Search"

	products, total, err := p.productRepo.Search(ctx, req)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}

	response, err := p.mapProducts(products)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}

	return response, total, nil
}

func (p *productService) UploadImage(
	ctx context.Context,
	productID uuid.UUID,
	req dto.UploadProductImageRequest,
) (*dto.SignURLResponse, error) {
	const op = "productService.UploadProductImage"

	exists, err := p.productRepo.Exists(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if !exists {
		return nil, apperrors.ErrProductNotFound
	}

	signUrlReq := upload.SignURLRequest{
		ContentType: req.ContentType,
		Entity:      upload.NewProductEntity(productID),
		Ext:         req.Ext,
	}

	signUrlResp, err := p.uploadManager.SignURL(ctx, signUrlReq, upload.ProductImagePolicy)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	response, err := mapper.MapOne[*upload.SignURLResponse, dto.SignURLResponse](signUrlResp)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return response, nil
}

func (p *productService) ConfirmUploadImage(
	ctx context.Context,
	productID uuid.UUID,
	req dto.ConfirmUploadProductImageRequest,
) (*dto.UploadResponse, error) {
	const op = "productService.ConfirmUploadImage"

	exists, err := p.productRepo.Exists(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if !exists {
		return nil, apperrors.ErrProductNotFound
	}

	saveUploadReq := upload.SaveUploadRequest{
		UploadID:  req.UploadID,
		ObjectKey: req.ObjectKey,
		Entity:    upload.NewProductEntity(productID),
		IsMain:    false,
	}

	saveUploadResp, err := p.uploadManager.Save(ctx, saveUploadReq, upload.ProductImagePolicy)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	response, err := mapper.MapOne[*upload.ContentResponse, dto.UploadResponse](saveUploadResp)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return response, nil
}

func (p *productService) mapProduct(product *models.Product) (*dto.ProductResponse, error) {
	const op = "productService.mapProduct"

	upload.AssignPublicURLs(product.Images, p.uploadManager)

	response, err := mapper.MapOne[*models.Product, dto.ProductResponse](product)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return response, nil
}

func (p *productService) mapProducts(products []*models.Product) ([]*dto.ProductResponse, error) {
	const op = "productService.mapProducts"

	for _, product := range products {
		upload.AssignPublicURLs(product.Images, p.uploadManager)
	}

	response, err := mapper.MapList[*models.Product, *dto.ProductResponse](products)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return response, nil
}
