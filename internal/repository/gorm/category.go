package gorm

import (
	"context"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/pkg/database"
	"go-shop-backend/pkg/paging"

	"github.com/google/uuid"
)

type categoryRepository struct {
	db database.Provider
}

func NewCategoryRepository(db database.Provider) repository.CategoryRepository {
	return &categoryRepository{
		db: db,
	}
}

func (c *categoryRepository) ListCategories(ctx context.Context, req dto.ListCategoryRequest) ([]*models.Category, int64, error) {
	db := c.db.GetDB(ctx)

	query := db.Model(&models.Category{})

	if req.ID != uuid.Nil {
		query = query.Where("parent_id = ?", req.ID)
	} else {
		query = query.Where("parent_id IS NULL")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, repository.HandleSQLError(err)
	}

	if total == 0 {
		return nil, 0, nil
	}

	pagination := paging.New(req.Limit, req.Offset)

	var categories []*models.Category
	if err := query.Limit(pagination.Limit).Offset(pagination.Offset).Find(&categories).Error; err != nil {
		return nil, 0, repository.HandleSQLError(err)
	}

	return categories, total, nil
}
