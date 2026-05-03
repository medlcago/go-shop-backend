package core

import (
	"go-shop-backend/config"
	"go-shop-backend/pkg/notification"
)

func NewNotificationRegistry(cfg *config.Config) notification.SenderRegistry {
	emailSender := notification.NewEmailSender(
		cfg.Email.Host,
		cfg.Email.Port,
		cfg.Email.Username,
		cfg.Email.Password,
		cfg.Email.From,
	)

	notificationRegistry := notification.MapRegistry{
		notification.ChannelEmail: emailSender,
	}

	return notificationRegistry
}
