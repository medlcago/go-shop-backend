package config

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	AppName                 string        `env:"APP_NAME" env-required:"true"`
	Environment             string        `env:"ENVIRONMENT" env-required:"true" validate:"required,oneof=production development"`
	AuthSecret              string        `env:"AUTH_SECRET" env-required:"true"`
	AccessTokenExpiredTime  time.Duration `env:"AUTH_ACCESS_TOKEN_EXPIRED_TIME" env-required:"true"`
	RefreshTokenExpiredTime time.Duration `env:"AUTH_REFRESH_TOKEN_EXPIRED_TIME" env-required:"true"`
	PartialTokenExpiredTime time.Duration `env:"AUTH_PARTIAL_TOKEN_EXPIRED_TIME" env-required:"true"`
	MasterKey               string        `env:"MASTER_KEY" env-required:"true"`

	HttpServer HttpServer `env-prefix:"HTTP_"`
	Worker     Worker     `env-prefix:"WORKER_"`
	Cors       Cors       `env-prefix:"CORS_"`
	Database   Database   `env-prefix:"DB_"`
	Minio      Minio      `env-prefix:"MINIO_"`
	Upload     Upload     `env-prefix:"UPLOAD_"`
	Yookassa   Yookassa   `env-prefix:"YOOKASSA_"`
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
