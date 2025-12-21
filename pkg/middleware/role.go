package middleware

import (
	"go-shop-backend/pkg/apperrors"
	"slices"

	"github.com/gofiber/fiber/v3"
)

func RequireRole[T ~string](roles ...T) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		userRole := fiber.Locals[string](ctx, "userRole")

		if userRole == "" {
			return apperrors.ErrForbidden
		}

		allowed := slices.Contains(roles, T(userRole))

		if !allowed {
			return apperrors.ErrForbidden
		}

		return ctx.Next()
	}
}
