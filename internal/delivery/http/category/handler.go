package category

import (
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/service"
	"go-shop-backend/pkg/response"

	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	categoryService service.CategoryService
}

func NewHandler(categoryService service.CategoryService) *Handler {
	return &Handler{
		categoryService: categoryService,
	}
}

func (h *Handler) ListCategories(ctx fiber.Ctx) error {
	var req dto.ListCategoryRequest
	if err := ctx.Bind().Query(&req); err != nil {
		return err
	}

	id := ctx.Params("id")
	req.ID = id

	resp, total, err := h.categoryService.ListCategories(ctx, req)
	if err != nil {
		return err
	}

	return response.PaginatedJSON(ctx, fiber.StatusOK, resp, total)
}
