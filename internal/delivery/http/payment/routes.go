package payment

import (
	"go-shop-backend/pkg/middleware"

	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(r fiber.Router, paymentHandler *Handler) {
	paymentGroup := r.Group("/payments")
	{
		paymentGroup.Post("/", middleware.RequireAuth(), paymentHandler.CreatePayment)
		paymentGroup.Post("/webhook/yookassa", middleware.YookassaIPWhitelist(), paymentHandler.HandleYookassaWebhook)
	}
}
