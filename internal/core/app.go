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

	signalCtx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	g, groupCtx := errgroup.WithContext(signalCtx)

	for _, srv := range a.servers {
		g.Go(func() error {
			if err := srv.Start(groupCtx); err != nil {
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
	case <-groupCtx.Done():
		a.container.Logger().Info("Shutdown signal received")
	case err := <-done:
		if err != nil {
			a.container.Logger().Error("Server error", logger.Err(err))
		}
	}

	return a.shutdown(ctx)
}

func (a *App) shutdown(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, a.container.Config().ShutdownTimeout)
	defer cancel()

	g, groupCtx := errgroup.WithContext(shutdownCtx)

	for _, srv := range a.servers {
		g.Go(func() error {
			if err := srv.Stop(groupCtx); err != nil {
				return fmt.Errorf("srv.Stop failed: %s: %w", srv.Name(), err)
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		a.container.Logger().Error("Shutdown error", logger.Err(err))
		return err
	}

	if err := a.container.Close(); err != nil {
		a.container.Logger().Error("Close container failed", logger.Err(err))
		return err
	}

	return nil
}
