package middleware

import (
	"encoding/json"
	"errors"
	"go-shop-backend/pkg/apperrors"
	"go-shop-backend/pkg/logger"
	"go-shop-backend/pkg/response"
	"go-shop-backend/pkg/utils"
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
			message = "Validation failed"
			details = utils.HumanizeValidationError(err)
		}

		var jsonUnmarshalTypeErr *json.UnmarshalTypeError
		if errors.As(err, &jsonUnmarshalTypeErr) {
			status = http.StatusBadRequest
			message = http.StatusText(status)
		}

		var jsonSyntaxErr *json.SyntaxError
		if errors.As(err, &jsonSyntaxErr) {
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
