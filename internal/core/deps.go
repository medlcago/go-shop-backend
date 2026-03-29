package core

import (
	"context"
	"go-shop-backend/config"
	"go-shop-backend/internal/repository"
	gormRepo "go-shop-backend/internal/repository/gorm"
	"go-shop-backend/internal/service"
	"go-shop-backend/internal/tasks"
	"go-shop-backend/internal/upload"
	"go-shop-backend/pkg/contenttype"
	"go-shop-backend/pkg/crypto"
	"go-shop-backend/pkg/database"
	"go-shop-backend/pkg/hasher"
	"go-shop-backend/pkg/logger"
	"go-shop-backend/pkg/paymentprovider"
	"go-shop-backend/pkg/paymentprovider/yookassa"
	"go-shop-backend/pkg/redis"
	"go-shop-backend/pkg/storage"
	"go-shop-backend/pkg/storage/minio"
	"go-shop-backend/pkg/token"
	"go-shop-backend/pkg/totp"
	"log/slog"

	"github.com/go-playground/validator/v10"
)

type Dependencies struct {
	Cfg       *config.Config
	Logger    *slog.Logger
	Validator *validator.Validate

	DB        database.DB
	TxManager database.TxManager

	Storage       storage.Storage
	Redis         *redis.Client
	TokenManager  token.Manager
	UploadManager upload.Manager
	TaskFactory   tasks.TaskFactory

	PaymentProvider paymentprovider.Provider

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
	OrderService    service.OrderService
}

func NewDependencies(cfg *config.Config) *Dependencies {
	log := logger.NewSlog(logger.Env(cfg.Environment))
	slog.SetDefault(log)

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
		logger.Fatal(log, "failed to connect to database", err)
	}

	txManager := database.NewGormManager(db.GetDB(ctx))

	rdb, err := redis.New(
		cfg.Redis.Address,
		cfg.Redis.Password,
		redis.WithDB(cfg.Redis.DB),
		redis.WithDialTimeout(cfg.Redis.DialTimeout),
		redis.WithReadTimeout(cfg.Redis.ReadTimeout),
		redis.WithWriteTimeout(cfg.Redis.WriteTimeout),
		redis.WithPoolSize(cfg.Redis.PoolSize),
		redis.WithMinIdleConns(cfg.Redis.MinIdleConns),
	)
	if err != nil {
		logger.Fatal(log, "failed to create redis client", err)
	}

	minioStorage, err := minio.New(cfg.Minio)
	if err != nil {
		logger.Fatal(log, "failed to create minio storage", err)
	}

	jwtManager := token.NewJWT(cfg.AuthSecret, cfg.AccessTokenExpiredTime, cfg.RefreshTokenExpiredTime, cfg.PartialTokenExpiredTime)

	passwordHasher := hasher.NewArgon2ID()

	contentTypeDetector := contenttype.NewMagicDetector()

	paymentProvider, err := yookassa.New(
		yookassa.NewConfig(cfg.Yookassa.AccountID, cfg.Yookassa.SecretKey, cfg.Yookassa.ReturnURL),
	)
	if err != nil {
		logger.Fatal(log, "failed to create payment provider", err)
	}

	totpManager := totp.New(cfg.AppName)

	encryptionManager, err := crypto.NewAESGCMEncryptionManagerFromBase64(cfg.MasterKey)
	if err != nil {
		logger.Fatal(log, "failed to create encryption manager", err)
	}

	taskFactory := tasks.NewTaskFactory(rdb.RDB())

	uploadPolicyProvider, err := NewUploadPolicyProvider()
	if err != nil {
		logger.Fatal(log, "failed to create upload policy provider", err)
	}

	userRepo := gormRepo.NewUserRepository(db)
	productRepo := gormRepo.NewProductRepository(db)
	categoryRepo := gormRepo.NewCategoryRepository(db)
	uploadRepo := gormRepo.NewUploadRepository(db)
	orderRepo := gormRepo.NewOrderRepository(db)
	orderItemRepo := gormRepo.NewOrderItemRepository(db)

	uploadManager := upload.NewManager(minioStorage, uploadRepo, cfg.Upload, contentTypeDetector, uploadPolicyProvider, log)

	authService := service.NewAuthService(userRepo, jwtManager, passwordHasher, totpManager, encryptionManager, txManager)
	userService := service.NewUserService(userRepo)
	productService := service.NewProductService(productRepo, uploadManager)
	categoryService := service.NewCategoryService(categoryRepo)
	orderService := service.NewOrderService(orderRepo, orderItemRepo, productRepo, paymentProvider, taskFactory.Orders(), txManager, cfg.OrderCancelDelay)

	return &Dependencies{
		Cfg:       cfg,
		Logger:    log,
		Validator: validate,

		DB:        db,
		TxManager: txManager,

		Storage:      minioStorage,
		Redis:        rdb,
		TokenManager: jwtManager,
		TaskFactory:  taskFactory,

		UploadManager:    uploadManager,
		UploadRepository: uploadRepo,

		PaymentProvider: paymentProvider,

		UserRepository:      userRepo,
		ProductRepository:   productRepo,
		CategoryRepository:  categoryRepo,
		OrderRepository:     orderRepo,
		OrderItemRepository: orderItemRepo,

		AuthService:     authService,
		UserService:     userService,
		ProductService:  productService,
		CategoryService: categoryService,
		OrderService:    orderService,
	}
}
