package user

import (
	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(r fiber.Router, userHandler *Handler, authMiddleware fiber.Handler) {
	userGroup := r.Group("/users")
	{
		userGroup.Get("/me", authMiddleware, userHandler.GetMe)
	}
}
