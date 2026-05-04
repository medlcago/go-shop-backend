package http

import (
	"go-shop-backend/internal/core"
	"go-shop-backend/pkg/middleware"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

func SetupApp(container *core.Container) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:         container.Config().AppName,
		ReadTimeout:     container.Config().HttpServer.ReadTimeout,
		WriteTimeout:    container.Config().HttpServer.WriteTimeout,
		IdleTimeout:     container.Config().HttpServer.IdleTimeout,
		ErrorHandler:    middleware.ErrorHandler(container.Logger()),
		StructValidator: container.Validator(),
	})

	app.Use(recover.New())
	app.Use(middleware.Logger(container.Logger()))
	app.Use(cors.New(cors.Config{
		AllowMethods:        container.Config().Cors.AllowMethods,
		AllowOrigins:        container.Config().Cors.AllowOrigins,
		AllowHeaders:        container.Config().Cors.AllowHeaders,
		ExposeHeaders:       container.Config().Cors.ExposeHeaders,
		AllowCredentials:    container.Config().Cors.AllowCredentials,
		MaxAge:              container.Config().Cors.MaxAge,
		AllowPrivateNetwork: container.Config().Cors.AllowPrivateNetwork,
	}))

	return app
}
