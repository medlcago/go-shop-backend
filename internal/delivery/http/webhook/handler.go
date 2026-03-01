package webhook

import (
	"go-shop-backend/internal/service"

	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	orderService service.OrderService
}

func NewHandler(orderService service.OrderService) *Handler {
	return &Handler{
		orderService: orderService,
	}
}

func (h *Handler) YookassaWebhook(ctx fiber.Ctx) error {
	err := h.orderService.HandlePaymentWebhook(ctx, ctx.Body())
	if err != nil {
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	return ctx.SendStatus(fiber.StatusOK)
}
