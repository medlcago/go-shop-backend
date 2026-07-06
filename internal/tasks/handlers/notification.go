package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"go-shop-backend/internal/tasks"
	"go-shop-backend/pkg/logger"
	"go-shop-backend/pkg/notification"
	"log/slog"

	"github.com/hibiken/asynq"
)

type Notification interface {
	SendEmailConfirmationCode(ctx context.Context, to string, code string, channel notification.Channel) error
}

type NotificationTaskHandler struct {
	notification Notification
	logger       *slog.Logger
}

func NewNotificationTaskHandler(notification Notification, logger *slog.Logger) *NotificationTaskHandler {
	return &NotificationTaskHandler{
		notification: notification,
		logger:       logger,
	}
}

func (h *NotificationTaskHandler) SendEmailConfirmationCode(ctx context.Context, task *asynq.Task) error {
	var payload tasks.SendEmailConfirmationCodePayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	if err := h.notification.SendEmailConfirmationCode(ctx, payload.Email, payload.Code, notification.ChannelEmail); err != nil {
		h.logger.Error(
			"failed to send email confirmation code",
			logger.Err(err),
		)

		return err
	}

	return nil
}
