package user

import (
	"go-shop-backend/pkg/middleware"

	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(r fiber.Router, userHandler *Handler) {
	userGroup := r.Group("/users")
	{
		userGroup.Get("/me", middleware.RequireAuth(), userHandler.GetMe)
	}
}
