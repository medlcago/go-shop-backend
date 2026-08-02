package payment

import (
	"go-shop-backend/pkg/middleware"
	"go-shop-backend/pkg/token"

	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(r fiber.Router, paymentHandler *Handler) {
	paymentGroup := r.Group("/payments")
	{
		paymentGroup.Post(
			"/webhook/yookassa",
			middleware.YookassaIPWhitelist(),
			paymentHandler.HandleYookassaWebhook,
		)
	}

	protectedPaymentGroup := paymentGroup.Group(
		"/",
		middleware.RequireAuth(),
		middleware.RequireTokenType(token.AccessTokenType),
	)
	{
		protectedPaymentGroup.Post(
			"/",
			paymentHandler.CreatePayment,
		)
		protectedPaymentGroup.Get(
			"/user-methods",
			paymentHandler.GetUserPaymentMethods,
		)
		protectedPaymentGroup.Put(
			":id<guid>/default",
			paymentHandler.SetDefaultPaymentMethod,
		)
		protectedPaymentGroup.Delete(
			":id<guid>",
			paymentHandler.DeletePaymentMethod,
		)
	}
}
