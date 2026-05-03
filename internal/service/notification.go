package service

import (
	"context"
	"fmt"
	"go-shop-backend/pkg/apperror"
	"go-shop-backend/pkg/notification"
	tmpl "go-shop-backend/pkg/template"
)

type notificationService struct {
	registry        notification.SenderRegistry
	templateManager tmpl.Manager
}

func NewNotificationService(
	registry notification.SenderRegistry,
	templateManager tmpl.Manager,
) *notificationService {
	return &notificationService{
		registry:        registry,
		templateManager: templateManager,
	}
}

func (n *notificationService) SendEmailConfirmationCode(
	ctx context.Context,
	to string,
	code string,
	channel notification.Channel,
) error {
	const op = "notificationService.SendEmailConfirmationCode"

	if to == "" {
		return apperror.Wrap(op, fmt.Errorf("to is empty"))
	}

	sender, ok := n.registry.For(channel)
	if !ok {
		return apperror.Wrap(op, fmt.Errorf("no sender for channel %q", channel))
	}

	data := map[string]string{
		"Code": code,
	}

	htmlBody, err := n.templateManager.Render("email_confirmation_code.gohtml", data)
	if err != nil {
		return apperror.Wrap(op, err)
	}

	textBody := fmt.Sprintf("Код подтверждения адреса электронной почты: %s", code)
	subject := fmt.Sprintf("Ваш код - %s", code)

	nt := notification.Notification{
		To:       to,
		Subject:  subject,
		HTMLBody: htmlBody,
		TextBody: textBody,
	}

	if err := sender.Send(ctx, nt); err != nil {
		return apperror.Wrap(op, err)
	}

	return nil
}
