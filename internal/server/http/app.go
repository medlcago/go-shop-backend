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
		ReadTimeout:     container.Config().HTTPServer.ReadTimeout,
		WriteTimeout:    container.Config().HTTPServer.WriteTimeout,
		IdleTimeout:     container.Config().HTTPServer.IdleTimeout,
		ErrorHandler:    middleware.ErrorHandler(container.Logger()),
		StructValidator: container.Validator(),
		ProxyHeader:     fiber.HeaderXForwardedFor,
		TrustProxy:      true,
		TrustProxyConfig: fiber.TrustProxyConfig{
			Proxies: []string{"10.100.0.0/24"},
		},
		EnableIPValidation: true,
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
