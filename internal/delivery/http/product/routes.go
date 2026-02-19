package product

import (
	"go-shop-backend/internal/models"
	"go-shop-backend/pkg/middleware"

	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(r fiber.Router, productHandler *Handler) {
	productGroup := r.Group("/products")
	{
		productGroup.Get("/:id<guid>", productHandler.GetProductByID)
		productGroup.Get("/", productHandler.ListProducts)
		productGroup.Post("/",
			middleware.RequireAuth(),
			middleware.RequireRole(models.UserRoleSeller, models.UserRoleAdmin),
			productHandler.CreateProduct,
		)
		productGroup.Patch("/:id<guid>",
			middleware.RequireAuth(),
			middleware.RequireRole(models.UserRoleSeller, models.UserRoleAdmin),
			productHandler.UpdateProduct,
		)
		productGroup.Get("/search", productHandler.Search)
	}
}
