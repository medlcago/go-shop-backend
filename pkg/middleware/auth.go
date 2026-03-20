package middleware

import (
	"go-shop-backend/pkg/apperrors"
	"go-shop-backend/pkg/token"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

const (
	ctxUserID    = "userID"
	ctxUserRole  = "userRole"
	ctxSessionID = "sessionID"
	ctxIsAuth    = "isAuth"
)

func OptionalAuth(manager token.Manager) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		sid := getSessionID(ctx)
		ctx.Locals(ctxSessionID, sid)

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

		if claims.TokenType != token.AccessTokenType {
			return ctx.Next()
		}

		ctx.Locals(ctxUserID, uid)
		ctx.Locals(ctxUserRole, claims.UserRole)
		ctx.Locals(ctxIsAuth, true)

		return ctx.Next()
	}
}

func RequireAuth() fiber.Handler {
	return func(ctx fiber.Ctx) error {
		userCtx := GetUserContext(ctx)

		if !userCtx.IsAuth || userCtx.UserID == nil {
			return apperrors.ErrInvalidCredentials
		}

		return ctx.Next()
	}
}

func RequireSessionID() fiber.Handler {
	return func(ctx fiber.Ctx) error {
		userCtx := GetUserContext(ctx)

		if userCtx.SessionID == nil {
			return apperrors.ErrInvalidCredentials
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
	IsAuth    bool
}

func GetUserContext(ctx fiber.Ctx) UserContext {
	var userCtx UserContext

	if v := ctx.Locals(ctxUserID); v != nil {
		if id, ok := v.(uuid.UUID); ok {
			userCtx.UserID = &id
		}
	}

	if v := ctx.Locals(ctxSessionID); v != nil {
		if id, ok := v.(*uuid.UUID); ok {
			userCtx.SessionID = id
		}
	}

	if v := ctx.Locals(ctxUserRole); v != nil {
		if role, ok := v.(string); ok {
			userCtx.Role = role
		}
	}

	if v := ctx.Locals(ctxIsAuth); v != nil {
		if isAuth, ok := v.(bool); ok {
			userCtx.IsAuth = isAuth
		}
	}

	return userCtx
}
