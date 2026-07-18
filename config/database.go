package config

import "time"

type Database struct {
	URI         string `env:"URI" env-required:"true"`
	Dialect     string `env:"DIALECT" env-default:"postgres"`
	AutoMigrate bool   `env:"AUTO_MIGRATE" env-default:"false"`

	User     string `env:"USER" env-required:"true"`
	Password string `env:"PASSWORD" env-required:"true"`
	Name     string `env:"NAME" env-required:"true"`

	MaxOpenConns    int           `env:"MAX_OPEN_CONNS" env-default:"100"`
	MaxIdleConns    int           `env:"MAX_IDLE_CONNS" env-default:"50"`
	ConnMaxLifetime time.Duration `env:"CONN_MAX_LIFETIME" env-default:"30m"`
	ConnMaxIdleTime time.Duration `env:"CONN_MAX_IDLE_TIME" env-default:"5m"`
}
