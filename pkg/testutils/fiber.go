package testutils

import (
	"go-shop-backend/pkg/middleware"
	structValidator "go-shop-backend/pkg/validator"
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

func CreateTestApp() *fiber.App {
	return fiber.New(fiber.Config{
		ErrorHandler:    middleware.ErrorHandler(slog.New(slog.DiscardHandler)),
		StructValidator: structValidator.New(validator.New()),
	})
}
