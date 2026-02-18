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
		sessionID := getSessionID(ctx)
		if sessionID == nil {
			return apperrors.ErrInvalidCredentials
		}
		ctx.Locals(ctxSessionID, *sessionID)

		authHeader := ctx.Get("Authorization")
		if authHeader == "" {
			return ctx.Next()
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

		ctx.Locals(ctxUserID, userID)
		ctx.Locals(ctxUserRole, userRole)
		ctx.Locals(ctxIsAuth, true)

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
	SessionID uuid.UUID
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
		if id, ok := v.(uuid.UUID); ok {
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
