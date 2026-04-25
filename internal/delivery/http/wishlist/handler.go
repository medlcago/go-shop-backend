package wishlist

import (
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/service"
	"go-shop-backend/pkg/apperror"
	"go-shop-backend/pkg/middleware"
	"go-shop-backend/pkg/response"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type Handler struct {
	wishlistService service.WishlistService
}

func NewHandler(wishlistService service.WishlistService) *Handler {
	return &Handler{
		wishlistService: wishlistService,
	}
}

func (h *Handler) CreateWishlist(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.UserID == nil {
		return apperror.ErrInvalidCredentials
	}

	var req dto.CreateWishlistRequest
	if err := ctx.Bind().JSON(&req); err != nil {
		return err
	}

	resp, err := h.wishlistService.CreateWishlist(ctx, *userCtx.UserID, req)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusCreated, resp)
}

func (h *Handler) GetWishlist(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.UserID == nil {
		return apperror.ErrInvalidCredentials
	}

	wishlistID := uuid.MustParse(ctx.Params("wishlistID"))

	resp, err := h.wishlistService.GetWishlist(ctx, *userCtx.UserID, wishlistID)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, resp)
}

func (h *Handler) GetWishlists(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.UserID == nil {
		return apperror.ErrInvalidCredentials
	}

	var req dto.ListWishlistRequest
	if err := ctx.Bind().Query(&req); err != nil {
		return err
	}

	resp, total, err := h.wishlistService.GetWishlists(ctx, *userCtx.UserID, req)
	if err != nil {
		return err
	}

	return response.PaginatedJSON(ctx, fiber.StatusOK, resp, total)
}

func (h *Handler) GetSharedWishlist(ctx fiber.Ctx) error {
	token := ctx.Params("token")

	resp, err := h.wishlistService.GetSharedWishlist(ctx, token)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, resp)
}

func (h *Handler) RegenerateShareToken(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.UserID == nil {
		return apperror.ErrInvalidCredentials
	}

	wishlistID := uuid.MustParse(ctx.Params("wishlistID"))

	resp, err := h.wishlistService.RegenerateShareToken(ctx, *userCtx.UserID, wishlistID)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, resp)
}

func (h *Handler) UpdateWishlist(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.UserID == nil {
		return apperror.ErrInvalidCredentials
	}

	wishlistID := uuid.MustParse(ctx.Params("wishlistID"))

	var req dto.UpdateWishlistRequest
	if err := ctx.Bind().JSON(&req); err != nil {
		return err
	}

	resp, err := h.wishlistService.UpdateWishlist(ctx, *userCtx.UserID, wishlistID, req)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, resp)
}

func (h *Handler) AddItem(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.UserID == nil {
		return apperror.ErrInvalidCredentials
	}

	var req dto.AddWishlistItemRequest
	if err := ctx.Bind().JSON(&req); err != nil {
		return err
	}

	wishlistID := uuid.MustParse(ctx.Params("wishlistID"))

	resp, err := h.wishlistService.AddItem(ctx, *userCtx.UserID, wishlistID, req)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, resp)
}

func (h *Handler) UpdateItem(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.UserID == nil {
		return apperror.ErrInvalidCredentials
	}

	wishlistID := uuid.MustParse(ctx.Params("wishlistID"))
	itemID := uuid.MustParse(ctx.Params("itemID"))

	var req dto.UpdateWishlistItemRequest
	if err := ctx.Bind().JSON(&req); err != nil {
		return err
	}

	resp, err := h.wishlistService.UpdateItem(ctx, *userCtx.UserID, wishlistID, itemID, req)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, resp)
}

func (h *Handler) RemoveItem(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.UserID == nil {
		return apperror.ErrInvalidCredentials
	}

	wishlistID := uuid.MustParse(ctx.Params("wishlistID"))
	itemID := uuid.MustParse(ctx.Params("itemID"))

	resp, err := h.wishlistService.RemoveItem(ctx, *userCtx.UserID, wishlistID, itemID)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, resp)
}
