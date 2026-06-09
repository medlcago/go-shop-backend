package token

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWT struct {
	secretKey               string
	accessTokenExpiredTime  time.Duration
	refreshTokenExpiredTime time.Duration
	partialTokenExpiredTime time.Duration
}

func NewJWT(
	secretKey string,
	accessTokenExpiredTime time.Duration,
	refreshTokenExpiredTime time.Duration,
	partialTokenExpiredTime time.Duration,
) *JWT {
	return &JWT{
		secretKey:               secretKey,
		accessTokenExpiredTime:  accessTokenExpiredTime,
		refreshTokenExpiredTime: refreshTokenExpiredTime,
		partialTokenExpiredTime: partialTokenExpiredTime,
	}
}

func (j JWT) GenerateAccessToken(payload Payload) (string, error) {
	exp := time.Now().UTC().Add(j.accessTokenExpiredTime)
	return j.generateToken(payload, AccessTokenType, exp)
}

func (j JWT) GenerateRefreshToken(payload Payload) (string, error) {
	exp := time.Now().UTC().Add(j.refreshTokenExpiredTime)
	return j.generateToken(payload, RefreshTokenType, exp)
}

func (j JWT) GeneratePartialToken(payload Payload) (string, error) {
	exp := time.Now().UTC().Add(j.partialTokenExpiredTime)
	return j.generateToken(payload, PartialTokenType, exp)
}

func (j JWT) ValidateToken(tokenString string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrInvalidKey
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return claims, nil
}

func (j JWT) generateToken(payload Payload, tokenType string, exp time.Time) (string, error) {
	claims := UserClaims{
		Payload:   payload,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			ID:        uuid.NewString(),
		},
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := jwtToken.SignedString([]byte(j.secretKey))
	if err != nil {
		return "", err
	}

	return token, nil
}
