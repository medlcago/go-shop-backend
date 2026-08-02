package config

import "time"

type HTTPServer struct {
	Port int `env:"PORT" env-default:"8080"`

	ReadTimeout  time.Duration `env:"READ_TIMEOUT" env-default:"10s"`
	WriteTimeout time.Duration `env:"WRITE_TIMEOUT" env-default:"10s"`
	IdleTimeout  time.Duration `env:"IDLE_TIMEOUT" env-default:"30s"`
}
