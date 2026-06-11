package user

import (
	"go-shop-backend/pkg/middleware"
	"go-shop-backend/pkg/token"

	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(r fiber.Router, userHandler *Handler) {
	userGroup := r.Group("/users")
	{
		userGroup.Post("/login", userHandler.Login)
		userGroup.Post("/register", userHandler.Register)
	}

	protectedUserGroup := userGroup.Group(
		"/",
		middleware.RequireAuth(),
		middleware.RequireTokenType(token.AccessTokenType),
	)
	{
		protectedUserGroup.Get("/me", userHandler.GetMe)
		protectedUserGroup.Post("/me/setup-2fa", userHandler.Setup2FA)
		protectedUserGroup.Post("/me/confirm-2fa", userHandler.Confirm2FA)
		protectedUserGroup.Post("/me/disable-2fa", userHandler.Disable2FA)
		protectedUserGroup.Post("/me/send-email-confirmation", userHandler.SendEmailConfirmationCode)
		protectedUserGroup.Post("/me/confirm-email", userHandler.ConfirmEmail)
	}
}
