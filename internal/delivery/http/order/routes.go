package order

import (
	"go-shop-backend/pkg/middleware"

	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(r fiber.Router, orderHandler *Handler, auth middleware.Auth) {
	orderGroup := r.Group("/orders")
	orderGroup.Use(auth.Handle())
	{
		orderGroup.Post("/", orderHandler.CreateOrder)
		orderGroup.Get("/", orderHandler.GetOrders)
		orderGroup.Get("/:id<guid>", orderHandler.GetOrder)
		orderGroup.Post("/:id<guid>/items", orderHandler.AddItem)
		orderGroup.Delete("/:id<guid>/items/:item_id<guid>", orderHandler.DeleteItem)
	}
}
