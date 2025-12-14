package jtoken

import (
	"fmt"
	"go-shop-backend/pkg/utils"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	AccessTokenExpiredTime  = 30 * time.Minute
	RefreshTokenExpiredTime = 24 * 30 * time.Hour
	AccessTokenType         = "x-access"
	RefreshTokenType        = "x-refresh"
)

func GenerateAccessToken(payload map[string]any, secretKey string) (string, error) {
	payload["type"] = AccessTokenType
	claims := jwt.MapClaims{
		"payload": payload,
		"exp":     time.Now().Add(AccessTokenExpiredTime).Unix(),
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := jwtToken.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return token, nil
}

func GenerateRefreshToken(payload map[string]any, secretKey string) (string, error) {
	payload["type"] = RefreshTokenType
	claims := jwt.MapClaims{
		"payload": payload,
		"exp":     time.Now().Add(RefreshTokenExpiredTime).Unix(),
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := jwtToken.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return token, nil
}

func ValidateToken(tokenString string, secretKey string) (map[string]interface{}, error) {
	tokenData := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, &tokenData, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
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
