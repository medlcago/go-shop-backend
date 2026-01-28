package hasher

import (
	"github.com/alexedwards/argon2id"
)

type Argon2ID struct {
}

func NewArgon2ID() *Argon2ID {
	return &Argon2ID{}
}

func (a Argon2ID) Hash(password string) (string, error) {
	return argon2id.CreateHash(password, argon2id.DefaultParams)
}

func (a Argon2ID) Verify(password string, hash string) (bool, error) {
	return argon2id.ComparePasswordAndHash(password, hash)
}
