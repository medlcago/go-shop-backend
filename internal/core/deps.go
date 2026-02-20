package core

import (
	"context"
	"go-shop-backend/config"
	"go-shop-backend/internal/repository"
	gormRepo "go-shop-backend/internal/repository/gorm"
	"go-shop-backend/internal/service"
	"go-shop-backend/pkg/contenttype"
	"go-shop-backend/pkg/database"
	"go-shop-backend/pkg/hasher"
	"go-shop-backend/pkg/logger"
	"go-shop-backend/pkg/storage"
	"go-shop-backend/pkg/storage/minio"
	"go-shop-backend/pkg/token"
	"log/slog"

	"github.com/go-playground/validator/v10"
)

type Dependencies struct {
	Cfg       *config.Config
	Logger    *slog.Logger
	Validator *validator.Validate

	DB        database.DB
	TxManager database.TxManager

	Storage      storage.Storage
	TokenManager token.Manager

	UserRepository      repository.UserRepository
	ProductRepository   repository.ProductRepository
	CategoryRepository  repository.CategoryRepository
	UploadRepository    repository.UploadRepository
	OrderRepository     repository.OrderRepository
	OrderItemRepository repository.OrderItemRepository

	AuthService     service.AuthService
	UserService     service.UserService
	ProductService  service.ProductService
	CategoryService service.CategoryService
	EntityService   service.EntityService
	UploadService   service.UploadService
	OrderService    service.OrderService
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

	jwtManager := token.NewJWT(cfg.AuthSecret, cfg.AccessTokenExpiredTime, cfg.RefreshTokenExpiredTime)

	passwordHasher := hasher.NewArgon2ID()

	contentTypeDetector := contenttype.NewMagicDetector()

	userRepo := gormRepo.NewUserRepository(db)
	productRepo := gormRepo.NewProductRepository(db)
	categoryRepo := gormRepo.NewCategoryRepository(db)
	uploadRepo := gormRepo.NewUploadRepository(db)
	orderRepo := gormRepo.NewOrderRepository(db)
	orderItemRepo := gormRepo.NewOrderItemRepository(db)

	authService := service.NewAuthService(userRepo, jwtManager, passwordHasher)
	userService := service.NewUserService(userRepo)
	entityService := service.NewEntityService(productRepo)
	uploadService := service.NewUploadService(minioStorage, entityService, uploadRepo, cfg.Upload, contentTypeDetector)
	productService := service.NewProductService(productRepo, uploadService)
	categoryService := service.NewCategoryService(categoryRepo)
	orderService := service.NewOrderService(orderRepo, orderItemRepo, productRepo, txManager)

	return &Dependencies{
		Cfg:       cfg,
		Logger:    l,
		Validator: validate,

		DB:        db,
		TxManager: txManager,

		Storage:      minioStorage,
		TokenManager: jwtManager,

		UserRepository:      userRepo,
		ProductRepository:   productRepo,
		CategoryRepository:  categoryRepo,
		UploadRepository:    uploadRepo,
		OrderRepository:     orderRepo,
		OrderItemRepository: orderItemRepo,

		AuthService:     authService,
		UserService:     userService,
		ProductService:  productService,
		CategoryService: categoryService,
		EntityService:   entityService,
		UploadService:   uploadService,
		OrderService:    orderService,
	}
}
