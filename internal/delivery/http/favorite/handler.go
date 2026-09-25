package favorite

import (
	"go-shop-backend/internal/service"
	"go-shop-backend/pkg/apperror"
	"go-shop-backend/pkg/middleware"
	"go-shop-backend/pkg/response"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type Handler struct {
	favoriteService service.FavoriteService
}

func NewHandler(favoriteService service.FavoriteService) *Handler {
	return &Handler{
		favoriteService: favoriteService,
	}
}

// AddProduct godoc
//
//	@Summary		Add product
//	@Description	Add product to favorites
//	@Tags			Favorites
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			productID	path		string	true	"Product ID"	Format(uuid)
//	@Success		200			{object}	response.Response[response.PaginatedResponse[[]go-shop-backend_internal_dto.FavoriteProduct]]
//	@Failure		404			{object}	response.Response[any]
//	@Failure		500			{object}	response.Response[any]
//	@Router			/favorites/{productID} [put]
func (h *Handler) AddProduct(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.UserID == nil {
		return apperror.ErrInvalidCredentials
	}

	productID := uuid.MustParse(ctx.Params("productID"))

	resp, total, err := h.favoriteService.AddProduct(ctx.Context(), *userCtx.UserID, productID)
	if err != nil {
		return err
	}

	return response.PaginatedJSON(ctx, fiber.StatusOK, resp, total)
}

// RemoveProduct godoc
//
//	@Summary		Remove product
//	@Description	Remove product from favorites
//	@Tags			Favorites
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			productID	path		string	true	"Product ID"	Format(uuid)
//	@Success		200			{object}	response.Response[response.PaginatedResponse[[]go-shop-backend_internal_dto.FavoriteProduct]]
//	@Failure		404			{object}	response.Response[any]
//	@Failure		500			{object}	response.Response[any]
//	@Router			/favorites/{productID} [delete]
func (h *Handler) RemoveProduct(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.UserID == nil {
		return apperror.ErrInvalidCredentials
	}

	productID := uuid.MustParse(ctx.Params("productID"))

	resp, total, err := h.favoriteService.RemoveProduct(ctx.Context(), *userCtx.UserID, productID)
	if err != nil {
		return err
	}

	return response.PaginatedJSON(ctx, fiber.StatusOK, resp, total)
}

// List godoc
//
//	@Summary		List
//	@Description	Favorites list
//	@Tags			Favorites
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.Response[response.PaginatedResponse[[]go-shop-backend_internal_dto.FavoriteProduct]]
//	@Failure		500	{object}	response.Response[any]
//	@Router			/favorites [get]
func (h *Handler) List(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.UserID == nil {
		return apperror.ErrInvalidCredentials
	}

	resp, total, err := h.favoriteService.List(ctx.Context(), *userCtx.UserID)
	if err != nil {
		return err
	}

	return response.PaginatedJSON(ctx, fiber.StatusOK, resp, total)
}

// IsFavorite godoc
//
//	@Summary		Is favorite product
//	@Description	Check if the product is in the favorites list
//	@Tags			Favorites
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			productID	path		string	true	"Product ID"	Format(uuid)
//	@Success		200			{object}	response.Response[go-shop-backend_internal_dto.FavoriteStatusResponse]
//	@Failure		500			{object}	response.Response[any]
//	@Router			/favorites/{productID} [get]
func (h *Handler) IsFavorite(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.UserID == nil {
		return apperror.ErrInvalidCredentials
	}

	productID := uuid.MustParse(ctx.Params("productID"))

	resp, err := h.favoriteService.IsFavorite(ctx.Context(), *userCtx.UserID, productID)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, resp)
}
