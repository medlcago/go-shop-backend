package address

import (
	"go-shop-backend/pkg/middleware"
	"go-shop-backend/pkg/token"

	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(r fiber.Router, addressHandler *Handler) {
	addressGroup := r.Group(
		"/addresses",
		middleware.RequireAuth(),
		middleware.RequireTokenType(token.AccessTokenType),
	)
	{
		addressGroup.Post("/", addressHandler.CreateAddress)
		addressGroup.Get("/", addressHandler.ListAddresses)
		addressGroup.Get("/:addressID<guid>", addressHandler.GetAddress)
		addressGroup.Put("/:addressID<guid>", addressHandler.UpdateAddress)
		addressGroup.Delete("/:addressID<guid>", addressHandler.DeleteAddress)
		addressGroup.Put("/:addressID<guid>/default", addressHandler.SetDefaultAddress)
	}
}
