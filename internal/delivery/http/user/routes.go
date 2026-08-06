package user

import (
	"go-shop-backend/pkg/middleware"
	"go-shop-backend/pkg/token"

	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(r fiber.Router, userHandler *Handler) {
	// auth
	authGroup := r.Group("/auth")
	{
		authGroup.Post("/login", userHandler.Login)
		authGroup.Post("/register", userHandler.Register)
		authGroup.Post(
			"/refresh",
			middleware.RequireTokenType(token.RefreshTokenType),
			userHandler.RefreshToken,
		)

		authGroup.Post(
			"/passkeys/register/begin",
			middleware.RequireAuth(),
			middleware.RequireTokenType(token.AccessTokenType),
			userHandler.BeginPasskeyRegistration,
		)

		authGroup.Post(
			"/passkeys/register/finish",
			middleware.RequireAuth(),
			middleware.RequireTokenType(token.AccessTokenType),
			userHandler.FinishPasskeyRegistration,
		)

		authGroup.Post("/passkeys/login/begin", userHandler.BeginPasskeyLogin)
		authGroup.Post("/passkeys/login/finish", userHandler.FinishPasskeyLogin)

		authGroup.Get(
			"/passkeys",
			middleware.RequireAuth(),
			middleware.RequireTokenType(token.AccessTokenType),
			userHandler.GetUserPasskeys,
		)

		authGroup.Put(
			"/passkeys/:id<guid>",
			middleware.RequireAuth(),
			middleware.RequireTokenType(token.AccessTokenType),
			userHandler.UpdatePasskeyName,
		)

		authGroup.Delete(
			"/passkeys/:id<guid>",
			middleware.RequireAuth(),
			middleware.RequireTokenType(token.AccessTokenType),
			userHandler.DeletePasskey,
		)
	}

	protectedAuthGroup := authGroup.Group(
		"/",
		middleware.RequireAuth(),
		middleware.RequireTokenType(token.AccessTokenType),
	)
	{
		protectedAuthGroup.Post("/setup-2fa", userHandler.Setup2FA)
		protectedAuthGroup.Post("/confirm-2fa", userHandler.Confirm2FA)
		protectedAuthGroup.Post("/disable-2fa", userHandler.Disable2FA)
	}

	// users
	protectedUserGroup := r.Group(
		"/users",
		middleware.RequireAuth(),
		middleware.RequireTokenType(token.AccessTokenType),
	)
	{
		protectedUserGroup.Get("/me", userHandler.GetMe)
		protectedUserGroup.Post("/me/send-email-confirmation", userHandler.SendEmailConfirmationCode)
		protectedUserGroup.Post("/me/confirm-email", userHandler.ConfirmEmail)
		protectedUserGroup.Post("/me/change-password", userHandler.ChangePassword)
	}
}
