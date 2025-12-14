package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type CorsConfig struct {
	AllowMethods        []string `env:"CORS_ALLOW_METHODS" mapstructure:"cors_allow_methods" validate:"required,min=1"`
	AllowOrigins        []string `env:"CORS_ALLOW_ORIGINS" mapstructure:"cors_allow_origins" validate:"required,min=1"`
	AllowHeaders        []string `env:"CORS_ALLOW_HEADERS" mapstructure:"cors_allow_headers" validate:"required,min=1"`
	ExposeHeaders       []string `env:"CORS_EXPOSE_HEADERS" mapstructure:"cors_expose_headers" validate:"required,min=1"`
	AllowCredentials    bool     `env:"CORS_ALLOW_CREDENTIALS" mapstructure:"cors_allow_credentials"`
	MaxAge              int      `env:"CORS_ALLOW_MAX_AGE" mapstructure:"cors_max_age"`
	AllowPrivateNetwork bool     `env:"CORS_ALLOW_PRIVATE_NETWORK" mapstructure:"cors_allow_private_network"`
}

type Config struct {
	Environment string `env:"ENVIRONMENT" mapstructure:"environment" validate:"required,oneof=production development"`
	HttpPort    int    `env:"HTTP_PORT" mapstructure:"http_port" validate:"required"`
	AuthSecret  string `env:"AUTH_SECRET" mapstructure:"auth_secret" validate:"required"`
	DatabaseURI string `env:"APP_+DATABASE_URI" mapstructure:"database_uri" validate:"required"`

	ServerReadTimeout     time.Duration `env:"SERVER_READ_TIMEOUT" mapstructure:"server_read_timeout" validate:"required"`
	ServerWriteTimeout    time.Duration `env:"SERVER_WRITE_TIMEOUT" mapstructure:"server_write_timeout" validate:"required"`
	ServerIdleTimeout     time.Duration `env:"SERVER_IDLE_TIMEOUT" mapstructure:"server_idle_timeout" validate:"required"`
	ServerShutdownTimeout time.Duration `env:"SERVER_SHUTDOWN_TIMEOUT" mapstructure:"server_shutdown_timeout" validate:"required"`

	CorsConfig `mapstructure:",squash"`
}

func (cfg *Config) Validate() error {
	v := validator.New()
	return v.Struct(cfg)
}

func Load() (*Config, error) {
	v := viper.New()

	v.AddConfigPath(".")
	v.SetConfigName(".env")
	v.SetConfigType("env")

	if err := v.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return nil, fmt.Errorf("error reading .env file: %w", err)
		}
	}

	v.SetEnvPrefix("APP")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal schema: %w", err)
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
