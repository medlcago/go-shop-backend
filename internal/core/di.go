package core

import (
	"context"
	"database/sql"
	"fmt"
	"go-shop-backend/config"
	"go-shop-backend/internal/metrics"
	"go-shop-backend/internal/repository"
	gormRepo "go-shop-backend/internal/repository/gorm"
	"go-shop-backend/internal/service"
	"go-shop-backend/internal/tasks"
	"go-shop-backend/internal/templates"
	"go-shop-backend/internal/upload"
	"go-shop-backend/pkg/cache"
	"go-shop-backend/pkg/contenttype"
	"go-shop-backend/pkg/crypto"
	"go-shop-backend/pkg/database"
	"go-shop-backend/pkg/hasher"
	"go-shop-backend/pkg/logger"
	"go-shop-backend/pkg/notification"
	"go-shop-backend/pkg/paymentprovider"
	"go-shop-backend/pkg/paymentprovider/yookassa"
	"go-shop-backend/pkg/redis"
	"go-shop-backend/pkg/storage"
	"go-shop-backend/pkg/storage/minio"
	"go-shop-backend/pkg/template"
	"go-shop-backend/pkg/token"
	"go-shop-backend/pkg/totp"
	"go-shop-backend/pkg/validator"
	"log/slog"
	"sync"
)

type Lazy[T any] struct {
	once sync.Once
	val  T
}

func (l *Lazy[T]) Get(factory func() T) T {
	l.once.Do(func() {
		l.val = factory()
	})

	return l.val
}

type Container struct {
	cfg                  *config.Config
	logger               Lazy[*slog.Logger]
	validator            Lazy[validator.Validator]
	hasher               Lazy[hasher.Hasher]
	tokenManager         Lazy[token.Manager]
	totpManager          Lazy[totp.Manager]
	encryptionManager    Lazy[crypto.EncryptionManager]
	taskFactory          Lazy[*tasks.Factory]
	paymentProvider      Lazy[paymentprovider.Provider]
	db                   Lazy[database.DB]
	txManager            Lazy[database.TxManager]
	storage              Lazy[storage.Storage]
	redisClient          Lazy[*redis.Client]
	contentTypeDetector  Lazy[contenttype.Detector]
	uploadPolicyRegistry Lazy[upload.PolicyRegistry]
	uploadManager        Lazy[upload.Manager]
	cache                Lazy[cache.Cache]
	templateManager      Lazy[template.Manager]
	notificationRegistry Lazy[notification.SenderRegistry]
	metricsFactory       Lazy[*metrics.Factory]

	// repositories
	userRepository         Lazy[repository.UserRepository]
	productRepository      Lazy[repository.ProductRepository]
	categoryRepository     Lazy[repository.CategoryRepository]
	uploadRepository       Lazy[repository.UploadRepository]
	orderRepository        Lazy[repository.OrderRepository]
	orderItemRepository    Lazy[repository.OrderItemRepository]
	wishlistRepository     Lazy[repository.WishlistRepository]
	wishlistItemRepository Lazy[repository.WishlistItemRepository]
	addressRepository      Lazy[repository.AddressRepository]

	// services
	userService         Lazy[service.UserService]
	productService      Lazy[service.ProductService]
	categoryService     Lazy[service.CategoryService]
	orderService        Lazy[service.OrderService]
	wishlistService     Lazy[service.WishlistService]
	notificationService Lazy[service.NotificationService]
	inventoryService    Lazy[service.InventoryService]
	paymentService      Lazy[service.PaymentService]
	addressService      Lazy[service.AddressService]
}

func NewContainer(cfg *config.Config) *Container {
	return &Container{
		cfg: cfg,
	}
}

func (c *Container) Config() *config.Config {
	return c.cfg
}

func (c *Container) Logger() *slog.Logger {
	return c.logger.Get(func() *slog.Logger {
		return logger.NewSlog(logger.Env(c.Config().Environment))
	})
}

func (c *Container) Validator() validator.Validator {
	return c.validator.Get(func() validator.Validator {
		return NewValidator()
	})
}

func (c *Container) DB() database.DB {
	return c.db.Get(func() database.DB {
		db, err := database.New(
			c.Config().Database.URI,
			database.WithMaxOpenConns(c.Config().Database.MaxOpenConns),
			database.WithMaxIdleConns(c.Config().Database.MaxIdleConns),
			database.WithConnMaxLifetime(c.Config().Database.ConnMaxLifetime),
			database.WithConnMaxIdleTime(c.Config().Database.ConnMaxIdleTime),
			database.WithLogger(c.Logger()),
		)
		if err != nil {
			logger.Fatal(c.Logger(), "failed to connect to database", err)
		}

		return db
	})

}

func (c *Container) TxManager() database.TxManager {
	return c.txManager.Get(func() database.TxManager {
		return database.NewGormTxManager(c.DB().GetDB(context.Background()))
	})
}

func (c *Container) RedisClient() *redis.Client {
	return c.redisClient.Get(func() *redis.Client {
		redisClient, err := redis.New(
			c.Config().Redis.Address,
			c.Config().Redis.Password,
			redis.WithDB(c.Config().Redis.DB),
			redis.WithDialTimeout(c.Config().Redis.DialTimeout),
			redis.WithReadTimeout(c.Config().Redis.ReadTimeout),
			redis.WithWriteTimeout(c.Config().Redis.WriteTimeout),
			redis.WithPoolSize(c.Config().Redis.PoolSize),
			redis.WithMinIdleConns(c.Config().Redis.MinIdleConns),
		)
		if err != nil {
			logger.Fatal(c.Logger(), "failed to create redis client", err)
		}

		return redisClient
	})
}

func (c *Container) Storage() storage.Storage {
	return c.storage.Get(func() storage.Storage {
		minioStorage, err := minio.New(&minio.Config{
			Endpoint:  c.Config().S3.Endpoint,
			AccessKey: c.Config().S3.AccessKey,
			SecretKey: c.Config().S3.SecretKey,
			Secure:    c.Config().S3.UseSSL,
			Bucket:    c.Config().S3.Bucket,
			Region:    c.Config().S3.Region,
			BaseURL:   c.Config().S3.BaseURL,
		})

		if err != nil {
			logger.Fatal(c.Logger(), "failed to create minio storage", err)
		}

		return minioStorage
	})
}

func (c *Container) PasswordHasher() hasher.Hasher {
	return c.hasher.Get(func() hasher.Hasher {
		return hasher.NewArgon2ID()
	})
}

func (c *Container) TokenManager() token.Manager {
	return c.tokenManager.Get(func() token.Manager {
		tokenManager := token.NewJWT(
			c.Config().AuthSecret,
			c.Config().AppName,
			c.Config().AccessTokenExpiredTime,
			c.Config().RefreshTokenExpiredTime,
			c.Config().PartialTokenExpiredTime,
		)

		return tokenManager
	})
}

func (c *Container) ContentTypeDetector() contenttype.Detector {
	return c.contentTypeDetector.Get(func() contenttype.Detector {
		return contenttype.NewMagicDetector()
	})
}

func (c *Container) PaymentProvider() paymentprovider.Provider {
	return c.paymentProvider.Get(func() paymentprovider.Provider {
		yookassaConfig := yookassa.NewConfig(
			c.Config().Yookassa.AccountID,
			c.Config().Yookassa.SecretKey,
			c.Config().Yookassa.ReturnURL,
		)

		paymentProvider, err := yookassa.New(yookassaConfig)
		if err != nil {
			logger.Fatal(c.Logger(), "failed to create payment provider", err)
		}

		return paymentProvider
	})
}

func (c *Container) TOTPManager() totp.Manager {
	return c.totpManager.Get(func() totp.Manager {
		return totp.New(c.Config().AppName)
	})
}

func (c *Container) EncryptionManager() crypto.EncryptionManager {
	return c.encryptionManager.Get(func() crypto.EncryptionManager {
		encryptionManager, err := crypto.NewAESGCMEncryptionManagerFromBase64(c.Config().MasterKey)
		if err != nil {
			logger.Fatal(c.Logger(), "failed to create encryption manager", err)
		}

		return encryptionManager
	})
}

func (c *Container) TaskFactory() *tasks.Factory {
	return c.taskFactory.Get(func() *tasks.Factory {
		return tasks.NewFactory(c.RedisClient().RDB())
	})
}

func (c *Container) UploadPolicyRegistry() upload.PolicyRegistry {
	return c.uploadPolicyRegistry.Get(func() upload.PolicyRegistry {
		return NewUploadPolicyRegistry()
	})
}

func (c *Container) UploadManager() upload.Manager {
	return c.uploadManager.Get(func() upload.Manager {
		uploadManager := upload.NewManager(
			c.Storage(),
			c.UploadRepo(),
			c.Config().Upload,
			c.ContentTypeDetector(),
			c.UploadPolicyRegistry(),
			c.Logger(),
		)

		return uploadManager
	})
}

func (c *Container) Cache() cache.Cache {
	return c.cache.Get(func() cache.Cache {
		return cache.NewRedisCache(c.RedisClient().RDB(), c.Config().AppName)
	})
}

func (c *Container) TemplateManager() template.Manager {
	return c.templateManager.Get(func() template.Manager {
		templateManager := template.NewManager()
		if err := templateManager.LoadFromFS(templates.FS, "*.gohtml"); err != nil {
			logger.Fatal(c.Logger(), "failed to load templates from FS", err)
		}

		return templateManager
	})
}

func (c *Container) NotificationRegistry() notification.SenderRegistry {
	return c.notificationRegistry.Get(func() notification.SenderRegistry {
		return NewNotificationRegistry(c.Config())
	})
}

func (c *Container) MetricsFactory() *metrics.Factory {
	return c.metricsFactory.Get(func() *metrics.Factory {
		return metrics.New(NewPrometheusRegistry(c))
	})
}

func (c *Container) UserRepo() repository.UserRepository {
	return c.userRepository.Get(func() repository.UserRepository {
		return gormRepo.NewUserRepository(c.DB())
	})
}

func (c *Container) ProductRepo() repository.ProductRepository {
	return c.productRepository.Get(func() repository.ProductRepository {
		return gormRepo.NewProductRepository(c.DB())
	})
}

func (c *Container) CategoryRepo() repository.CategoryRepository {
	return c.categoryRepository.Get(func() repository.CategoryRepository {
		return gormRepo.NewCategoryRepository(c.DB())
	})
}

func (c *Container) OrderRepo() repository.OrderRepository {
	return c.orderRepository.Get(func() repository.OrderRepository {
		return gormRepo.NewOrderRepository(c.DB())
	})
}

func (c *Container) OrderItemRepo() repository.OrderItemRepository {
	return c.orderItemRepository.Get(func() repository.OrderItemRepository {
		return gormRepo.NewOrderItemRepository(c.DB())
	})
}

func (c *Container) UploadRepo() repository.UploadRepository {
	return c.uploadRepository.Get(func() repository.UploadRepository {
		return gormRepo.NewUploadRepository(c.DB())
	})
}

func (c *Container) WishlistRepo() repository.WishlistRepository {
	return c.wishlistRepository.Get(func() repository.WishlistRepository {
		return gormRepo.NewWishlistRepository(c.DB())
	})
}

func (c *Container) WishlistItemRepo() repository.WishlistItemRepository {
	return c.wishlistItemRepository.Get(func() repository.WishlistItemRepository {
		return gormRepo.NewWishlistItemRepository(c.DB())
	})
}

func (c *Container) AddressRepo() repository.AddressRepository {
	return c.addressRepository.Get(func() repository.AddressRepository {
		return gormRepo.NewAddressRepository(c.DB())
	})
}

func (c *Container) UserService() service.UserService {
	return c.userService.Get(func() service.UserService {
		userService := service.NewUserService(
			c.UserRepo(),
			c.TokenManager(),
			c.PasswordHasher(),
			c.TOTPManager(),
			c.EncryptionManager(),
			c.TaskFactory().Notifications(),
			c.Cache(),
			&service.UserEmailConfig{
				EmailConfirmationCodeLength: c.Config().Email.ConfirmationCodeLength,
				EmailConfirmationCodeTTL:    c.Config().Email.ConfirmationCodeTTL,
			},
		)

		return userService
	})
}

func (c *Container) ProductService() service.ProductService {
	return c.productService.Get(func() service.ProductService {
		return service.NewProductService(c.ProductRepo(), c.UploadManager())
	})
}

func (c *Container) CategoryService() service.CategoryService {
	return c.categoryService.Get(func() service.CategoryService {
		return service.NewCategoryService(c.CategoryRepo())
	})
}

func (c *Container) InventoryService() service.InventoryService {
	return c.inventoryService.Get(func() service.InventoryService {
		return service.NewInventoryService(
			c.ProductRepo(),
			c.TxManager(),
		)
	})
}

func (c *Container) OrderService() service.OrderService {
	return c.orderService.Get(func() service.OrderService {
		return service.NewOrderService(
			c.OrderRepo(),
			c.OrderItemRepo(),
			c.AddressRepo(),
			c.TaskFactory().Orders(),
			c.TxManager(),
			c.Config().OrderCancelDelay,
			c.UploadManager(),
			c.InventoryService(),
		)
	})
}

func (c *Container) WishlistService() service.WishlistService {
	return c.wishlistService.Get(func() service.WishlistService {
		return service.NewWishlistService(
			c.WishlistRepo(),
			c.WishlistItemRepo(),
			c.ProductRepo(),
		)
	})
}

func (c *Container) NotificationService() service.NotificationService {
	return c.notificationService.Get(func() service.NotificationService {
		return service.NewNotificationService(
			c.NotificationRegistry(),
			c.TemplateManager(),
		)
	})
}

func (c *Container) PaymentService() service.PaymentService {
	return c.paymentService.Get(func() service.PaymentService {
		return service.NewPaymentService(
			c.PaymentProvider(),
			c.OrderRepo(),
			c.OrderService(),
			c.TxManager(),
		)
	})
}

func (c *Container) AddressService() service.AddressService {
	return c.addressService.Get(func() service.AddressService {
		return service.NewAddressService(c.AddressRepo())
	})
}

func (c *Container) Close() error {
	if err := c.DB().Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}

	if err := c.RedisClient().Close(); err != nil {
		return fmt.Errorf("failed to close redis client: %w", err)
	}

	return nil
}

func sqlDBFromContainer(container *Container) (*sql.DB, error) {
	sqlDB, err := container.DB().GetDB(context.Background()).DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql db from container: %w", err)
	}

	return sqlDB, nil
}
