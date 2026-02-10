package cart

import (
	"go-shop-backend/pkg/middleware"

	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(r fiber.Router, cartHandler *Handler, auth middleware.Auth) {
	cartGroup := r.Group("/cart")
	{
		cartGroup.Get("/", auth.Handle(), cartHandler.GetCart)
		cartGroup.Post("/items", auth.Handle(), cartHandler.AddItem)
		cartGroup.Delete("/items/:id<guid>", auth.Handle(), cartHandler.DeleteItem)
	}
}
