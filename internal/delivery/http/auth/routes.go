package auth

import (
	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(r fiber.Router, authHandler *Handler) {
	authGroup := r.Group("/auth")
	{
		authGroup.Post("/login", authHandler.Login)
		authGroup.Post("/register", authHandler.Register)
	}
}
