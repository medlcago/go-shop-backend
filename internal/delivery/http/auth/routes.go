package auth

import (
	"go-shop-backend/pkg/middleware"

	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(r fiber.Router, authHandler *Handler) {
	authGroup := r.Group("/auth")
	{
		authGroup.Post("/login", authHandler.Login)
		authGroup.Post("/login/2fa", authHandler.LoginWith2FA)
		authGroup.Post("/register", authHandler.Register)

		twoFAGroup := authGroup.Group("/2fa", middleware.RequireAuth())
		{
			twoFAGroup.Post("/setup", authHandler.Setup2FA)
			twoFAGroup.Post("/confirm", authHandler.Confirm2FA)
			twoFAGroup.Post("/disable", authHandler.Disable2FA)
		}
	}
}
