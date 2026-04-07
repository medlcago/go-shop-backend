package core

import (
	"context"
	"errors"
	"go-shop-backend/internal/server"
	"go-shop-backend/pkg/logger"
	"os/signal"
	"syscall"
)

type App struct {
	servers   []server.Server
	container *Container
}

func NewApp(container *Container, servers ...server.Server) *App {
	return &App{
		servers:   servers,
		container: container,
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
		a.container.Logger().Info("Shutdown signal received")
	case err := <-errCh:
		a.container.Logger().Error("Server error", logger.Err(err))
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), a.container.Config().ShutdownTimeout)
	defer cancel()

	for _, srv := range a.servers {
		err := srv.Stop(shutdownCtx)
		if err != nil {
			a.container.Logger().Error("srv.Stop failed", logger.Err(err))
		}
	}

	return a.container.Close()
}
