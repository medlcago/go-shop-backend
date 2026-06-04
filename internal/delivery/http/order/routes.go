package order

import (
	"go-shop-backend/pkg/middleware"

	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(r fiber.Router, orderHandler *Handler) {
	orderGroup := r.Group("/orders")
	orderGroup.Use(middleware.RequireSessionID())
	{
		orderGroup.Post("/", orderHandler.CreateOrder)
		orderGroup.Get("/", orderHandler.GetOrders)
		orderGroup.Get("/:id<guid>", orderHandler.GetOrder)
		orderGroup.Post("/:id<guid>/items", orderHandler.AddItem)
		orderGroup.Delete("/:id<guid>/items/:itemID<guid>", orderHandler.RemoveItem)
		orderGroup.Delete("/:id<guid>/items", orderHandler.ClearItems)

		protectedOrderGroup := orderGroup.Group("/", middleware.RequireAuth())
		{
			protectedOrderGroup.Post("/:id<guid>/checkout", orderHandler.Checkout)
			protectedOrderGroup.Post("/:id<guid>/cancel", orderHandler.CancelOrder)
		}
	}
}
