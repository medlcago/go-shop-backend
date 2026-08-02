package httpclient

import "time"

const (
	defaultRetryMax     = 3
	defaultRetryWaitMin = 300 * time.Millisecond
	defaultRetryWaitMax = 2 * time.Second
	defaultTimeout      = 15 * time.Second
)

type Config struct {
	RetryMax     int
	RetryWaitMin time.Duration
	RetryWaitMax time.Duration
	Timeout      time.Duration
}

func applyConfig(cfg Config) Config {
	if cfg.RetryMax <= 0 {
		cfg.RetryMax = defaultRetryMax
	}

	if cfg.RetryWaitMin <= 0 {
		cfg.RetryWaitMin = defaultRetryWaitMin
	}

	if cfg.RetryWaitMax <= 0 {
		cfg.RetryWaitMax = defaultRetryWaitMax
	}

	if cfg.Timeout <= 0 {
		cfg.Timeout = defaultTimeout
	}

	return cfg
}
