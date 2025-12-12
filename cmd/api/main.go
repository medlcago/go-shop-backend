package main

import (
	"go-shop-backend/config"
	_ "go-shop-backend/docs"
	"go-shop-backend/internal/core"
	httpServer "go-shop-backend/internal/server/http"
	"go-shop-backend/pkg/database/postgres"
	"go-shop-backend/pkg/logger"
	"time"
)

func init() {
	time.Local = time.UTC
}

//	@title		Go Shop Backend API
//	@version	1.0

//	@host		localhost:8080
//	@BasePath	/api/v1

// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
// @description				Type "Bearer" followed by a space and JWT token. "Bearer {token}"
func main() {
	cfg := config.MustLoad()

	deps := core.NewDependencies(cfg)

	deps.Logger.Info("starting database migration...")
	if err := postgres.Migrate(deps.DB.DB); err != nil {
		logger.Fatal(deps.Logger, "failed to migrate database", err)
	}

	httpSrv := httpServer.NewServer(deps)
	httpSrv.Run()
}
