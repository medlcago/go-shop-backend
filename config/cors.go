package config

type Cors struct {
	AllowMethods        []string `env:"ALLOW_METHODS" validate:"required,min=1"`
	AllowOrigins        []string `env:"ALLOW_ORIGINS" validate:"required,min=1"`
	AllowHeaders        []string `env:"ALLOW_HEADERS" validate:"required,min=1"`
	ExposeHeaders       []string `env:"EXPOSE_HEADERS"  validate:"required,min=1"`
	AllowCredentials    bool     `env:"ALLOW_CREDENTIALS"`
	MaxAge              int      `env:"ALLOW_MAX_AGE"`
	AllowPrivateNetwork bool     `env:"ALLOW_PRIVATE_NETWORK"`
}
