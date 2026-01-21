package middleware

import (
	"go-shop-backend/pkg/apperrors"
	"go-shop-backend/pkg/token"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type Auth interface {
	Handle() fiber.Handler
}

type jwtMiddleware struct {
	manager token.Manager
}

func NewJWT(manager token.Manager) Auth {
	return &jwtMiddleware{
		manager: manager,
	}
}

func (j jwtMiddleware) Handle() fiber.Handler {
	return func(ctx fiber.Ctx) error {
		authHeader := ctx.Get("Authorization")

		if authHeader == "" {
			return apperrors.ErrInvalidCredentials
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			return apperrors.ErrInvalidCredentials
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := j.manager.ValidateToken(tokenString)
		if err != nil {
			return apperrors.ErrInvalidCredentials
		}

		t, ok := claims["type"].(string)
		if !ok || t != token.AccessTokenType {
			return apperrors.ErrInvalidCredentials
		}

		userIDStr, ok := claims["user_id"].(string)
		if !ok {
			return apperrors.ErrInvalidCredentials
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return apperrors.ErrInvalidCredentials
		}

		userRole, ok := claims["role"].(string)
		if !ok {
			return apperrors.ErrInvalidCredentials
		}

		ctx.Locals("userID", userID)
		ctx.Locals("userRole", userRole)
		return ctx.Next()
	}
}
