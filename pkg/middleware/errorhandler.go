package middleware

import (
	"encoding/json"
	"errors"
	"go-shop-backend/pkg/apperrors"
	"go-shop-backend/pkg/logger"
	"go-shop-backend/pkg/response"
	structValidator "go-shop-backend/pkg/validator"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

func ErrorHandler(log *slog.Logger) fiber.ErrorHandler {
	return func(ctx fiber.Ctx, err error) error {
		status := fiber.StatusInternalServerError
		message := http.StatusText(status)
		var details any

		if fiberErr, ok := errors.AsType[*fiber.Error](err); ok {
			status = fiberErr.Code
			message = fiberErr.Message
		}

		if appErr, ok := errors.AsType[*apperrors.AppError](err); ok {
			status = appErr.Code
			message = appErr.Message
		}

		if _, ok := errors.AsType[validator.ValidationErrors](err); ok {
			status = http.StatusBadRequest
			message = "Validation failed"
			details = structValidator.HumanizeValidationError(err)
		}

		if _, ok := errors.AsType[*json.UnmarshalTypeError](err); ok {
			status = http.StatusBadRequest
			message = http.StatusText(status)
		}

		if _, ok := errors.AsType[*json.SyntaxError](err); ok {
			status = http.StatusBadRequest
			message = http.StatusText(status)
		}

		// Handling binding errors
		if strings.Contains(err.Error(), "bind:") {
			status = http.StatusBadRequest
			message = http.StatusText(status)
		}

		log.Error(
			"request error",
			slog.Int("status", status),
			logger.Err(err),
		)

		return response.Error(ctx, status, message, details)
	}
}
