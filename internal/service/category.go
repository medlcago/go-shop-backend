package service

import (
	"context"
	"fmt"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/pkg/mapper"
)

type categoryService struct {
	categoryRepo repository.CategoryRepository
}

func NewCategoryService(categoryRepo repository.CategoryRepository) *categoryService {
	return &categoryService{
		categoryRepo: categoryRepo,
	}
}

func (c *categoryService) ListCategories(ctx context.Context, req dto.ListCategoryRequest) ([]*dto.ProductCategoryResponse, int64, error) {
	const op = "categoryService.ListCategories"

	categories, totalCategories, err := c.categoryRepo.ListCategories(ctx, req)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}

	response, err := mapper.MapList[*models.Category, *dto.ProductCategoryResponse](categories)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: failed to map categories: %w", op, err)
	}

	return response, totalCategories, nil
}
