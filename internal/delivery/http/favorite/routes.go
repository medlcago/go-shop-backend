package favorite

import (
	"go-shop-backend/pkg/middleware"
	"go-shop-backend/pkg/token"

	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(r fiber.Router, favoriteHandler *Handler) {
	favoriteGroup := r.Group(
		"/favorites",
		middleware.RequireAuth(),
		middleware.RequireTokenType(token.AccessTokenType),
	)
	{
		favoriteGroup.Put("/:productID<guid>", favoriteHandler.AddProduct)
		favoriteGroup.Delete("/:productID<guid>", favoriteHandler.RemoveProduct)
		favoriteGroup.Get("/", favoriteHandler.List)
		favoriteGroup.Get("/:productID<guid>", favoriteHandler.IsFavorite)
	}
}
