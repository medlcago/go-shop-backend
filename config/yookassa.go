package config

type Yookassa struct {
	AccountID string `env:"ACCOUNT_ID" env-required:"true"`
	SecretKey string `env:"SECRET_KEY" env-required:"true"`
	ReturnURL string `env:"RETURN_URL" env-required:"true"`
}
