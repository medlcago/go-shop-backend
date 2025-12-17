package product

import "github.com/gofiber/fiber/v3"

func RegisterRoutes(r fiber.Router, productHandler *Handler) {
	productGroup := r.Group("/products")
	{
		productGroup.Get("/:id<guid>", productHandler.GetProductByID)
		productGroup.Get("/", productHandler.ListProducts)
	}
}
