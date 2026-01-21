package user

import (
	"go-shop-backend/pkg/middleware"

	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(r fiber.Router, userHandler *Handler, auth middleware.Auth) {
	userGroup := r.Group("/users")
	{
		userGroup.Get("/me", auth.Handle(), userHandler.GetMe)
	}
}
