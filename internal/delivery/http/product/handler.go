package product

import (
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/service"
	"go-shop-backend/pkg/response"

	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	productService service.ProductService
}

func NewHandler(productService service.ProductService) *Handler {
	return &Handler{
		productService: productService,
	}
}

func (h *Handler) GetProductByID(ctx fiber.Ctx) error {
	id := ctx.Params("id")

	resp, err := h.productService.GetProductByID(ctx, id)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, resp)
}

func (h *Handler) ListProducts(ctx fiber.Ctx) error {
	var req dto.ListProductRequest

	if err := ctx.Bind().Query(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid query params")
	}

	resp, total, err := h.productService.ListProducts(ctx, req)
	if err != nil {
		return err
	}

	return response.PaginatedJSON(ctx, fiber.StatusOK, resp, total)
}
