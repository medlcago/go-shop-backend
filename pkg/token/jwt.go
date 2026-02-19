package token

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWT struct {
	secretKey               string
	accessTokenExpiredTime  time.Duration
	refreshTokenExpiredTime time.Duration
}

func NewJWT(secretKey string, accessTokenExpiredTime, refreshTokenExpiredTime time.Duration) *JWT {
	return &JWT{
		secretKey:               secretKey,
		accessTokenExpiredTime:  accessTokenExpiredTime,
		refreshTokenExpiredTime: refreshTokenExpiredTime,
	}
}

func (j JWT) GenerateAccessToken(payload Payload) (string, error) {
	exp := time.Now().UTC().Add(j.accessTokenExpiredTime)

	claims := UserClaims{
		UserID:    payload.UserID,
		UserRole:  payload.UserRole,
		TokenType: AccessTokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := jwtToken.SignedString([]byte(j.secretKey))
	if err != nil {
		return "", err
	}

	return token, nil
}

func (j JWT) GenerateRefreshToken(payload Payload) (string, error) {
	exp := time.Now().UTC().Add(j.refreshTokenExpiredTime)

	claims := UserClaims{
		UserID:    payload.UserID,
		UserRole:  payload.UserRole,
		TokenType: RefreshTokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := jwtToken.SignedString([]byte(j.secretKey))
	if err != nil {
		return "", err
	}

	return token, nil
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
