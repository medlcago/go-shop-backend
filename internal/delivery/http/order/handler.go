package order

import (
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/service"
	"go-shop-backend/pkg/apperrors"
	"go-shop-backend/pkg/middleware"
	"go-shop-backend/pkg/response"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type Handler struct {
	orderService service.OrderService
}

func NewHandler(orderService service.OrderService) *Handler {
	return &Handler{
		orderService: orderService,
	}
}

// CreateOrder godoc
//
//	@Summary		Create new order
//	@Description	Create new order
//	@Tags			Orders
//	@Security		BearerAuth
//	@Produce		json
//	@Param			X-Session-ID	header		string	true	"Session ID"	Format(uuid)
//	@Success		201				{object}	response.Response[dto.OrderResponse]
//	@Failure		401				{object}	response.Response[any]
//	@Failure		500				{object}	response.Response[any]
//	@Router			/orders [post]
func (h *Handler) CreateOrder(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.SessionID == nil {
		return apperrors.ErrInvalidCredentials
	}

	resp, err := h.orderService.CreateOrder(ctx, userCtx.UserID, *userCtx.SessionID)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusCreated, resp)

}

// GetOrder godoc
//
//	@Summary		Get order by ID
//	@Description	Get single order by ID
//	@Tags			Orders
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id				path		string	true	"Order ID"		Format(uuid)
//	@Param			X-Session-ID	header		string	true	"Session ID"	Format(uuid)
//	@Success		200				{object}	response.Response[dto.OrderResponse]
//	@Failure		401				{object}	response.Response[any]
//	@Failure		403				{object}	response.Response[any]
//	@Failure		500				{object}	response.Response[any]
//	@Router			/orders/{id} [get]
func (h *Handler) GetOrder(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.SessionID == nil {
		return apperrors.ErrInvalidCredentials
	}

	orderID := uuid.MustParse(ctx.Params("id"))

	resp, err := h.orderService.GetOrder(ctx, userCtx.UserID, *userCtx.SessionID, orderID)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, resp)
}

// GetOrders godoc
//
//	@Summary		Get list of orders
//	@Description	Get paginated list of orders
//	@Tags			Orders
//	@Security		BearerAuth
//	@Produce		json
//	@Param			limit			query		int		false	"Maximum number of items to return"	minimum(1)	default(50)
//	@Param			offset			query		int		false	"Number of items to skip"			minimum(0)	default(0)
//	@Param			X-Session-ID	header		string	true	"Session ID"						Format(uuid)
//	@Param			status			query		string	false	"Order status"						Enums(draft, pending, paid, canceled, completed)
//	@Success		200				{object}	response.PaginatedResponse[[]dto.OrderResponse]
//	@Failure		401				{object}	response.Response[any]
//	@Failure		500				{object}	response.Response[any]
//	@Router			/orders [get]
func (h *Handler) GetOrders(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.SessionID == nil {
		return apperrors.ErrInvalidCredentials
	}

	var req dto.ListOrderRequest
	if err := ctx.Bind().Query(&req); err != nil {
		return err
	}

	resp, total, err := h.orderService.GetOrders(ctx, userCtx.UserID, *userCtx.SessionID, req)
	if err != nil {
		return err
	}

	return response.PaginatedJSON(ctx, fiber.StatusOK, resp, total)
}

// AddItem godoc
//
//	@Summary		Add item to order
//	@Description	Add product to existing order
//	@Tags			Orders
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			id				path		string					true	"Order ID"		Format(uuid)
//	@Param			X-Session-ID	header		string					true	"Session ID"	Format(uuid)
//	@Param			request			body		dto.AddOrderItemRequest	true	"Add item request"
//	@Success		200				{object}	response.Response[dto.OrderResponse]
//	@Failure		400				{object}	response.Response[any]
//	@Failure		401				{object}	response.Response[any]
//	@Failure		403				{object}	response.Response[any]
//	@Failure		404				{object}	response.Response[any]
//	@Failure		409				{object}	response.Response[any]
//	@Failure		500				{object}	response.Response[any]
//	@Router			/orders/{id}/items [post]
func (h *Handler) AddItem(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.SessionID == nil {
		return apperrors.ErrInvalidCredentials
	}

	var req dto.AddOrderItemRequest
	if err := ctx.Bind().JSON(&req); err != nil {
		return err
	}

	orderID := uuid.MustParse(ctx.Params("id"))

	resp, err := h.orderService.AddItem(ctx, userCtx.UserID, *userCtx.SessionID, orderID, req)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, resp)
}

// DeleteItem godoc
//
//	@Summary		Delete item from order
//	@Description	Delete item from order
//	@Tags			Orders
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id				path		string	true	"Order ID"		Format(uuid)
//	@Param			item_id			path		string	true	"Item ID"		Format(uuid)
//	@Param			X-Session-ID	header		string	true	"Session ID"	Format(uuid)
//	@Success		200				{object}	response.Response[dto.OrderResponse]
//	@Failure		401				{object}	response.Response[any]
//	@Failure		403				{object}	response.Response[any]
//	@Failure		500				{object}	response.Response[any]
//	@Router			/orders/{id}/items/{item_id} [delete]
func (h *Handler) DeleteItem(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.SessionID == nil {
		return apperrors.ErrInvalidCredentials
	}

	orderID := uuid.MustParse(ctx.Params("id"))
	itemID := uuid.MustParse(ctx.Params("item_id"))

	resp, err := h.orderService.DeleteItem(ctx, userCtx.UserID, *userCtx.SessionID, orderID, itemID)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, resp)
}
