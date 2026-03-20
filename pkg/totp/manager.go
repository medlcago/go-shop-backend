package totp

type Secret struct {
	Secret string
	QRCode string
}

type Manager interface {
	GenerateSecret(email string) (*Secret, error)
	ValidateCode(secret, code string) bool
}
