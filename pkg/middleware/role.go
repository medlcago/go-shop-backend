package middleware

import (
	"go-shop-backend/pkg/apperror"
	"slices"

	"github.com/gofiber/fiber/v3"
)

func RequireRole[T ~string](roles ...T) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		userCtx := GetUserContext(ctx)

		if userCtx.Role == "" {
			return apperror.ErrForbidden
		}

		allowed := slices.Contains(roles, T(userCtx.Role))

		if !allowed {
			return apperror.ErrForbidden
		}

		return ctx.Next()
	}
}
