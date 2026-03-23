package config

import "time"

type Worker struct {
	Interval  time.Duration `env:"INTERVAL" env-default:"10m"`
	Timeout   time.Duration `env:"TIMEOUT" env-default:"10s"`
	BatchSize int           `env:"BATCH_SIZE" env-default:"100"`
}
