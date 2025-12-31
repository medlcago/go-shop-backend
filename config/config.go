package config

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Environment string `env:"ENVIRONMENT" env-required:"true" validate:"required,oneof=production development"`
	AuthSecret  string `env:"AUTH_SECRET" env-required:"true"`

	HttpServer HttpServer `env-prefix:"HTTP_"`
	Cors       Cors       `env-prefix:"CORS_"`
	Database   Database   `env-prefix:"DB_"`
}

func (cfg *Config) Validate() error {
	v := validator.New()
	return v.Struct(cfg)
}

func Load() (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadConfig(".env", &cfg); err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	return &cfg, nil
}

func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		panic(fmt.Errorf("failed to load config: %w", err))
	}

	return cfg
}
