package token

import (
	"fmt"
	"go-shop-backend/pkg/utils"
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

func (j JWT) GenerateAccessToken(payload map[string]any) (string, error) {
	payload["type"] = AccessTokenType
	claims := jwt.MapClaims{
		"payload": payload,
		"exp":     time.Now().UTC().Add(j.accessTokenExpiredTime).Unix(),
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := jwtToken.SignedString([]byte(j.secretKey))
	if err != nil {
		return "", err
	}

	return token, nil
}

func (j JWT) GenerateRefreshToken(payload map[string]any) (string, error) {
	payload["type"] = RefreshTokenType
	claims := jwt.MapClaims{
		"payload": payload,
		"exp":     time.Now().UTC().Add(j.refreshTokenExpiredTime).Unix(),
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := jwtToken.SignedString([]byte(j.secretKey))
	if err != nil {
		return "", err
	}

	return token, nil
}

func (j JWT) ValidateToken(tokenString string) (map[string]interface{}, error) {
	tokenData := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, &tokenData, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrInvalidKey
	}

	var data map[string]interface{}
	if err := utils.Copy(&data, tokenData["payload"]); err != nil {
		return nil, fmt.Errorf("validate token: failed to copy payload: %w", err)
	}

	return data, nil
}
