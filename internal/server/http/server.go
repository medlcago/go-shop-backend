package http

import (
	"context"
	"fmt"
	"go-shop-backend/internal/core"
	authHttp "go-shop-backend/internal/delivery/http/auth"
	categoryHttp "go-shop-backend/internal/delivery/http/category"
	productHttp "go-shop-backend/internal/delivery/http/product"
	userHttp "go-shop-backend/internal/delivery/http/user"
	"go-shop-backend/pkg/logger"
	"go-shop-backend/pkg/middleware"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	httpSwagger "github.com/swaggo/http-swagger"
)

type Server struct {
	app  *fiber.App
	deps *core.Dependencies
}

func NewServer(deps *core.Dependencies) *Server {
	return &Server{
		app:  SetupApp(deps.Cfg, deps.Logger, deps.Validator),
		deps: deps,
	}
}

func (s *Server) App() *fiber.App {
	return s.app
}

func (s *Server) Run() {
	s.Init()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	serverErr := make(chan error, 1)
	go s.runServer(serverErr)

	select {
	case <-ctx.Done():
		s.deps.Logger.Info("Shutdown signal received")
		s.Shutdown(s.deps.Cfg.HttpServer.ShutdownTimeout)
	case err := <-serverErr:
		if err != nil {
			s.deps.Logger.Error("Http Server error", logger.Err(err))
		}
		s.Shutdown(s.deps.Cfg.HttpServer.ShutdownTimeout)
	}
}

func (s *Server) runServer(errChan chan<- error) {
	addr := fmt.Sprintf(":%d", s.deps.Cfg.HttpServer.Port)
	s.deps.Logger.Info("HTTP server starting",
		slog.String("addr", addr),
		slog.String("env", s.deps.Cfg.Environment),
	)

	err := s.app.Listen(addr, fiber.ListenConfig{
		DisableStartupMessage: s.deps.Cfg.Environment == string(logger.EnvProduction),
	})

	if err != nil {
		errChan <- err
	}
}

func (s *Server) Shutdown(timeout time.Duration) {
	s.deps.Logger.Info("Starting graceful shutdown",
		slog.String("timeout", timeout.String()),
	)

	if err := s.app.ShutdownWithTimeout(timeout); err != nil {
		s.deps.Logger.Error("HTTP server shutdown failed", logger.Err(err))
	} else {
		s.deps.Logger.Info("HTTP server stopped successfully")
	}

	if err := s.closeResources(); err != nil {
		s.deps.Logger.Error("Close resources failed", logger.Err(err))
	} else {
		s.deps.Logger.Info("Close resources successfully")
	}

	s.deps.Logger.Info("Graceful shutdown completed")
}

func (s *Server) closeResources() error {
	if err := s.deps.DB.Close(); err != nil {
		return err
	}
	return nil
}

func (s *Server) Init() {
	s.app.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")))

	v1 := s.app.Group("/api/v1")

	authMiddleware := middleware.JWTAuth(s.deps.Cfg.AuthSecret)

	authHandler := authHttp.NewHandler(s.deps.AuthService)
	authHttp.RegisterRoutes(v1, authHandler)

	userHandler := userHttp.NewHandler(s.deps.UserService)
	userHttp.RegisterRoutes(v1, userHandler, authMiddleware)

	productHandler := productHttp.NewHandler(s.deps.ProductService)
	productHttp.RegisterRoutes(v1, productHandler, authMiddleware)

	categoryHandler := categoryHttp.NewHandler(s.deps.CategoryService)
	categoryHttp.RegisterRoutes(v1, categoryHandler)
}
