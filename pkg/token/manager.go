package token

import (
	"github.com/golang-jwt/jwt/v5"
)

const (
	AccessTokenType  = "x-access"
	RefreshTokenType = "x-refresh"
	PartialTokenType = "x-partial"
)

type Payload struct {
	UserID         string `json:"user_id"`
	UserRole       string `json:"user_role"`
	TwoFAEnabled   bool   `json:"two_fa_enabled"`
	EmailConfirmed bool   `json:"email_confirmed"`
}

type UserClaims struct {
	Payload
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

type Manager interface {
	GenerateAccessToken(payload Payload) (string, *UserClaims, error)
	GenerateRefreshToken(payload Payload) (string, *UserClaims, error)
	GeneratePartialToken(payload Payload) (string, *UserClaims, error)
	ValidateToken(tokenString string) (*UserClaims, error)
}
