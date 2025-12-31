package http

import (
	"go-shop-backend/config"
	"go-shop-backend/pkg/middleware"
	structValidator "go-shop-backend/pkg/validator"
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

func SetupApp(cfg *config.Config, log *slog.Logger, validate *validator.Validate) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:         "Go Shop API",
		ReadTimeout:     cfg.HttpServer.ReadTimeout,
		WriteTimeout:    cfg.HttpServer.WriteTimeout,
		IdleTimeout:     cfg.HttpServer.IdleTimeout,
		ErrorHandler:    middleware.ErrorHandler(log),
		StructValidator: structValidator.New(validate),
	})

	app.Use(recover.New())
	app.Use(middleware.Logger(log))
	app.Use(cors.New(cors.Config{
		AllowMethods:        cfg.Cors.AllowMethods,
		AllowOrigins:        cfg.Cors.AllowOrigins,
		AllowHeaders:        cfg.Cors.AllowHeaders,
		ExposeHeaders:       cfg.Cors.ExposeHeaders,
		AllowCredentials:    cfg.Cors.AllowCredentials,
		MaxAge:              cfg.Cors.MaxAge,
		AllowPrivateNetwork: cfg.Cors.AllowPrivateNetwork,
	}))

	return app
}
