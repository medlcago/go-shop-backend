package http

import (
	"go-shop-backend/config"
	"go-shop-backend/pkg/middleware"
	"log/slog"

	structValidator "go-shop-backend/pkg/validator"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

func SetupApp(cfg *config.Config, log *slog.Logger, validate *validator.Validate) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:         "Go Shop API",
		ReadTimeout:     cfg.ServerReadTimeout,
		WriteTimeout:    cfg.ServerWriteTimeout,
		IdleTimeout:     cfg.ServerIdleTimeout,
		ErrorHandler:    middleware.ErrorHandler(log),
		StructValidator: structValidator.New(validate),
	})

	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowMethods:        cfg.AllowMethods,
		AllowOrigins:        cfg.AllowOrigins,
		AllowHeaders:        cfg.AllowHeaders,
		ExposeHeaders:       cfg.ExposeHeaders,
		AllowCredentials:    cfg.AllowCredentials,
		MaxAge:              cfg.MaxAge,
		AllowPrivateNetwork: cfg.AllowPrivateNetwork,
	}))

	return app
}
