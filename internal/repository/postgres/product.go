package postgres

import (
	"context"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/pkg/paging"
	"go-shop-backend/pkg/transaction"

	"github.com/google/uuid"
	"github.com/huandu/go-sqlbuilder"
)

type productRepository struct {
	getQueryer transaction.QueryerFunc
	flavor     sqlbuilder.Flavor
}

func NewProductRepository(getQueryer transaction.QueryerFunc) repository.ProductRepository {
	return &productRepository{
		getQueryer: getQueryer,
		flavor:     sqlbuilder.PostgreSQL,
	}
}

func (p productRepository) GetByID(ctx context.Context, id string) (*models.Product, error) {
	db := p.getQueryer(ctx)

	query := `SELECT * FROM products WHERE id = $1`

	var product models.Product

	err := db.GetContext(ctx, &product, query, id)
	if err != nil {
		return nil, repository.HandleSQLError(err)
	}

	return &product, nil
}

func (p productRepository) ListProducts(ctx context.Context, req dto.ListProductRequest) ([]*models.Product, int64, error) {
	db := p.getQueryer(ctx)

	productBuilder := p.flavor.NewSelectBuilder()
	productBuilder.Select("p.*").From("products p")

	countBuilder := p.flavor.NewSelectBuilder()
	countBuilder.Select("COUNT(p.id)").From("products p")

	if req.CategoryID != uuid.Nil {
		productBuilder.JoinWithOption(
			sqlbuilder.LeftJoin, "product_categories pc", "p.id = pc.product_id",
		).Where(productBuilder.EQ("pc.category_id", req.CategoryID))

		countBuilder.JoinWithOption(
			sqlbuilder.LeftJoin, "product_categories pc", "p.id = pc.product_id",
		).Where(countBuilder.EQ("pc.category_id", req.CategoryID))
	}

	productBuilder.Where(
		productBuilder.Equal("p.is_active", true),
		productBuilder.GT("p.stock", 0),
	)

	countBuilder.Where(
		countBuilder.Equal("p.is_active", true),
		countBuilder.GT("p.stock", 0),
	)

	countQuery, countArgs := countBuilder.Build()
	var total int64

	err := db.GetContext(ctx, &total, countQuery, countArgs...)
	if err != nil {
		return nil, 0, repository.HandleSQLError(err)
	}

	if total == 0 {
		return []*models.Product{}, 0, nil
	}

	allowedOrderBy := map[string]bool{
		"id":         true,
		"created_at": true,
		"price":      true,
	}

	if allowedOrderBy[req.OrderBy] {
		if req.OrderDesc {
			productBuilder.OrderByDesc(req.OrderBy)
		} else {
			productBuilder.OrderByAsc(req.OrderBy)
		}
	}

	pagination := paging.New(req.Limit, req.Offset)

	productBuilder.Limit(pagination.Limit).Offset(pagination.Offset)

	productsQuery, productsArgs := productBuilder.Build()
	var products []*models.Product

	err = db.SelectContext(ctx, &products, productsQuery, productsArgs...)
	if err != nil {
		return nil, 0, repository.HandleSQLError(err)
	}

	return products, total, nil
}
