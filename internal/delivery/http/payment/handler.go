package payment

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
	paymentService service.PaymentService
}

func NewHandler(paymentService service.PaymentService) *Handler {
	return &Handler{
		paymentService: paymentService,
	}
}

// CreatePayment godoc
//
//	@Summary		Create payment
//	@Description	Create a new payment for an order
//	@Tags			Payments
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.CreatePaymentRequest	true	"Request to Create a Payment"
//	@Success		201		{object}	response.Response[dto.PaymentResponse]
//	@Failure		400		{object}	response.Response[any]
//	@Failure		401		{object}	response.Response[any]
//	@Failure		403		{object}	response.Response[any]
//	@Failure		404		{object}	response.Response[any]
//	@Failure		409		{object}	response.Response[any]
//	@Failure		500		{object}	response.Response[any]
//	@Router			/payments [post]
func (h *Handler) CreatePayment(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.UserID == nil {
		return apperror.ErrInvalidCredentials
	}

	var req dto.CreatePaymentRequest
	if err := ctx.Bind().JSON(&req); err != nil {
		return err
	}

	resp, err := h.paymentService.CreatePayment(ctx.Context(), *userCtx.UserID, req)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusCreated, resp)

}

// HandleYookassaWebhook godoc
//
//	@Summary		Yookassa webhook callback
//	@Description	Process incoming webhook events from Yookassa payment gateway
//	@Tags			Payments
//	@Accept			json
//	@Produce		plain
//	@Param			request	body	object	true	"Yookassa Event Payload"
//	@Success		200		"OK"
//	@Failure		400		{object}	response.Response[any]
//	@Failure		500		{object}	response.Response[any]
//	@Router			/payments/webhook/yookassa [post]
func (h *Handler) HandleYookassaWebhook(ctx fiber.Ctx) error {
	err := h.paymentService.HandleWebhook(ctx.Context(), ctx.Body())
	if err != nil {
		return err
	}

	return ctx.SendStatus(fiber.StatusOK)
}

// GetUserPaymentMethods godoc
//
//	@Summary		Get user payment methods
//	@Description	Retrieve saved payment methods for current authenticated user
//	@Tags			Payments
//	@Security		BearerAuth
//	@Produce		json
//	@Success		200	{object}	response.PaginatedResponse[[]dto.UserPaymentMethodResponse]
//	@Failure		401	{object}	response.Response[any]
//	@Failure		500	{object}	response.Response[any]
//	@Router			/payments/user-methods [get]
func (h *Handler) GetUserPaymentMethods(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.UserID == nil {
		return apperror.ErrInvalidCredentials
	}

	resp, total, err := h.paymentService.GetUserPaymentMethods(ctx.Context(), *userCtx.UserID)
	if err != nil {
		return err
	}

	return response.PaginatedJSON(ctx, fiber.StatusOK, resp, total)
}

// SetDefaultPaymentMethod godoc
//
//	@Summary		Set default payment method
//	@Description	Set a specific payment method as default for the user
//	@Tags			Payments
//	@Security		BearerAuth
//	@Param			id	path	string	true	"Payment Method UUID"	format(uuid)
//	@Success		200	"OK"
//	@Failure		400	{object}	response.Response[any]
//	@Failure		401	{object}	response.Response[any]
//	@Failure		404	{object}	response.Response[any]
//	@Failure		500	{object}	response.Response[any]
//	@Router			/payments/user-methods/{id}/default [put]
func (h *Handler) SetDefaultPaymentMethod(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.UserID == nil {
		return apperror.ErrInvalidCredentials
	}

	id := uuid.MustParse(ctx.Params("id"))

	if err := h.paymentService.SetDefaultPaymentMethod(ctx.Context(), id, *userCtx.UserID); err != nil {
		return err
	}

	return ctx.SendStatus(fiber.StatusOK)
}

// DeletePaymentMethod godoc
//
//	@Summary		Delete payment method
//	@Description	Remove a saved payment method by its ID
//	@Tags			Payments
//	@Security		BearerAuth
//	@Param			id	path	string	true	"Payment Method UUID"	format(uuid)
//	@Success		204	"No Content"
//	@Failure		400	{object}	response.Response[any]
//	@Failure		401	{object}	response.Response[any]
//	@Failure		404	{object}	response.Response[any]
//	@Failure		500	{object}	response.Response[any]
//	@Router			/payments/methods/{id} [delete]
func (h *Handler) DeletePaymentMethod(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.UserID == nil {
		return apperror.ErrInvalidCredentials
	}

	id := uuid.MustParse(ctx.Params("id"))

	if err := h.paymentService.DeletePaymentMethod(ctx.Context(), id, *userCtx.UserID); err != nil {
		return err
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}
