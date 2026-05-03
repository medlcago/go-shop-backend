package user

import (
	"go-shop-backend/pkg/middleware"

	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(r fiber.Router, userHandler *Handler) {
	userGroup := r.Group("/users", middleware.RequireAuth())
	{
		userGroup.Get("/me", userHandler.GetMe)

		userGroup.Post("/email/confirmation", userHandler.EmailConfirmation)
		userGroup.Post("/email/confirmation/confirm", userHandler.ConfirmEmail)
	}
}
