package main

import (
	"context"
	"go-shop-backend/config"
	_ "go-shop-backend/docs"
	"go-shop-backend/internal/core"
	asynqServer "go-shop-backend/internal/server/asynq"
	httpServer "go-shop-backend/internal/server/http"
	"go-shop-backend/pkg/logger"
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

	deps := core.NewDependencies(cfg)

	deps.Logger.Info("starting database migration...")
	if err := deps.DB.Migrate("postgres"); err != nil {
		logger.Fatal(deps.Logger, "failed to migrate database", err)
	}

	httpSrv := httpServer.NewServer(deps)
	asynqSrv := asynqServer.NewServer(deps)

	application := core.NewApp(
		deps,
		httpSrv,
		asynqSrv,
	)

	if err := application.Run(context.Background()); err != nil {
		logger.Fatal(deps.Logger, "application.Run failed", err)
	}
}
