package config

import "time"

type Redis struct {
	Address      string        `env:"ADDRESS" env-required:"true"`
	Password     string        `env:"PASSWORD" env-required:"true"`
	DB           int           `env:"DB" env-default:"0"`
	DialTimeout  time.Duration `env:"DIAL_TIMEOUT" env-default:"5s"`
	ReadTimeout  time.Duration `env:"READ_TIMEOUT" env-default:"3s"`
	WriteTimeout time.Duration `env:"WRITE_TIMEOUT" env-default:"3s"`
	PoolSize     int           `env:"POOL_SIZE" env-default:"10"`
	MinIdleConns int           `env:"MIN_IDLE_CONNS" env-default:"5"`
}
