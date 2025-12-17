package service

import (
	"context"
	"fmt"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/repository"
	"go-shop-backend/pkg/utils"
)

type categoryService struct {
	categoryRepo repository.CategoryRepository
}

func NewCategoryService(categoryRepo repository.CategoryRepository) CategoryService {
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

	var res []*dto.ProductCategoryResponse
	if err := utils.Copy(&res, categories); err != nil {
		return nil, 0, fmt.Errorf("%s: failed to copy categories: %w", op, err)
	}

	return res, totalCategories, nil
}
