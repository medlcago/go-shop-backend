package postgres

import (
	"context"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/pkg/paging"
	"go-shop-backend/pkg/transaction"

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

func (c categoryRepository) ListCategories(ctx context.Context, req dto.ListCategoryRequest) ([]*models.ProductCategory, int64, error) {
	db := c.getQueryer(ctx)

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

	countBuilder := c.flavor.NewSelectBuilder()
	countBuilder.Select("COUNT(c.id)").From("categories c")

	if req.ID != "" {
		categoryBuilder.Where(categoryBuilder.Equal("c.parent_id", req.ID))
		countBuilder.Where(countBuilder.Equal("c.parent_id", req.ID))
	} else {
		categoryBuilder.Where(categoryBuilder.IsNull("c.parent_id"))
		countBuilder.Where(countBuilder.IsNull("c.parent_id"))
	}

	countQuery, countArgs := countBuilder.Build()
	var total int64

	err := db.GetContext(ctx, &total, countQuery, countArgs...)
	if err != nil {
		return nil, 0, repository.HandleSQLError(err)
	}

	if total == 0 {
		return []*models.ProductCategory{}, 0, nil
	}

	pagination := paging.New(req.Limit, req.Offset)

	categoryBuilder.Limit(pagination.Limit).Offset(pagination.Offset)

	categoriesQuery, categoriesArgs := categoryBuilder.Build()
	var categories []*models.ProductCategory

	err = db.SelectContext(ctx, &categories, categoriesQuery, categoriesArgs...)
	if err != nil {
		return nil, 0, repository.HandleSQLError(err)
	}

	return categories, total, nil

}
