package gorm

import (
	"context"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/internal/repository/gorm/scopes"
	"go-shop-backend/pkg/database"
)

type categoryRepository struct {
	db database.Provider
}

func NewCategoryRepository(db database.Provider) *categoryRepository {
	return &categoryRepository{
		db: db,
	}
}

func (c *categoryRepository) ListCategories(ctx context.Context, req dto.ListCategoryRequest) ([]*models.Category, int64, error) {
	db := c.db.GetDB(ctx)

	db = db.Model(&models.Category{}).
		Scopes(
			scopes.CategoryWithParentID(req.ID),
		)

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, repository.HandleError(err)
	}

	if total == 0 {
		return nil, 0, nil
	}

	var categories []*models.Category
	if err := db.
		Scopes(
			scopes.Paginate(req.Limit, req.Offset),
		).
		Select(`
		categories.*,
		EXISTS (
			SELECT 1
			FROM categories c2
			WHERE c2.parent_id = categories.id
			AND c2.deleted_at IS NULL
		) AS has_children
	`).
		Order("sort_order DESC, created_at DESC").
		Find(&categories).Error; err != nil {
		return nil, 0, repository.HandleError(err)
	}

	return categories, total, nil
}
