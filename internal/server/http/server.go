package http

import (
	"context"
	"errors"
	"fmt"
	"go-shop-backend/config"
	authHttp "go-shop-backend/internal/delivery/http/auth"
	postgresRepo "go-shop-backend/internal/repository/postgres"
	"go-shop-backend/internal/service"
	"go-shop-backend/pkg/database/postgres"
	"go-shop-backend/pkg/logger"
	"go-shop-backend/pkg/transaction"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jmoiron/sqlx"
	httpSwagger "github.com/swaggo/http-swagger"
)

type Server struct {
	app    *fiber.App
	logger *slog.Logger
	cfg    *config.Config
	pgDB   *sqlx.DB
}

func NewServer(cfg *config.Config) *Server {
	return &Server{
		app:    SetupApp(cfg),
		logger: slog.Default(),
		cfg:    cfg,
	}
}

func (s *Server) Run() {
	s.Init()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	serverErr := make(chan error, 1)
	go s.runServer(serverErr)

	select {
	case <-ctx.Done():
		s.logger.Info("Shutdown signal received")
		s.Shutdown(30 * time.Second)
	case err := <-serverErr:
		if err != nil {
			s.logger.Error("Server error", logger.Err(err))
		}
		s.Shutdown(10 * time.Second)
	}
}

func (s *Server) runServer(errChan chan<- error) {
	addr := fmt.Sprintf(":%d", s.cfg.HttpPort)
	s.logger.Info("HTTP server starting",
		slog.Int("port", s.cfg.HttpPort),
		slog.String("env", s.cfg.Environment))

	if err := s.app.Listen(addr, fiber.ListenConfig{
		DisableStartupMessage: s.cfg.Environment == string(logger.EnvProduction),
	}); err != nil && !errors.Is(err, http.ErrServerClosed) {
		errChan <- err
	} else {
		close(errChan)
	}
}

func (s *Server) Shutdown(timeout time.Duration) {
	s.logger.Info("Starting graceful shutdown",
		slog.String("timeout", timeout.String()))

	if err := s.shutdownHTTPServer(timeout); err != nil {
		s.logger.Error("HTTP server shutdown failed", logger.Err(err))
	} else {
		s.logger.Info("HTTP server stopped successfully")
	}

	if err := s.closeResources(); err != nil {
		s.logger.Error("Close resources failed", logger.Err(err))
	} else {
		s.logger.Info("Close resources successfully")
	}

	s.logger.Info("Graceful shutdown completed")
}

func (s *Server) shutdownHTTPServer(timeout time.Duration) error {
	if s.app == nil {
		return nil
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return s.app.ShutdownWithContext(shutdownCtx)
}

func (s *Server) closeResources() error {
	if s.pgDB == nil {
		return nil
	}

	if err := s.pgDB.Close(); err != nil {
		return err
	}
	return nil
}

func (s *Server) Init() {
	pgDB, err := postgres.New(s.cfg.DatabaseURI)
	if err != nil {
		logger.Fatal(s.logger, "failed to connect to database", err)
	}
	s.pgDB = pgDB

	s.logger.Info("starting database migration...")
	if err := postgres.Migrate(pgDB.DB); err != nil {
		logger.Fatal(s.logger, "failed to migrate database", err)
	}

	s.app.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")))

	getQueryer := func(ctx context.Context) transaction.Queryer {
		return transaction.GetQueryer(ctx, pgDB)
	}

	txManager := transaction.NewManager(pgDB)

	userRepo := postgresRepo.NewUserRepository(getQueryer)
	authService := service.NewAuthService(userRepo, txManager, s.cfg.AuthSecret)

	v1 := s.app.Group("/api/v1")

	authHandler := authHttp.NewHandler(authService)
	authHttp.RegisterRoutes(v1, authHandler)
}
