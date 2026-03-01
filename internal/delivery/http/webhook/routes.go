package webhook

import (
	"go-shop-backend/pkg/middleware"

	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(r fiber.Router, webhookHandler *Handler) {
	webhookGroup := r.Group("/webhook")
	{
		webhookGroup.Post(
			"/yookassa",
			middleware.YookassaIPWhitelist(),
			webhookHandler.YookassaWebhook,
		)
	}
}
