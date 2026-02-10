package cart

import (
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/service"
	"go-shop-backend/pkg/middleware"
	"go-shop-backend/pkg/response"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type Handler struct {
	cartService service.CartService
}

func NewHandler(cartService service.CartService) *Handler {
	return &Handler{
		cartService: cartService,
	}
}

// GetCart godoc
//
//	@Summary		Get cart
//	@Description	Get user's cart
//	@Tags			cart
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Response[dto.CartResponse]
//	@Failure		400	{object}	response.Response[any]
//	@Failure		500	{object}	response.Response[any]
//	@Router			/cart [get]
func (h *Handler) GetCart(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)

	cart, err := h.cartService.GetCart(ctx, userCtx.UserID, userCtx.SessionID)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, cart)
}

// AddItem godoc
//
//	@Summary		Add item
//	@Description	Add item to cart
//	@Tags			cart
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.AddItemRequest	true	"Item data"
//	@Success		200		{object}	response.Response[dto.CartResponse]
//	@Failure		400		{object}	response.Response[any]
//	@Failure		500		{object}	response.Response[any]
//	@Router			/cart/items [post]
func (h *Handler) AddItem(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)

	var req dto.AddItemRequest
	if err := ctx.Bind().JSON(&req); err != nil {
		return err
	}

	cart, err := h.cartService.AddItem(ctx, userCtx.UserID, userCtx.SessionID, req)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, cart)
}

// DeleteItem godoc
//
//	@Summary		Delete item
//	@Description	Delete item from cart by ID
//	@Tags			cart
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Item ID"	Format(uuid)
//	@Success		200	{object}	response.Response[dto.CartResponse]
//	@Failure		400	{object}	response.Response[any]
//	@Failure		500	{object}	response.Response[any]
//	@Router			/cart/items/{id} [delete]
func (h *Handler) DeleteItem(ctx fiber.Ctx) error {
	itemID := uuid.MustParse(ctx.Params("id"))
	userCtx := middleware.GetUserContext(ctx)

	cart, err := h.cartService.DeleteItem(ctx, userCtx.UserID, userCtx.SessionID, itemID)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, cart)
}
