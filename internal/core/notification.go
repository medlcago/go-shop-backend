package core

import (
	"go-shop-backend/config"
	"go-shop-backend/pkg/notification"
)

func NewNotificationRegistry(cfg *config.Config) notification.Registry {
	emailSender := notification.NewEmailSender(
		cfg.Email.Host,
		cfg.Email.Port,
		cfg.Email.Username,
		cfg.Email.Password,
		cfg.Email.From,
	)

	registry := notification.NewRegistry()
	registry.Register(
		notification.ChannelEmail,
		emailSender,
	)

	return registry
}
