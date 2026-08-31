package service

import (
	"context"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/internal/upload"
	"go-shop-backend/pkg/apperror"
	"go-shop-backend/pkg/mapper"
	"go-shop-backend/pkg/utils"

	"github.com/google/uuid"
)

const (
	ProductImageType upload.Type = "product_image"
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

	product, err := p.getProductByID(ctx, productID, true)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	response, err := p.mapProduct(product)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	return response, nil
}

func (p *productService) ListProducts(ctx context.Context, req dto.ListProductRequest) ([]*dto.ProductResponse, int64, error) {
	const op = "productService.ListProducts"

	products, total, err := p.productRepo.ListProducts(ctx, req)
	if err != nil {
		return nil, 0, apperror.Wrap(op, err)
	}

	response, err := p.mapProducts(products)
	if err != nil {
		return nil, 0, apperror.Wrap(op, err)
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
		Slug:        utils.Slugify(req.Name),
	}

	if req.IsActive != nil {
		product.IsActive = *req.IsActive
	}

	err := p.productRepo.Create(ctx, product)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	response, err := p.mapProduct(product)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	return response, nil
}

func (p *productService) UpdateProduct(ctx context.Context, productID uuid.UUID, req dto.ProductUpdateRequest) (*dto.ProductResponse, error) {
	const op = "productService.UpdateProduct"

	product, err := p.getProductByID(ctx, productID, false)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	if err := mapper.Copy(product, req, true); err != nil {
		return nil, apperror.Wrap(op, err)
	}
	product.Slug = utils.Slugify(product.Name)

	if err := p.productRepo.Update(ctx, product); err != nil {
		return nil, apperror.Wrap(op, err)
	}

	response, err := p.mapProduct(product)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	return response, nil
}

func (p *productService) UpdateStock(ctx context.Context, productID uuid.UUID, req dto.ProductUpdateStockRequest) (*dto.ProductResponse, error) {
	const op = "productService.UpdateStock"

	product, err := p.getProductByID(ctx, productID, false)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	if req.Stock < product.Reserved {
		return nil, apperror.Wrap(op, apperror.ErrStockLessThanReserved)
	}

	product.Stock = req.Stock

	if err := p.productRepo.Update(ctx, product); err != nil {
		return nil, apperror.Wrap(op, err)
	}

	response, err := p.mapProduct(product)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	return response, nil
}

func (p *productService) Search(ctx context.Context, req dto.SearchProductRequest) ([]*dto.ProductResponse, int64, error) {
	const op = "productService.Search"

	products, total, err := p.productRepo.Search(ctx, req)
	if err != nil {
		return nil, 0, apperror.Wrap(op, err)
	}

	response, err := p.mapProducts(products)
	if err != nil {
		return nil, 0, apperror.Wrap(op, err)
	}

	return response, total, nil
}

func (p *productService) UploadImage(
	ctx context.Context,
	productID uuid.UUID,
	req dto.UploadProductImageSignURLRequest,
) (*dto.GeneratePresignedURLResponse, error) {
	const op = "productService.UploadImage"

	if err := p.productExists(ctx, productID); err != nil {
		return nil, apperror.Wrap(op, err)
	}

	signRequest := dto.GeneratePresignedURLRequest{
		ContentType: req.ContentType,
		Entity:      dto.NewUploadEntity(productID, string(models.EntityTypeProduct)),
		Ext:         req.Ext,
	}

	response, err := p.uploadManager.SignURL(ctx, signRequest, ProductImageType)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	return response, nil
}

func (p *productService) AttachImage(
	ctx context.Context,
	productID uuid.UUID,
	req dto.AttachProductImageRequest,
) (*dto.UploadResponse, error) {
	const op = "productService.AttachImage"

	if err := p.productExists(ctx, productID); err != nil {
		return nil, apperror.Wrap(op, err)
	}

	attachRequest := dto.AttachFileRequest{
		UploadID:  req.UploadID,
		ObjectKey: req.ObjectKey,
		Entity:    dto.NewUploadEntity(productID, string(models.EntityTypeProduct)),
	}

	response, err := p.uploadManager.Attach(ctx, attachRequest, ProductImageType)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	return response, nil
}

func (p *productService) getProductByID(
	ctx context.Context,
	productID uuid.UUID,
	preload bool,
) (*models.Product, error) {
	const op = "productService.getProductByID"

	product, err := p.productRepo.GetByID(ctx, productID, preload)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, apperror.Wrap(op, apperror.ErrProductNotFound)
		}

		return nil, apperror.Wrap(op, err)
	}

	return product, nil
}

func (p *productService) productExists(ctx context.Context, productID uuid.UUID) error {
	const op = "productService.productExists"

	exists, err := p.productRepo.Exists(ctx, productID)
	if err != nil {
		return apperror.Wrap(op, err)
	}

	if !exists {
		return apperror.Wrap(op, apperror.ErrProductNotFound)
	}

	return nil
}

func (p *productService) mapProduct(product *models.Product) (*dto.ProductResponse, error) {
	const op = "productService.mapProduct"

	upload.AssignPublicURLs(product.Images, p.uploadManager)

	response, err := mapper.MapOne[*models.Product, dto.ProductResponse](product)
	if err != nil {
		return nil, apperror.Wrap(op, err)
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
		return nil, apperror.Wrap(op, err)
	}

	return response, nil
}
