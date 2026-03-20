package totp

import (
	"fmt"

	"github.com/pquerna/otp/totp"
)

type TOTP struct {
	issuer string
}

func New(issuer string) *TOTP {
	return &TOTP{
		issuer: issuer,
	}
}

func (t *TOTP) GenerateSecret(email string) (*Secret, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      t.issuer,
		AccountName: email,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate TOTP key: %w", err)
	}

	return &Secret{
		Secret: key.Secret(),
		QRCode: key.URL(),
	}, nil
}

func (t *TOTP) ValidateCode(secret, code string) bool {
	return totp.Validate(code, secret)
}
