package token

import "github.com/golang-jwt/jwt/v5"

const (
	AccessTokenType  = "x-access"
	RefreshTokenType = "x-refresh"
)

type Payload struct {
	UserID   string `json:"user_id"`
	UserRole string `json:"user_role"`
}

type UserClaims struct {
	UserID    string `json:"user_id"`
	UserRole  string `json:"user_role"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

type Manager interface {
	GenerateAccessToken(payload Payload) (string, error)
	GenerateRefreshToken(payload Payload) (string, error)
	ValidateToken(tokenString string) (*UserClaims, error)
}
