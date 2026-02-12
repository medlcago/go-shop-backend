package category

import (
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/service"
	"go-shop-backend/pkg/response"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type Handler struct {
	categoryService service.CategoryService
}

func NewHandler(categoryService service.CategoryService) *Handler {
	return &Handler{
		categoryService: categoryService,
	}
}

// ListCategories godoc
//
//	@Summary		List categories or subcategories
//	@Description	Get a paginated list of all categories or subcategories of a specific category. If ID is provided in the path, returns subcategories of that category. Otherwise returns all root categories.
//	@Tags			categories
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string	false	"Parent category ID. If provided, returns subcategories of this category. If omitted, returns all root categories."	Format(uuid)
//	@Param			limit	query		int		false	"Maximum number of items to return"																					minimum(1)	default(50)
//	@Param			offset	query		int		false	"Number of items to skip"																							minimum(0)	default(0)
//	@Success		200		{object}	response.Response[response.PaginatedResponse[[]dto.ProductCategoryResponse]]
//	@Failure		400		{object}	response.Response[any]
//	@Failure		404		{object}	response.Response[any]
//	@Failure		500		{object}	response.Response[any]
//	@Router			/categories/{id} [get]
//	@Router			/categories/ [get]
func (h *Handler) ListCategories(ctx fiber.Ctx) error {
	var req dto.ListCategoryRequest
	if err := ctx.Bind().Query(&req); err != nil {
		return err
	}

	id, err := fiber.Convert(ctx.Params("id"), uuid.Parse, uuid.Nil)
	if err != nil {
		return fiber.ErrBadRequest
	}

	req.ID = id

	resp, total, err := h.categoryService.ListCategories(ctx, req)
	if err != nil {
		return err
	}

	return response.PaginatedJSON(ctx, fiber.StatusOK, resp, total)
}
