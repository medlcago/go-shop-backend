package gorm

import (
	"context"
	"database/sql"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/internal/repository/gorm/scopes"
	"go-shop-backend/pkg/database"
	"go-shop-backend/pkg/utils"

	"github.com/google/uuid"
	"gorm.io/gorm/clause"
)

type productRepository struct {
	db database.Provider
}

func NewProductRepository(db database.Provider) *productRepository {
	return &productRepository{
		db: db,
	}
}

func (p *productRepository) GetByID(ctx context.Context, id uuid.UUID, preload bool) (*models.Product, error) {
	db := p.db.GetDB(ctx)

	db = db.Where("id = ?", id)

	if preload {
		db = db.Scopes(
			scopes.ProductWithRelations(),
		)
	}

	var product models.Product
	if err := db.First(&product).Error; err != nil {
		return nil, repository.HandleError(err)
	}

	return &product, nil
}

func (p *productRepository) ListProducts(ctx context.Context, req dto.ListProductRequest) ([]*models.Product, int64, error) {
	db := p.db.GetDB(ctx)

	db = db.Model(&models.Product{}).
		Group("products.id").
		Scopes(
			scopes.AvailableProducts(),
			scopes.ProductWithCategory(req.CategoryID),
		)

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, repository.HandleError(err)
	}

	if total == 0 {
		return nil, 0, nil
	}

	var products []*models.Product

	if err := db.
		Scopes(
			scopes.Paginate(req.Limit, req.Offset),
			scopes.ProductOrderBy(req.OrderBy, req.OrderDesc),
			scopes.ProductWithRelations(),
		).
		Find(&products).Error; err != nil {
		return nil, 0, repository.HandleError(err)
	}

	return products, total, nil
}

func (p *productRepository) Create(ctx context.Context, product *models.Product) error {
	db := p.db.GetDB(ctx)

	err := db.Create(product).Error
	return repository.HandleError(err)
}

func (p *productRepository) Update(ctx context.Context, product *models.Product) error {
	db := p.db.GetDB(ctx)

	err := db.Select("*").Updates(product).Error
	return repository.HandleError(err)
}

func (p *productRepository) BulkUpsert(ctx context.Context, products []*models.Product) error {
	db := p.db.GetDB(ctx)

	err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(products).Error

	return repository.HandleError(err)
}

func (p *productRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	db := p.db.GetDB(ctx)

	var exists bool
	if err := db.Raw("SELECT EXISTS(SELECT 1 FROM products WHERE id = ?)", id).Scan(&exists).Error; err != nil {
		return false, repository.HandleError(err)
	}

	return exists, nil
}

func (p *productRepository) GetByIDsForUpdate(ctx context.Context, ids []uuid.UUID) ([]*models.Product, error) {
	db := p.db.GetDB(ctx)

	var products []*models.Product
	if err := db.Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate}).
		Where("id IN (?)", ids).
		Find(&products).Error; err != nil {
		return nil, repository.HandleError(err)
	}

	return products, nil
}

func (p *productRepository) Search(ctx context.Context, req dto.SearchProductRequest) ([]*models.Product, int64, error) {
	if req.Query == "" {
		return nil, 0, nil
	}

	db := p.db.GetDB(ctx)

	query := utils.BuildSearchQuery(req.Query)
	tsQuery := "to_tsquery('english', @query)"
	namedQuery := sql.Named("query", query)

	db = db.Model(&models.Product{}).
		Scopes(
			scopes.AvailableProducts(),
		).
		Where("search_vector @@ "+tsQuery, namedQuery)

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, repository.HandleError(err)
	}

	if total == 0 {
		return nil, 0, nil
	}

	var products []*models.Product

	err := db.Select("products.*, ts_rank(search_vector, "+tsQuery+") AS rank", namedQuery).
		Order("rank DESC").
		Scopes(
			scopes.Paginate(req.Limit, req.Offset),
			scopes.ProductWithRelations(),
		).
		Find(&products).Error

	if err != nil {
		return nil, 0, repository.HandleError(err)
	}

	return products, total, nil
}
