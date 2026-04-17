package config

type S3 struct {
	Endpoint  string `env:"ENDPOINT" env-required:"true"`
	AccessKey string `env:"ACCESS_KEY" env-required:"true"`
	SecretKey string `env:"SECRET_KEY" env-required:"true"`
	UseSSL    bool   `env:"USE_SSL" env-default:"false"`
	BaseURL   string `env:"BASE_URL" env-required:"true"`
	Bucket    string `env:"BUCKET" env-required:"true"`
	Region    string `env:"REGION" env-required:"true"`
}
