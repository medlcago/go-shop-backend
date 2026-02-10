package gorm

import (
	"context"
	"database/sql"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/pkg/database"
	"go-shop-backend/pkg/paging"
	"go-shop-backend/pkg/utils"

	"github.com/google/uuid"
)

var productAllowedOrderBy = map[string]struct{}{
	"id":         {},
	"created_at": {},
	"price":      {},
}

type productRepository struct {
	db database.Provider
}

func NewProductRepository(db database.Provider) repository.ProductRepository {
	return &productRepository{
		db: db,
	}
}

func (p *productRepository) GetByID(ctx context.Context, id uuid.UUID, preload bool) (*models.Product, error) {
	db := p.db.GetDB(ctx)

	db = db.Where("id = ?", id)

	if preload {
		db = db.Preload("Categories").Preload("Images")
	}

	var product models.Product
	if err := db.First(&product).Error; err != nil {
		return nil, repository.HandleSQLError(err)
	}

	return &product, nil
}

func (p *productRepository) ListProducts(ctx context.Context, req dto.ListProductRequest) ([]*models.Product, int64, error) {
	db := p.db.GetDB(ctx)

	query := db.Model(&models.Product{}).
		Where("is_active = ?", true).
		Where("stock > ?", 0)

	if req.CategoryID != uuid.Nil {
		query = query.
			Joins("JOIN product_categories pc ON pc.product_id = products.id").
			Where("pc.category_id = ?", req.CategoryID)
	}

	var total int64
	if err := query.
		Group("products.id").
		Count(&total).Error; err != nil {
		return nil, 0, repository.HandleSQLError(err)
	}

	if total == 0 {
		return nil, 0, nil
	}

	order := "products.created_at"
	if _, ok := productAllowedOrderBy[req.OrderBy]; ok {
		order = "products." + req.OrderBy
		if req.OrderDesc {
			order += " DESC"
		}
	}

	pagination := paging.New(req.Limit, req.Offset)

	var products []*models.Product

	if err := query.
		Group("products.id").
		Limit(pagination.Limit).
		Offset(pagination.Offset).
		Order(order).
		Preload("Categories").
		Preload("Images").
		Find(&products).Error; err != nil {
		return nil, 0, repository.HandleSQLError(err)
	}

	return products, total, nil
}

func (p *productRepository) CreateProduct(ctx context.Context, product *models.Product) error {
	db := p.db.GetDB(ctx)

	err := db.Create(product).Error
	return repository.HandleSQLError(err)
}

func (p *productRepository) UpdateProduct(ctx context.Context, product *models.Product) error {
	db := p.db.GetDB(ctx)

	err := db.Updates(product).Error
	return repository.HandleSQLError(err)
}
func (p *productRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	db := p.db.GetDB(ctx)

	var exists bool
	if err := db.Raw("SELECT EXISTS(SELECT 1 FROM products WHERE id = ?)", id).Scan(&exists).Error; err != nil {
		return false, repository.HandleSQLError(err)
	}

	return exists, nil
}

func (p *productRepository) Search(ctx context.Context, req dto.SearchProductRequest) ([]*models.Product, int64, error) {
	db := p.db.GetDB(ctx)

	if req.Query == "" {
		return nil, 0, nil
	}

	tsQuery := utils.BuildSearchQuery(req.Query)

	fullTsQuery := "to_tsquery('english', @query)"

	namedQuery := sql.Named("query", tsQuery)

	db = db.Model(&models.Product{}).
		Where("is_active = ?", true).
		Where("stock > ?", 0).
		Where("search_vector @@ "+fullTsQuery, namedQuery)

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, repository.HandleSQLError(err)
	}

	if total == 0 {
		return nil, 0, nil
	}

	pagination := paging.New(req.Limit, req.Offset)

	var products []*models.Product

	err := db.Select("products.*, ts_rank(search_vector, "+fullTsQuery+") AS rank", namedQuery).
		Order("rank DESC").
		Limit(pagination.Limit).
		Offset(pagination.Offset).
		Find(&products).Error

	if err != nil {
		return nil, 0, repository.HandleSQLError(err)
	}

	return products, total, nil
}
