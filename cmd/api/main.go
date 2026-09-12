package main

import (
	"context"
	"go-shop-backend/config"
	_ "go-shop-backend/docs"
	"go-shop-backend/internal/core"
	asynqServer "go-shop-backend/internal/server/asynq"
	httpServer "go-shop-backend/internal/server/http"
	"go-shop-backend/pkg/logger"
	"log/slog"
	"time"
)

//	@title		Go Shop Backend API
//	@version	1.0

//	@host		localhost:8080
//	@BasePath	/api/v1

// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
// @description				Type "Bearer" followed by a space and JWT token. "Bearer {token}"
func main() {
	time.Local = time.UTC

	cfg := config.MustLoad()

	container := core.NewContainer(cfg)
	log := container.Logger()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if cfg.Database.AutoMigrate {
		log.Info("starting database migration...")
		result, err := container.DB().Migrate(ctx, cfg.Database.Dialect)
		if err != nil {
			logger.Fatal(log, "failed to migrate database", err)
		}

		log.InfoContext(
			ctx,
			"database migrated successfully",
			slog.String("result", result),
		)
	}

	httpSrv := httpServer.NewServer(container)
	asynqSrv := asynqServer.NewServer(container)

	application := core.NewApp(
		container,
		httpSrv,
		asynqSrv,
	)

	if err := application.Run(ctx); err != nil {
		logger.Fatal(log, "application.Run failed", err)
	}
}
