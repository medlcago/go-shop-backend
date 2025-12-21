package product

import (
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/service"
	"go-shop-backend/pkg/response"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type Handler struct {
	productService service.ProductService
}

func NewHandler(productService service.ProductService) *Handler {
	return &Handler{
		productService: productService,
	}
}

// GetProductByID godoc
//
//	@Summary		Get product by ID
//	@Description	Get detailed information about a specific product by its UUID
//	@Tags			products
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Product ID"	Format(uuid)
//	@Success		200	{object}	response.Response[dto.ProductResponse]
//	@Failure		400	{object}	response.Response[any]
//	@Failure		404	{object}	response.Response[any]
//	@Failure		500	{object}	response.Response[any]
//	@Router			/products/{id} [get]
func (h *Handler) GetProductByID(ctx fiber.Ctx) error {
	id, err := fiber.Convert(ctx.Params("id"), uuid.Parse, uuid.Nil)
	if err != nil {
		return fiber.ErrBadRequest
	}

	resp, err := h.productService.GetProductByID(ctx, id)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, resp)
}

// ListProducts godoc
//
//	@Summary		List products with filtering and pagination
//	@Description	Get a paginated list of products with optional filtering by category and sorting
//	@Tags			products
//	@Accept			json
//	@Produce		json
//	@Param			limit		query		int		false	"Maximum number of items to return"	minimum(1)	default(50)
//	@Param			offset		query		int		false	"Number of items to skip"			minimum(0)	default(0)
//	@Param			order_by	query		string	false	"Field to sort by"					Enums(name, created_at, price)
//	@Param			order_desc	query		bool	false	"Sort in descending order"			default(false)
//	@Param			category_id	query		string	false	"Filter by category ID"				Format(uuid)
//	@Success		200			{object}	response.Response[response.PaginatedResponse[[]dto.ProductResponse]]
//	@Failure		400			{object}	response.Response[any]
//	@Failure		500			{object}	response.Response[any]
//	@Router			/products/ [get]
func (h *Handler) ListProducts(ctx fiber.Ctx) error {
	var req dto.ListProductRequest

	if err := ctx.Bind().Query(&req); err != nil {
		return fiber.ErrBadRequest
	}

	resp, total, err := h.productService.ListProducts(ctx, req)
	if err != nil {
		return err
	}

	return response.PaginatedJSON(ctx, fiber.StatusOK, resp, total)
}

// CreateProduct godoc
//
//	@Summary		Create a new product
//	@Description	Create a new product with the provided details including categories
//	@Tags			products
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.ProductCreateRequest	true	"Product creation data"
//	@Success		201		{object}	response.Response[dto.ProductResponse]
//	@Failure		400		{object}	response.Response[any]	"Invalid request data or validation failed"
//	@Failure		401		{object}	response.Response[any]	"Unauthorized (e.g., invalid token)"
//	@Failure		403		{object}	response.Response[any]	"Access denied (e.g., no access rights)"
//	@Failure		500		{object}	response.Response[any]	"Internal server error"
//	@Router			/products [post]
//
//	@Security		BearerAuth
//
//	@Description	## Fields Description
//	@Description	- **name** (required): Product name, 2-100 characters
//	@Description	- **description** (optional): Product description, 5-500 characters if provided
//	@Description	- **price** (required): Product price, must be greater than 0
//	@Description	- **stock** (required): Initial stock quantity, must be greater than 0
//	@Description	- **is_active** (optional): Whether product is active for sale. Defaults to true if not specified
//
//	@Description	## Example Request
//	@Description	```json
//	@Description	{
//	@Description	"name": "Premium Laptop",
//	@Description	"description": "High-performance laptop with 16GB RAM and 512GB SSD",
//	@Description	"price": 1299.99,
//	@Description	"stock": 50,
//	@Description	"is_active": true,
//	@Description	}
//	@Description	```
func (h *Handler) CreateProduct(ctx fiber.Ctx) error {
	var req dto.ProductCreateRequest

	if err := ctx.Bind().JSON(&req); err != nil {
		return fiber.ErrBadRequest
	}

	resp, err := h.productService.CreateProduct(ctx, req)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusCreated, resp)
}
