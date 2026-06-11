package http

import (
	"context"
	"fmt"
	"go-shop-backend/internal/core"
	categoryHttp "go-shop-backend/internal/delivery/http/category"
	orderHttp "go-shop-backend/internal/delivery/http/order"
	paymentHttp "go-shop-backend/internal/delivery/http/payment"
	productHttp "go-shop-backend/internal/delivery/http/product"
	userHttp "go-shop-backend/internal/delivery/http/user"
	wishlistHttp "go-shop-backend/internal/delivery/http/wishlist"
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
		app:       SetupApp(container),
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
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	s.Init()

	addr := fmt.Sprintf(":%d", s.container.Config().HttpServer.Port)

	s.logger.Info(
		"HTTP server starting",
		slog.String("addr", addr),
		slog.String("env", s.container.Config().Environment),
	)

	return s.app.Listen(addr, fiber.ListenConfig{
		DisableStartupMessage: !s.IsDevMode(),
		EnablePrintRoutes:     s.IsDevMode(),
	})

}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Stopping HTTP server")
	return s.app.ShutdownWithContext(ctx)
}

func (s *Server) Name() string {
	return "http"
}

func (s *Server) Init() {
	s.app.Use(middleware.IdentityUser(s.container.TokenManager()))

	if s.IsDevMode() {
		s.app.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")))
	}

	v1 := s.app.Group("/api/v1")

	userHandler := userHttp.NewHandler(s.container.UserService())
	userHttp.RegisterRoutes(v1, userHandler)

	productHandler := productHttp.NewHandler(s.container.ProductService())
	productHttp.RegisterRoutes(v1, productHandler)

	categoryHandler := categoryHttp.NewHandler(s.container.CategoryService())
	categoryHttp.RegisterRoutes(v1, categoryHandler)

	orderHandler := orderHttp.NewHandler(s.container.OrderService())
	orderHttp.RegisterRoutes(v1, orderHandler)

	wishlistHandler := wishlistHttp.NewHandler(s.container.WishlistService())
	wishlistHttp.RegisterRoutes(v1, wishlistHandler)

	paymentHandler := paymentHttp.NewHandler(s.container.PaymentService())
	paymentHttp.RegisterRoutes(v1, paymentHandler)
}
