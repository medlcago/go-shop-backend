package config

import "time"

type HTTPClient struct {
	RetryMax     int           `env:"RETRY_MAX" env-default:"5"`
	RetryWaitMin time.Duration `env:"RETRY_WAIT_MIN" env-default:"300ms"`
	RetryWaitMax time.Duration `env:"RETRY_WAIT_MAX" env-default:"2s"`
	Timeout      time.Duration `env:"TIMEOUT" env-default:"10s"`
}
