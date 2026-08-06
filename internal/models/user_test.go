package models

import (
	"testing"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/stretchr/testify/assert"
)

func TestUser_WebAuthnCredentials(t *testing.T) {
	user := &User{
		Passkeys: []PasskeyCredential{
			{
				Credential: webauthn.Credential{
					ID: []byte{1, 2, 3},
				},
			},
			{
				Credential: webauthn.Credential{
					ID: []byte{4, 5, 6},
				},
			},
		},
	}

	credentials := user.WebAuthnCredentials()

	assert.Len(t, credentials, len(user.Passkeys))
	assert.Equal(t, user.Passkeys[0].Credential.ID, credentials[0].ID)
	assert.Equal(t, user.Passkeys[1].Credential.ID, credentials[1].ID)
}
