package address

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
	addressService service.AddressService
}

func NewHandler(addressService service.AddressService) *Handler {
	return &Handler{
		addressService: addressService,
	}
}

// CreateAddress godoc
//
//	@Summary		Create address
//	@Description	Create a new delivery address for the authenticated user
//	@Tags			Addresses
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.CreateAddressRequest	true	"Request body to create address"
//	@Success		201		{object}	response.Response[dto.AddressResponse]
//	@Failure		400		{object}	response.Response[any]
//	@Failure		401		{object}	response.Response[any]
//	@Failure		500		{object}	response.Response[any]
//	@Router			/addresses [post]
func (h *Handler) CreateAddress(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.UserID == nil {
		return apperror.ErrInvalidCredentials
	}

	var req dto.CreateAddressRequest
	if err := ctx.Bind().JSON(&req); err != nil {
		return err
	}

	resp, err := h.addressService.CreateAddress(ctx.Context(), *userCtx.UserID, req)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusCreated, resp)
}

// ListAddresses godoc
//
//	@Summary		List addresses
//	@Description	Get all saved addresses for the authenticated user
//	@Tags			Addresses
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.Response[[]dto.AddressResponse]
//	@Failure		401	{object}	response.Response[any]
//	@Failure		500	{object}	response.Response[any]
//	@Router			/addresses [get]
func (h *Handler) ListAddresses(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.UserID == nil {
		return apperror.ErrInvalidCredentials
	}

	resp, err := h.addressService.ListAddresses(ctx.Context(), *userCtx.UserID)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, resp)
}

// GetAddress godoc
//
//	@Summary		Get address
//	@Description	Get address details by ID
//	@Tags			Addresses
//	@Produce		json
//	@Security		BearerAuth
//	@Param			addressID	path		string	true	"Address UUID"	format(uuid)
//	@Success		200			{object}	response.Response[dto.AddressResponse]
//	@Failure		400			{object}	response.Response[any]
//	@Failure		401			{object}	response.Response[any]
//	@Failure		404			{object}	response.Response[any]
//	@Failure		500			{object}	response.Response[any]
//	@Router			/addresses/{addressID} [get]
func (h *Handler) GetAddress(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.UserID == nil {
		return apperror.ErrInvalidCredentials
	}

	addressID := uuid.MustParse(ctx.Params("addressID"))

	resp, err := h.addressService.GetAddress(ctx.Context(), addressID, *userCtx.UserID)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, resp)
}

// UpdateAddress godoc
//
//	@Summary		Update address
//	@Description	Update address details by ID
//	@Tags			Addresses
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			addressID	path		string						true	"Address UUID"	format(uuid)
//	@Param			request		body		dto.UpdateAddressRequest	true	"Request body to update address"
//	@Success		200			{object}	response.Response[dto.AddressResponse]
//	@Failure		400			{object}	response.Response[any]
//	@Failure		401			{object}	response.Response[any]
//	@Failure		404			{object}	response.Response[any]
//	@Failure		500			{object}	response.Response[any]
//	@Router			/addresses/{addressID} [put]
func (h *Handler) UpdateAddress(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.UserID == nil {
		return apperror.ErrInvalidCredentials
	}

	addressID := uuid.MustParse(ctx.Params("addressID"))

	var req dto.UpdateAddressRequest
	if err := ctx.Bind().JSON(&req); err != nil {
		return err
	}

	resp, err := h.addressService.UpdateAddress(ctx.Context(), addressID, *userCtx.UserID, req)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, resp)
}

// DeleteAddress godoc
//
//	@Summary		Delete address
//	@Description	Delete address by ID
//	@Tags			Addresses
//	@Security		BearerAuth
//	@Param			addressID	path	string	true	"Address UUID"	format(uuid)
//	@Success		204			"No Content"
//	@Failure		400			{object}	response.Response[any]
//	@Failure		401			{object}	response.Response[any]
//	@Failure		404			{object}	response.Response[any]
//	@Failure		500			{object}	response.Response[any]
//	@Router			/addresses/{addressID} [delete]
func (h *Handler) DeleteAddress(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.UserID == nil {
		return apperror.ErrInvalidCredentials
	}

	addressID := uuid.MustParse(ctx.Params("addressID"))

	err := h.addressService.DeleteAddress(ctx.Context(), addressID, *userCtx.UserID)
	if err != nil {
		return err
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}

// SetDefaultAddress godoc
//
//	@Summary		Set default address
//	@Description	Set specific address as default for user
//	@Tags			Addresses
//	@Security		BearerAuth
//	@Param			addressID	path	string	true	"Address UUID"	format(uuid)
//	@Success		200			"OK"
//	@Failure		400			{object}	response.Response[any]
//	@Failure		401			{object}	response.Response[any]
//	@Failure		404			{object}	response.Response[any]
//	@Failure		500			{object}	response.Response[any]
//	@Router			/addresses/{addressID}/default [put]
func (h *Handler) SetDefaultAddress(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.UserID == nil {
		return apperror.ErrInvalidCredentials
	}

	addressID := uuid.MustParse(ctx.Params("addressID"))

	err := h.addressService.SetDefault(ctx.Context(), addressID, *userCtx.UserID)
	if err != nil {
		return err
	}

	return ctx.SendStatus(fiber.StatusOK)
}
