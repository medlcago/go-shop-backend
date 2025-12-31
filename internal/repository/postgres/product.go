package postgres

import (
	"context"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/pkg/transaction"
	"time"

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

func (p *productRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	db := p.getQueryer(ctx)

	builder := p.flavor.NewSelectBuilder()
	builder.Select(`
		p.*,
		COALESCE(array_agg(pc.category_id) FILTER (WHERE pc.category_id IS NOT NULL), '{}') AS categories`)
	builder.From("products p")
	builder.JoinWithOption(sqlbuilder.LeftJoin, "product_categories pc", "p.id = pc.product_id")
	builder.Where(builder.Equal("p.id", id))
	builder.GroupBy("p.id")

	query, args := builder.Build()

	var product models.Product
	err := db.GetContext(ctx, &product, query, args...)
	if err != nil {
		return nil, repository.HandleSQLError(err)
	}

	return &product, nil
}

func (p *productRepository) ListProducts(ctx context.Context, req dto.ListProductRequest) ([]*models.Product, int64, error) {
	db := p.getQueryer(ctx)

	countBuilder := p.flavor.NewSelectBuilder()
	countBuilder.Select("COUNT(DISTINCT p.id)")
	countBuilder.From("products p")

	filters := []string{
		countBuilder.Equal("p.is_active", true),
		countBuilder.GT("p.stock", 0),
	}

	if req.CategoryID != uuid.Nil {
		countBuilder.Join("product_categories pc", "p.id = pc.product_id")
		filters = append(filters, countBuilder.EQ("pc.category_id", req.CategoryID))
	}

	countBuilder.Where(filters...)

	productBuilder := p.flavor.NewSelectBuilder()
	productBuilder.Select(`
		p.*,
		COALESCE(array_agg(pc.category_id) FILTER (WHERE pc.category_id IS NOT NULL), '{}') AS categories`)
	productBuilder.From("products p")
	productBuilder.JoinWithOption(sqlbuilder.LeftJoin, "product_categories pc", "p.id = pc.product_id")
	productBuilder.GroupBy("p.id")

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

	return repository.PaginatedQuery[models.Product](
		ctx,
		db,
		countBuilder,
		productBuilder,
		req.Limit,
		req.Offset,
	)
}

func (p *productRepository) CreateProduct(ctx context.Context, product *models.Product) error {
	db := p.getQueryer(ctx)

	query := `INSERT INTO products (name, description, price, stock, slug, is_active) VALUES ($1, $2, $3, $4, $5, $6) RETURNING *`
	err := db.GetContext(ctx, product, query,
		product.Name,
		product.Description,
		product.Price,
		product.Stock,
		product.Slug,
		product.IsActive,
	)

	return repository.HandleSQLError(err)
}

func (p *productRepository) UpdateProduct(ctx context.Context, product *models.Product) error {
	db := p.getQueryer(ctx)

	ub := p.flavor.NewUpdateBuilder()
	ub.Update("products")
	ub.Set(
		ub.Assign("name", product.Name),
		ub.Assign("description", product.Description),
		ub.Assign("price", product.Price),
		ub.Assign("stock", product.Stock),
		ub.Assign("is_active", product.IsActive),
		ub.Assign("slug", product.Slug),
		ub.Assign("updated_at", time.Now()),
	)
	ub.Where(ub.Equal("id", product.ID)).Returning("*")

	query, args := ub.Build()

	err := db.GetContext(ctx, product, query, args...)
	return repository.HandleSQLError(err)
}
