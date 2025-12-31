package postgres

import (
	"context"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/pkg/transaction"

	"github.com/google/uuid"
	"github.com/huandu/go-sqlbuilder"
)

type categoryRepository struct {
	getQueryer transaction.QueryerFunc
	flavor     sqlbuilder.Flavor
}

func NewCategoryRepository(getQueryer transaction.QueryerFunc) repository.CategoryRepository {
	return &categoryRepository{
		getQueryer: getQueryer,
		flavor:     sqlbuilder.PostgreSQL,
	}
}

func (c *categoryRepository) ListCategories(ctx context.Context, req dto.ListCategoryRequest) ([]*models.ProductCategory, int64, error) {
	db := c.getQueryer(ctx)

	countBuilder := c.flavor.NewSelectBuilder()
	countBuilder.Select("COUNT(c.id)").From("categories c")

	if req.ID != uuid.Nil {
		countBuilder.Where(countBuilder.Equal("c.parent_id", req.ID))
	} else {
		countBuilder.Where(countBuilder.IsNull("c.parent_id"))
	}

	categoryBuilder := c.flavor.NewSelectBuilder()
	categoryBuilder.
		Select(
			"c.id",
			"c.name",
			"c.slug",
			"c.parent_id",
			`EXISTS (
				SELECT 1
				FROM categories cc
				WHERE cc.parent_id = c.id
			) AS has_children`,
		).
		From("categories c")

	return repository.PaginatedQuery[models.ProductCategory](
		ctx,
		db,
		countBuilder,
		categoryBuilder,
		req.Limit,
		req.Offset,
	)
}
