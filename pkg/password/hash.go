package password

import "github.com/alexedwards/argon2id"

func HashPassword(password string) (string, error) {
	return argon2id.CreateHash(password, argon2id.DefaultParams)
}

func VerifyPassword(password string, hash string) bool {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	return match && err == nil
}
