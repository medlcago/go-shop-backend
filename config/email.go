package config

import "time"

type Email struct {
	Host     string `env:"HOST" env-required:"true"`
	Port     int    `env:"PORT" env-required:"true"`
	Username string `env:"USERNAME" env-required:"true"`
	Password string `env:"PASSWORD" env-required:"true"`
	From     string `env:"FROM" env-required:"true"`

	ConfirmationCodeLength int           `env:"CONFIRMATION_CODE_LENGTH" env-required:"true"`
	ConfirmationCodeTTL    time.Duration `env:"CONFIRMATION_CODE_TTL" env-required:"true"`
}
