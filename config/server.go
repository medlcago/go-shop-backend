package config

import "time"

type HttpServer struct {
	Port int `env:"SERVER_PORT" env-default:"8080"`

	ReadTimeout  time.Duration `env:"SERVER_READ_TIMEOUT" env-default:"10s"`
	WriteTimeout time.Duration `env:"SERVER_WRITE_TIMEOUT" env-default:"10s"`
	IdleTimeout  time.Duration `env:"SERVER_IDLE_TIMEOUT" env-default:"30s"`
}
