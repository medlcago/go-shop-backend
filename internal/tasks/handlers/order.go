package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go-shop-backend/internal/service"
	"go-shop-backend/internal/tasks"
	"go-shop-backend/pkg/apperrors"
	"go-shop-backend/pkg/logger"
	"log/slog"
	"slices"

	"github.com/hibiken/asynq"
)

type OrderTaskHandler struct {
	orderService service.OrderService
	logger       *slog.Logger

	nonRetryableErrors []error
}

func NewOrderTaskHandler(orderService service.OrderService, logger *slog.Logger) *OrderTaskHandler {
	logger = logger.With(
		slog.String("handler", "OrderTaskHandler"),
	)

	nonRetryableErrors := []error{
		apperrors.ErrForbidden,
		apperrors.ErrOrderNotFound,
		apperrors.ErrInvalidOrderStatus,
	}

	return &OrderTaskHandler{
		orderService:       orderService,
		logger:             logger,
		nonRetryableErrors: nonRetryableErrors,
	}
}

func (h *OrderTaskHandler) CancelOrder(ctx context.Context, task *asynq.Task) error {
	var payload tasks.CancelOrderPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	if err := h.orderService.CancelOrder(ctx, payload.UserID, payload.OrderID); err != nil {
		h.logger.Error(
			"failed to cancel order",
			slog.String("order_id", payload.OrderID.String()),
			logger.Err(err),
		)

		if slices.ContainsFunc(h.nonRetryableErrors, func(e error) bool {
			return errors.Is(err, e)
		}) {
			return fmt.Errorf("orderService.CancelOrder failed: %v: %w", err, asynq.SkipRetry)
		}

		return err
	}

	return nil
}
