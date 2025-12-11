package main

import (
	"go-shop-backend/config"
	_ "go-shop-backend/docs"
	httpServer "go-shop-backend/internal/server/http"
	"go-shop-backend/pkg/logger"
	"log/slog"
	"time"
)

func init() {
	time.Local = time.UTC
}

// @title Go Shop Backend API
// @version 1.0

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token. "Bearer {token}"
func main() {
	cfg := config.MustLoadConfig()

	l := logger.NewSlog(logger.Env(cfg.Environment))
	slog.SetDefault(l)

	httpSrv := httpServer.NewServer(cfg)
	httpSrv.Run()
}
