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

func (h *Handler) CreateAddress(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.UserID == nil {
		return apperror.ErrInvalidCredentials
	}

	var req dto.CreateAddressRequest
	if err := ctx.Bind().JSON(&req); err != nil {
		return err
	}

	resp, err := h.addressService.CreateAddress(ctx, *userCtx.UserID, req)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusCreated, resp)
}

func (h *Handler) ListAddresses(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.UserID == nil {
		return apperror.ErrInvalidCredentials
	}

	resp, err := h.addressService.ListAddresses(ctx, *userCtx.UserID)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, resp)
}

func (h *Handler) GetAddress(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.UserID == nil {
		return apperror.ErrInvalidCredentials
	}

	addressID := uuid.MustParse(ctx.Params("addressID"))

	resp, err := h.addressService.GetAddress(ctx, addressID, *userCtx.UserID)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, resp)
}

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

	resp, err := h.addressService.UpdateAddress(ctx, addressID, *userCtx.UserID, req)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, resp)
}

func (h *Handler) DeleteAddress(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.UserID == nil {
		return apperror.ErrInvalidCredentials
	}

	addressID := uuid.MustParse(ctx.Params("addressID"))

	err := h.addressService.DeleteAddress(ctx, addressID, *userCtx.UserID)
	if err != nil {
		return err
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) SetDefaultAddress(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.UserID == nil {
		return apperror.ErrInvalidCredentials
	}

	addressID := uuid.MustParse(ctx.Params("addressID"))

	err := h.addressService.SetDefault(ctx, addressID, *userCtx.UserID)
	if err != nil {
		return err
	}

	return ctx.SendStatus(fiber.StatusOK)
}
