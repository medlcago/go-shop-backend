package middleware

import (
	"errors"
	"go-shop-backend/pkg/apperrors"
	"go-shop-backend/pkg/logger"
	"go-shop-backend/pkg/response"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

func ErrorHandler() fiber.ErrorHandler {
	return func(ctx fiber.Ctx, err error) error {
		status := fiber.StatusInternalServerError
		message := http.StatusText(status)

		var fiberErr *fiber.Error
		if errors.As(err, &fiberErr) {
			status = fiberErr.Code
			message = fiberErr.Message
		}

		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			status = appErr.Code
			message = appErr.Message
		}

		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			status = http.StatusBadRequest
			message = err.Error()
		}

		slog.Error("error handler", slog.Int("code", status), logger.Err(err))

		return response.Error(ctx, status, message)
	}
}
