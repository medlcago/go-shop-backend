package middleware

import (
	"go-shop-backend/pkg/apperror"
	"go-shop-backend/pkg/token"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type contextType string

const (
	ctxUserContext contextType = "userContext"
)

func IdentityUser(manager token.Manager) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		userCtx := UserContext{
			SessionID: getSessionID(ctx),
		}

		ctx.Locals(ctxUserContext, userCtx)

		authHeader := ctx.Get("Authorization")
		if authHeader == "" {
			return ctx.Next()
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			return ctx.Next()
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := manager.ValidateToken(tokenString)
		if err != nil {
			return ctx.Next()
		}

		uid, err := uuid.Parse(claims.UserID)
		if err != nil {
			return ctx.Next()
		}

		userCtx.UserID = &uid
		userCtx.Role = claims.UserRole
		userCtx.TokenType = claims.TokenType

		ctx.Locals(ctxUserContext, userCtx)

		return ctx.Next()
	}
}

func RequireAuth() fiber.Handler {
	return func(ctx fiber.Ctx) error {
		userCtx := GetUserContext(ctx)

		if userCtx.UserID == nil {
			return apperror.ErrInvalidCredentials
		}

		return ctx.Next()
	}
}

func RequireTokenType(tokenType string) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		userCtx := GetUserContext(ctx)

		if userCtx.TokenType != tokenType {
			return apperror.ErrInvalidTokenType
		}

		return ctx.Next()
	}
}

func RequireSessionID() fiber.Handler {
	return func(ctx fiber.Ctx) error {
		userCtx := GetUserContext(ctx)

		if userCtx.SessionID == nil {
			return apperror.ErrInvalidSessionID
		}

		return ctx.Next()
	}
}

func getSessionID(ctx fiber.Ctx) *uuid.UUID {
	if h := ctx.Get("X-Session-ID"); h != "" {
		if id, err := uuid.Parse(h); err == nil {
			return &id
		}
	}
	return nil
}

type UserContext struct {
	UserID    *uuid.UUID
	SessionID *uuid.UUID
	Role      string
	TokenType string
}

func GetUserContext(ctx fiber.Ctx) UserContext {
	var userCtx UserContext

	if v := ctx.Locals(ctxUserContext); v != nil {
		if userContext, ok := v.(UserContext); ok {
			userCtx = userContext
		}
	}

	return userCtx
}
