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
		productGroup.Get("/search", productHandler.Search)

		protectedProductGroup := productGroup.Group(
			"/",
			middleware.RequireAuth(),
			middleware.RequireRole(models.UserRoleSeller, models.UserRoleAdmin),
		)
		{
			protectedProductGroup.Post("/", productHandler.CreateProduct)
			protectedProductGroup.Patch("/:id<guid>", productHandler.UpdateProduct)
			protectedProductGroup.Post("/:id<guid>/images/upload-url", productHandler.UploadImage)
			protectedProductGroup.Post("/:id<guid>/images", productHandler.ConfirmUploadImage)
		}
	}
}
