package wishlist

import (
	"go-shop-backend/pkg/middleware"
	"go-shop-backend/pkg/token"

	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(r fiber.Router, wishlistHandler *Handler) {
	wishlistGroup := r.Group("/wishlists")
	{
		wishlistGroup.Get(
			"/shared/:token",
			wishlistHandler.GetSharedWishlist,
		)
	}

	protectedWishlistGroup := wishlistGroup.Group(
		"/",
		middleware.RequireAuth(),
		middleware.RequireTokenType(token.AccessTokenType),
	)
	{
		protectedWishlistGroup.Post(
			"/",
			wishlistHandler.CreateWishlist,
		)
		protectedWishlistGroup.Get(
			"/:wishlistID<guid>",
			wishlistHandler.GetWishlist,
		)
		protectedWishlistGroup.Get(
			"/",
			wishlistHandler.GetWishlists,
		)
		protectedWishlistGroup.Post(
			":wishlistID<guid>/regenerate-token",
			wishlistHandler.RegenerateShareToken,
		)
		protectedWishlistGroup.Patch(
			"/:wishlistID<guid>",
			wishlistHandler.UpdateWishlist,
		)
		protectedWishlistGroup.Post(
			"/:wishlistID<guid>/items",
			wishlistHandler.AddItem,
		)
		protectedWishlistGroup.Patch(
			"/:wishlistID<guid>/items/:itemID<guid>",
			wishlistHandler.UpdateItem,
		)
		protectedWishlistGroup.Delete(
			"/:wishlistID<guid>/items/:itemID<guid>",
			wishlistHandler.RemoveItem,
		)
	}
}
