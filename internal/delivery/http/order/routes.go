package order

import (
	"go-shop-backend/pkg/middleware"
	"go-shop-backend/pkg/token"

	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(r fiber.Router, orderHandler *Handler) {
	orderGroup := r.Group(
		"/orders",
		middleware.RequireSessionID(),
	)
	{
		orderGroup.Post("/", orderHandler.CreateOrder)
		orderGroup.Get("/", orderHandler.GetOrders)
		orderGroup.Get("/:id<guid>", orderHandler.GetOrder)
		orderGroup.Post("/:id<guid>/items", orderHandler.AddItem)
		orderGroup.Delete("/:id<guid>/items/:itemID<guid>", orderHandler.RemoveItem)
		orderGroup.Delete("/:id<guid>/items", orderHandler.ClearItems)
	}

	protectedOrderGroup := orderGroup.Group(
		"/",
		middleware.RequireAuth(),
		middleware.RequireTokenType(token.AccessTokenType),
	)
	{
		protectedOrderGroup.Post("/:id<guid>/checkout", orderHandler.Checkout)
		protectedOrderGroup.Post("/:id<guid>/cancel", orderHandler.CancelOrder)
	}
}
