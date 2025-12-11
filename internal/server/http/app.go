package http

import (
	"go-shop-backend/config"
	"go-shop-backend/pkg/middleware"
	"go-shop-backend/pkg/validator"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

func SetupApp(cfg *config.Config) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:         "Go Shop API",
		ReadTimeout:     cfg.ServerReadTimeout,
		WriteTimeout:    cfg.ServerWriteTimeout,
		IdleTimeout:     cfg.ServerIdleTimeout,
		ErrorHandler:    middleware.ErrorHandler(),
		StructValidator: validator.New(),
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
