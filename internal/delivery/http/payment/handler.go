package payment

import (
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/service"
	"go-shop-backend/pkg/apperror"
	"go-shop-backend/pkg/middleware"
	"go-shop-backend/pkg/response"

	"github.com/gofiber/fiber/v3"
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

	resp, err := h.paymentService.CreatePayment(ctx, *userCtx.UserID, req)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusCreated, resp)

}

func (h *Handler) HandleYookassaWebhook(ctx fiber.Ctx) error {
	err := h.paymentService.HandleWebhook(ctx, ctx.Body())
	if err != nil {
		return err
	}

	return ctx.SendStatus(fiber.StatusOK)
}
