package middleware

import (
	"errors"
	"go-shop-backend/pkg/apperror"
	"go-shop-backend/pkg/token"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/extractors"
	"github.com/google/uuid"
)

type AuthContextType string

const (
	CtxUserContext AuthContextType = "_userContext"
)

func IdentityUser(manager token.Manager) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		userCtx := UserContext{
			SessionID: getSessionID(ctx),
		}

		ctx.Locals(CtxUserContext, userCtx)

		tokenString, err := ExtractBearer(ctx)
		if err != nil {
			return ctx.Next()
		}

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
		userCtx.Token = tokenString
		userCtx.TokenType = claims.TokenType

		ctx.Locals(CtxUserContext, userCtx)

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
	Token     string
	TokenType string
}

func GetUserContext(ctx fiber.Ctx) UserContext {
	var userCtx UserContext

	if v := ctx.Locals(CtxUserContext); v != nil {
		if userContext, ok := v.(UserContext); ok {
			userCtx = userContext
		}
	}

	return userCtx
}

func ExtractBearer(ctx fiber.Ctx) (string, error) {
	extractor := extractors.FromAuthHeader("Bearer")
	tokenString, err := extractor.Extract(ctx)
	if err != nil {
		if errors.Is(err, extractors.ErrNotFound) {
			return "", apperror.ErrInvalidToken
		}

		return "", err
	}

	return tokenString, nil
}
