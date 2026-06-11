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

// CreateWishlist godoc
//
//	@Summary		Create a new wishlist
//	@Description	Creates a new wishlist for the authenticated user
//	@Tags			Wishlists
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.CreateWishlistRequest	true	"Wishlist creation details"
//	@Success		201		{object}	response.Response[dto.WishlistResponse]
//	@Failure		400		{object}	response.Response[any]
//	@Failure		401		{object}	response.Response[any]
//	@Failure		500		{object}	response.Response[any]
//	@Router			/wishlists [post]
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

// GetWishlist godoc
//
//	@Summary		Get wishlist by ID
//	@Description	Retrieves a specific wishlist by its ID for the authenticated user
//	@Tags			Wishlists
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			wishlistID	path		string	true	"Wishlist ID"	Format(uuid)
//	@Success		200			{object}	response.Response[dto.WishlistResponse]
//	@Failure		400			{object}	response.Response[any]
//	@Failure		401			{object}	response.Response[any]
//	@Failure		403			{object}	response.Response[any]
//	@Failure		404			{object}	response.Response[any]
//	@Failure		500			{object}	response.Response[any]
//	@Router			/wishlists/{wishlistID} [get]
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

// GetWishlists godoc
//
//	@Summary		Get user's wishlists
//	@Description	Retrieves a paginated list of wishlists for the authenticated user
//	@Tags			Wishlists
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			limit	query		int	false	"Maximum number of items to return"	minimum(1)	default(50)
//	@Param			offset	query		int	false	"Number of items to skip"			minimum(0)	default(0)
//	@Success		200		{object}	response.PaginatedResponse[[]dto.WishlistResponse]
//	@Failure		400		{object}	response.Response[any]
//	@Failure		401		{object}	response.Response[any]
//	@Failure		500		{object}	response.Response[any]
//	@Router			/wishlists [get]
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

// GetSharedWishlist godoc
//
//	@Summary		Get shared wishlist by token
//	@Description	Retrieves a publicly shared wishlist using its share token (no authentication required)
//	@Tags			Wishlists
//	@Accept			json
//	@Produce		json
//	@Param			token	path		string	true	"Wishlist share token"
//	@Success		200		{object}	response.Response[dto.WishlistResponse]
//	@Failure		400		{object}	response.Response[any]
//	@Failure		404		{object}	response.Response[any]
//	@Failure		500		{object}	response.Response[any]
//	@Router			/wishlists/shared/{token} [get]
func (h *Handler) GetSharedWishlist(ctx fiber.Ctx) error {
	token := ctx.Params("token")

	resp, err := h.wishlistService.GetSharedWishlist(ctx, token)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, resp)
}

// RegenerateShareToken godoc
//
//	@Summary		Regenerate share token
//	@Description	Generates a new share token for a wishlist (invalidates the old one)
//	@Tags			Wishlists
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			wishlistID	path		string	true	"Wishlist ID"	Format(uuid)
//	@Success		200			{object}	response.Response[dto.WishlistShareTokenResponse]
//	@Failure		400			{object}	response.Response[any]
//	@Failure		401			{object}	response.Response[any]
//	@Failure		403			{object}	response.Response[any]
//	@Failure		404			{object}	response.Response[any]
//	@Failure		500			{object}	response.Response[any]
//	@Router			/wishlists/{wishlistID}/regenerate-token [post]
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

// UpdateWishlist godoc
//
//	@Summary		Update wishlist
//	@Description	Updates an existing wishlist
//	@Tags			Wishlists
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			wishlistID	path		string						true	"Wishlist ID"	Format(uuid)
//	@Param			request		body		dto.UpdateWishlistRequest	true	"Wishlist update details"
//	@Success		200			{object}	response.Response[dto.WishlistResponse]
//	@Failure		400			{object}	response.Response[any]
//	@Failure		401			{object}	response.Response[any]
//	@Failure		403			{object}	response.Response[any]
//	@Failure		404			{object}	response.Response[any]
//	@Failure		500			{object}	response.Response[any]
//	@Router			/wishlists/{wishlistID} [patch]
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

// AddItem godoc
//
//	@Summary		Add item to wishlist
//	@Description	Adds a new item (product) to the wishlist
//	@Tags			Wishlists
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			wishlistID	path		string						true	"Wishlist ID"	Format(uuid)
//	@Param			request		body		dto.AddWishlistItemRequest	true	"Item details"
//	@Success		200			{object}	response.Response[dto.WishlistResponse]
//	@Failure		400			{object}	response.Response[any]
//	@Failure		401			{object}	response.Response[any]
//	@Failure		403			{object}	response.Response[any]
//	@Failure		404			{object}	response.Response[any]
//	@Failure		409			{object}	response.Response[any]
//	@Failure		500			{object}	response.Response[any]
//	@Router			/wishlists/{wishlistID}/items [post]
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

// UpdateItem godoc
//
//	@Summary		Update wishlist item
//	@Description	Updates an existing item in a wishlist
//	@Tags			Wishlists
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			wishlistID	path		string							true	"Wishlist ID"	Format(uuid)
//	@Param			itemID		path		string							true	"Item ID"		Format(uuid)
//	@Param			request		body		dto.UpdateWishlistItemRequest	true	"Item update details"
//	@Success		200			{object}	response.Response[dto.WishlistResponse]
//	@Failure		400			{object}	response.Response[any]
//	@Failure		401			{object}	response.Response[any]
//	@Failure		403			{object}	response.Response[any]
//	@Failure		404			{object}	response.Response[any]
//	@Failure		500			{object}	response.Response[any]
//	@Router			/wishlists/{wishlistID}/items/{itemID} [patch]
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

// RemoveItem godoc
//
//	@Summary		Remove item from wishlist
//	@Description	Removes an item from a wishlist
//	@Tags			Wishlists
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			wishlistID	path		string	true	"Wishlist ID"	Format(uuid)
//	@Param			itemID		path		string	true	"Item ID"		Format(uuid)
//	@Success		200			{object}	response.Response[dto.WishlistResponse]
//	@Failure		400			{object}	response.Response[any]
//	@Failure		401			{object}	response.Response[any]
//	@Failure		403			{object}	response.Response[any]
//	@Failure		404			{object}	response.Response[any]
//	@Failure		500			{object}	response.Response[any]
//	@Router			/wishlists/{wishlistID}/items/{itemID} [delete]
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
