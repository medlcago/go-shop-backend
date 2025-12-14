package core

import (
	"context"
	"go-shop-backend/config"
	"go-shop-backend/internal/repository"
	postgresRepo "go-shop-backend/internal/repository/postgres"
	"go-shop-backend/internal/service"
	"go-shop-backend/pkg/database/postgres"
	"go-shop-backend/pkg/logger"
	"go-shop-backend/pkg/transaction"
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

type Dependencies struct {
	Cfg       *config.Config
	Logger    *slog.Logger
	Validator *validator.Validate

	DB        *sqlx.DB
	TxManager transaction.Manager

	UserRepository repository.UserRepository

	AuthService service.AuthService
	UserService service.UserService
}

func NewDependencies(cfg *config.Config) *Dependencies {
	l := logger.NewSlog(logger.Env(cfg.Environment))
	slog.SetDefault(l)

	validate := validator.New()

	pgDB, err := postgres.New(cfg.DatabaseURI)
	if err != nil {
		logger.Fatal(l, "failed to connect to database", err)
	}

	getQueryer := func(ctx context.Context) transaction.Queryer {
		return transaction.GetQueryer(ctx, pgDB)
	}

	txManager := transaction.NewManager(pgDB)

	userRepo := postgresRepo.NewUserRepository(getQueryer)

	authService := service.NewAuthService(userRepo, txManager, cfg.AuthSecret)
	userService := service.NewUserService(userRepo)

	return &Dependencies{
		Cfg:            cfg,
		Logger:         l,
		Validator:      validate,
		DB:             pgDB,
		TxManager:      txManager,
		UserRepository: userRepo,
		AuthService:    authService,
		UserService:    userService,
	}
}
