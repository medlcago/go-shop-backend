package http

import (
	"context"
	"fmt"
	"go-shop-backend/internal/core"
	authHttp "go-shop-backend/internal/delivery/http/auth"
	categoryHttp "go-shop-backend/internal/delivery/http/category"
	orderHttp "go-shop-backend/internal/delivery/http/order"
	productHttp "go-shop-backend/internal/delivery/http/product"
	userHttp "go-shop-backend/internal/delivery/http/user"
	webhookHttp "go-shop-backend/internal/delivery/http/webhook"
	"go-shop-backend/pkg/logger"
	"go-shop-backend/pkg/middleware"
	"log/slog"

	"github.com/gofiber/fiber/v3"
	httpSwagger "github.com/swaggo/http-swagger"
)

type Server struct {
	app       *fiber.App
	container *core.Container
	logger    *slog.Logger
}

func NewServer(container *core.Container) *Server {
	log := container.Logger().With("server", "http")

	return &Server{
		app:       SetupApp(container.Config(), container.Logger(), container.Validator()),
		container: container,
		logger:    log,
	}
}

func (s *Server) App() *fiber.App {
	return s.app
}

func (s *Server) IsDevMode() bool {
	return s.container.Config().Environment == string(logger.EnvDevelopment)
}

func (s *Server) Start(ctx context.Context) error {
	s.Init()

	addr := fmt.Sprintf(":%d", s.container.Config().HttpServer.Port)

	s.logger.Info(
		"HTTP server starting",
		slog.String("addr", addr),
		slog.String("env", s.container.Config().Environment),
	)

	go func() {
		<-ctx.Done()
		s.logger.Info("HTTP shutdown signal received")
		err := s.Stop(context.Background())
		if err != nil {
			s.logger.Error("s.Stop failed", logger.Err(err))
		}
	}()

	return s.app.Listen(addr, fiber.ListenConfig{
		DisableStartupMessage: !s.IsDevMode(),
	})

}

func (s *Server) Stop(ctx context.Context) error {
	timeout := s.container.Config().ShutdownTimeout

	s.logger.Info(
		"Stopping HTTP server",
		slog.String("timeout", timeout.String()),
	)

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return s.app.ShutdownWithContext(ctx)
}

func (s *Server) Init() {
	s.app.Use(middleware.OptionalAuth(s.container.TokenManager()))

	if s.IsDevMode() {
		s.app.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")))
	}

	v1 := s.app.Group("/api/v1")

	authHandler := authHttp.NewHandler(s.container.AuthService())
	authHttp.RegisterRoutes(v1, authHandler)

	userHandler := userHttp.NewHandler(s.container.UserService())
	userHttp.RegisterRoutes(v1, userHandler)

	productHandler := productHttp.NewHandler(s.container.ProductService())
	productHttp.RegisterRoutes(v1, productHandler)

	categoryHandler := categoryHttp.NewHandler(s.container.CategoryService())
	categoryHttp.RegisterRoutes(v1, categoryHandler)

	orderHandler := orderHttp.NewHandler(s.container.OrderService())
	orderHttp.RegisterRoutes(v1, orderHandler)

	webhookHandler := webhookHttp.NewHandler(s.container.OrderService())
	webhookHttp.RegisterRoutes(v1, webhookHandler)
}
