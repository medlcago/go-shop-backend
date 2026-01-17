package core

import (
	"context"
	"go-shop-backend/config"
	"go-shop-backend/internal/repository"
	gormRepo "go-shop-backend/internal/repository/gorm"
	"go-shop-backend/internal/service"
	"go-shop-backend/pkg/database"
	"go-shop-backend/pkg/logger"
	"go-shop-backend/pkg/storage"
	"go-shop-backend/pkg/storage/minio"
	"log/slog"

	"github.com/go-playground/validator/v10"
)

type Dependencies struct {
	Cfg       *config.Config
	Logger    *slog.Logger
	Validator *validator.Validate

	DB        database.DB
	TxManager database.TxManager

	Storage storage.Storage

	UserRepository     repository.UserRepository
	ProductRepository  repository.ProductRepository
	CategoryRepository repository.CategoryRepository

	AuthService     service.AuthService
	UserService     service.UserService
	ProductService  service.ProductService
	CategoryService service.CategoryService
	EntityService   service.EntityService
	UploadService   service.UploadService
}

func NewDependencies(cfg *config.Config) *Dependencies {
	l := logger.NewSlog(logger.Env(cfg.Environment))
	slog.SetDefault(l)

	validate := validator.New()
	ctx := context.Background()

	db, err := database.NewDatabase(
		cfg.Database.URI,
		database.WithMaxOpenConns(cfg.Database.MaxOpenConns),
		database.WithMaxIdleConns(cfg.Database.MaxIdleConns),
		database.WithConnMaxLifetime(cfg.Database.ConnMaxLifetime),
		database.WithConnMaxIdleTime(cfg.Database.ConnMaxIdleTime),
	)
	if err != nil {
		logger.Fatal(l, "failed to connect to database", err)
	}

	txManager := database.NewManager(db.GetDB(ctx))

	minioStorage, err := minio.New(cfg.Minio)
	if err != nil {
		logger.Fatal(l, "failed to create minio storage", err)
	}

	userRepo := gormRepo.NewUserRepository(db)
	productRepo := gormRepo.NewProductRepository(db)
	categoryRepo := gormRepo.NewCategoryRepository(db)

	authService := service.NewAuthService(userRepo, cfg.AuthSecret)
	userService := service.NewUserService(userRepo)
	productService := service.NewProductService(productRepo)
	categoryService := service.NewCategoryService(categoryRepo)
	entityService := service.NewEntityService(productRepo)
	uploadService := service.NewUploadService(minioStorage, entityService)

	return &Dependencies{
		Cfg:                cfg,
		Logger:             l,
		Validator:          validate,
		DB:                 db,
		TxManager:          txManager,
		Storage:            minioStorage,
		UserRepository:     userRepo,
		ProductRepository:  productRepo,
		CategoryRepository: categoryRepo,
		AuthService:        authService,
		UserService:        userService,
		ProductService:     productService,
		CategoryService:    categoryService,
		EntityService:      entityService,
		UploadService:      uploadService,
	}
}
