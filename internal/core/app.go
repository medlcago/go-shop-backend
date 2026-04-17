package core

import (
	"context"
	"errors"
	"fmt"
	"go-shop-backend/internal/server"
	"go-shop-backend/pkg/logger"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"
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

	g, ctx := errgroup.WithContext(ctx)

	for _, srv := range a.servers {
		g.Go(func() error {
			if err := srv.Start(ctx); err != nil {
				return fmt.Errorf("srv.Start failed: %w", err)
			}
			return nil
		})
	}

	done := make(chan error, 1)
	go func() {
		done <- g.Wait()
	}()

	select {
	case <-ctx.Done():
		a.container.Logger().Info("Shutdown signal received")
	case err := <-done:
		if err != nil {
			a.container.Logger().Error("Server error", logger.Err(err))
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), a.container.Config().ShutdownTimeout)
	defer cancel()

	gStop, shutdownCtx := errgroup.WithContext(shutdownCtx)

	for _, srv := range a.servers {
		gStop.Go(func() error {
			if err := srv.Stop(shutdownCtx); err != nil {
				return fmt.Errorf("srv.Stop failed: %s: %w", srv.Name(), err)
			}

			return nil
		})
	}

	if err := gStop.Wait(); err != nil {
		a.container.Logger().Error("Shutdown error", logger.Err(err))
		return err
	}

	if err := a.container.Close(); err != nil {
		a.container.Logger().Error("Close container failed", logger.Err(err))
		return err
	}

	return nil
}
