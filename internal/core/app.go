package core

import (
	"context"
	"errors"
	"fmt"
	"go-shop-backend/internal/server"
	"go-shop-backend/pkg/logger"
	"os/signal"
	"syscall"
)

type App struct {
	servers []server.Server
	deps    *Dependencies
}

func NewApp(deps *Dependencies, servers ...server.Server) *App {
	return &App{
		servers: servers,
		deps:    deps,
	}
}
func (a *App) Run(ctx context.Context) error {
	if len(a.servers) == 0 {
		return errors.New("no servers available to run")
	}

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, len(a.servers))

	for _, srv := range a.servers {
		go func(s server.Server) {
			errCh <- s.Start(ctx)
		}(srv)
	}

	select {
	case <-ctx.Done():
		a.deps.Logger.Info("Shutdown signal received")
	case err := <-errCh:
		a.deps.Logger.Error("Server error", logger.Err(err))
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), a.deps.Cfg.ShutdownTimeout)
	defer cancel()

	for _, srv := range a.servers {
		err := srv.Stop(shutdownCtx)
		if err != nil {
			a.deps.Logger.Error("srv.Stop failed", logger.Err(err))
		}
	}

	return a.closeResources()
}

func (a *App) closeResources() error {
	if err := a.deps.DB.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}

	if err := a.deps.Redis.Close(); err != nil {
		return fmt.Errorf("failed to close redis: %w", err)
	}

	return nil
}
