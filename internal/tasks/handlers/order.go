package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/tasks"
	"go-shop-backend/pkg/apperror"
	"go-shop-backend/pkg/logger"
	"log/slog"
	"slices"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type OrderStatusUpdater interface {
	UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status models.OrderStatus) error
}

type OrderTaskHandler struct {
	orderStatusUpdater OrderStatusUpdater
	logger             *slog.Logger

	nonRetryableErrors []error
}

func NewOrderTaskHandler(orderStatusUpdater OrderStatusUpdater, logger *slog.Logger) *OrderTaskHandler {
	logger = logger.With(
		slog.String("handler", "OrderTaskHandler"),
	)

	nonRetryableErrors := []error{
		apperror.ErrForbidden,
		apperror.ErrOrderNotFound,
		apperror.ErrInvalidOrderStatus,
	}

	return &OrderTaskHandler{
		orderStatusUpdater: orderStatusUpdater,
		logger:             logger,
		nonRetryableErrors: nonRetryableErrors,
	}
}

func (h *OrderTaskHandler) CancelOrder(ctx context.Context, task *asynq.Task) error {
	var payload tasks.CancelOrderPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	if err := h.orderStatusUpdater.UpdateOrderStatus(ctx, payload.OrderID, models.OrderStatusCanceled); err != nil {
		h.logger.Error(
			"failed to cancel order",
			slog.String("order_id", payload.OrderID.String()),
			logger.Err(err),
		)

		if slices.ContainsFunc(h.nonRetryableErrors, func(e error) bool {
			return errors.Is(err, e)
		}) {
			return fmt.Errorf("orderService.UpdateOrderStatus failed: %v: %w", err, asynq.SkipRetry)
		}

		return err
	}

	return nil
}
