package category

import (
	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(r fiber.Router, categoryHandler *Handler) {
	categoryGroup := r.Group("/categories/:id<guid>?")
	{
		categoryGroup.Get("/", categoryHandler.ListCategories)
	}
}
