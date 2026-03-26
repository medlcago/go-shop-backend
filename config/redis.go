package config

type Redis struct {
	Address  string `env:"ADDRESS" env-required:"true"`
	Password string `env:"PASSWORD" env-required:"true"`
}
