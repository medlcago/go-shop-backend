package upload

import (
	"go-shop-backend/internal/models"
	"go-shop-backend/pkg/middleware"

	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(r fiber.Router, uploadHandler *Handler, auth middleware.Auth) {
	uploadGroup := r.Group("/uploads")
	{
		uploadGroup.Post(
			"/signurl",
			auth.Handle(),
			middleware.RequireRole(models.UserRoleSeller, models.UserRoleAdmin),
			uploadHandler.SignURL,
		)
		uploadGroup.Post(
			"/save",
			auth.Handle(),
			middleware.RequireRole(models.UserRoleSeller, models.UserRoleAdmin),
			uploadHandler.Save,
		)
	}
}
