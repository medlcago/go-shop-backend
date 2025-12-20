package middleware

import (
	"go-shop-backend/pkg/apperrors"
	"go-shop-backend/pkg/jtoken"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

func JWTAuth(secretKey string) fiber.Handler {
	return JWT(jtoken.AccessTokenType, secretKey)
}

func JWT(tokenType string, secretKey string) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		authHeader := ctx.Get("Authorization")

		if authHeader == "" {
			return apperrors.ErrInvalidCredentials
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			return apperrors.ErrInvalidCredentials
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := jtoken.ValidateToken(tokenString, secretKey)
		if err != nil {
			return apperrors.ErrInvalidCredentials
		}

		t := claims["type"].(string)
		if t != tokenType {
			return apperrors.ErrInvalidCredentials
		}

		userIDStr := claims["user_id"].(string)
		if userIDStr == "" {
			return apperrors.ErrInvalidCredentials
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return apperrors.ErrInvalidCredentials
		}

		ctx.Locals("userID", userID)
		return ctx.Next()
	}
}
