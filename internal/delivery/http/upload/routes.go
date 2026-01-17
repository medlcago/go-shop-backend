package upload

import (
	"go-shop-backend/internal/models"
	"go-shop-backend/pkg/middleware"

	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(r fiber.Router, uploadHandler *Handler, authMiddleware fiber.Handler) {
	uploadGroup := r.Group("/uploads")
	{
		uploadGroup.Post(
			"/signurl",
			authMiddleware,
			middleware.RequireRole(models.UserRoleSeller, models.UserRoleAdmin),
			uploadHandler.SignURL,
		)
	}
}
