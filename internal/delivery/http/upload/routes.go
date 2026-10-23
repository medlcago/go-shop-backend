package upload

import (
	"go-shop-backend/pkg/middleware"
	"go-shop-backend/pkg/token"

	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(r fiber.Router, uploadHandler *Handler) {
	uploadGroup := r.Group(
		"/uploads",
		middleware.RequireAuth(),
		middleware.RequireTokenType(token.AccessTokenType),
	)
	{
		uploadGroup.Post("/signurl", uploadHandler.SignURL)
		uploadGroup.Post("/attach", uploadHandler.Attach)
	}
}
