package core

import (
	"context"
	"fmt"
	"go-shop-backend/config"
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
)

type Container struct {
	cfg                  *config.Config
	logger               *slog.Logger
	validator            validator.Validator
	hasher               hasher.Hasher
	tokenManager         token.Manager
	totpManager          totp.Manager
	encryptionManager    crypto.EncryptionManager
	taskFactory          tasks.Factory
	paymentProvider      paymentprovider.Provider
	db                   database.DB
	txManager            database.TxManager
	storage              storage.Storage
	redisClient          *redis.Client
	contentTypeDetector  contenttype.Detector
	uploadPolicyProvider upload.PolicyProvider
	uploadManager        upload.Manager
	cache                cache.Cache
	templateManager      template.Manager
	notificationRegistry notification.SenderRegistry

	// repositories
	userRepository         repository.UserRepository
	productRepository      repository.ProductRepository
	categoryRepository     repository.CategoryRepository
	uploadRepository       repository.UploadRepository
	orderRepository        repository.OrderRepository
	orderItemRepository    repository.OrderItemRepository
	wishlistRepository     repository.WishlistRepository
	wishlistItemRepository repository.WishlistItemRepository

	// services
	userService         service.UserService
	productService      service.ProductService
	categoryService     service.CategoryService
	orderService        service.OrderService
	wishlistService     service.WishlistService
	notificationService service.NotificationService
	inventoryService    service.InventoryService
	paymentService      service.PaymentService
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
	if c.logger == nil {
		c.logger = logger.NewSlog(logger.Env(c.Config().Environment))
		slog.SetDefault(c.logger)
	}

	return c.logger
}

func (c *Container) Validator() validator.Validator {
	if c.validator == nil {
		c.validator = NewValidator()
	}

	return c.validator
}

func (c *Container) DB() database.DB {
	if c.db != nil {
		return c.db
	}

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

	c.db = db
	return db
}

func (c *Container) TxManager() database.TxManager {
	if c.txManager == nil {
		c.txManager = database.NewGormTxManager(c.DB().GetDB(context.Background()))
	}

	return c.txManager
}

func (c *Container) RedisClient() *redis.Client {
	if c.redisClient != nil {
		return c.redisClient
	}

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

	c.redisClient = redisClient
	return redisClient
}

func (c *Container) Storage() storage.Storage {
	if c.storage != nil {
		return c.storage
	}

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

	c.storage = minioStorage
	return minioStorage
}

func (c *Container) PasswordHasher() hasher.Hasher {
	if c.hasher == nil {
		c.hasher = hasher.NewArgon2ID()
	}

	return c.hasher
}

func (c *Container) TokenManager() token.Manager {
	if c.tokenManager == nil {
		c.tokenManager = token.NewJWT(
			c.Config().AuthSecret,
			c.Config().AccessTokenExpiredTime,
			c.Config().RefreshTokenExpiredTime,
			c.Config().PartialTokenExpiredTime,
		)
	}

	return c.tokenManager
}

func (c *Container) ContentTypeDetector() contenttype.Detector {
	if c.contentTypeDetector == nil {
		c.contentTypeDetector = contenttype.NewMagicDetector()
	}

	return c.contentTypeDetector
}

func (c *Container) PaymentProvider() paymentprovider.Provider {
	if c.paymentProvider != nil {
		return c.paymentProvider
	}

	paymentProvider, err := yookassa.New(
		yookassa.NewConfig(c.Config().Yookassa.AccountID, c.Config().Yookassa.SecretKey, c.Config().Yookassa.ReturnURL),
	)
	if err != nil {
		logger.Fatal(c.Logger(), "failed to create payment provider", err)
	}

	c.paymentProvider = paymentProvider
	return paymentProvider
}

func (c *Container) TOTPManager() totp.Manager {
	if c.totpManager == nil {
		c.totpManager = totp.New(c.Config().AppName)
	}

	return c.totpManager
}

func (c *Container) EncryptionManager() crypto.EncryptionManager {
	if c.encryptionManager != nil {
		return c.encryptionManager
	}

	encryptionManager, err := crypto.NewAESGCMEncryptionManagerFromBase64(c.Config().MasterKey)
	if err != nil {
		logger.Fatal(c.Logger(), "failed to create encryption manager", err)
	}

	c.encryptionManager = encryptionManager
	return c.encryptionManager
}

func (c *Container) TaskFactory() tasks.Factory {
	if c.taskFactory == nil {
		c.taskFactory = tasks.NewFactory(c.RedisClient().RDB())
	}

	return c.taskFactory
}

func (c *Container) UploadPolicyProvider() upload.PolicyProvider {
	if c.uploadPolicyProvider != nil {
		return c.uploadPolicyProvider
	}

	uploadPolicyProvider, err := NewUploadPolicyProvider()
	if err != nil {
		logger.Fatal(c.Logger(), "failed to create upload policy provider", err)
	}

	c.uploadPolicyProvider = uploadPolicyProvider
	return c.uploadPolicyProvider
}

func (c *Container) UploadManager() upload.Manager {
	if c.uploadManager == nil {
		c.uploadManager = upload.NewManager(
			c.Storage(),
			c.UploadRepo(),
			c.Config().Upload,
			c.ContentTypeDetector(),
			c.UploadPolicyProvider(),
			c.Logger(),
		)
	}

	return c.uploadManager
}

func (c *Container) Cache() cache.Cache {
	if c.cache == nil {
		c.cache = cache.NewRedisCache(c.RedisClient().RDB(), c.Config().AppName)
	}

	return c.cache
}

func (c *Container) TemplateManager() template.Manager {
	if c.templateManager != nil {
		return c.templateManager
	}

	templateManager := template.NewManager()
	if err := templateManager.LoadFromFS(templates.FS, "*.gohtml"); err != nil {
		logger.Fatal(c.Logger(), "failed to load templates from FS", err)
	}

	c.templateManager = templateManager
	return c.templateManager
}

func (c *Container) NotificationRegistry() notification.SenderRegistry {
	if c.notificationRegistry == nil {
		c.notificationRegistry = NewNotificationRegistry(c.Config())
	}

	return c.notificationRegistry
}

func (c *Container) UserRepo() repository.UserRepository {
	if c.userRepository == nil {
		c.userRepository = gormRepo.NewUserRepository(c.DB())
	}

	return c.userRepository
}

func (c *Container) ProductRepo() repository.ProductRepository {
	if c.productRepository == nil {
		c.productRepository = gormRepo.NewProductRepository(c.DB())
	}

	return c.productRepository
}

func (c *Container) CategoryRepo() repository.CategoryRepository {
	if c.categoryRepository == nil {
		c.categoryRepository = gormRepo.NewCategoryRepository(c.DB())
	}

	return c.categoryRepository
}

func (c *Container) OrderRepo() repository.OrderRepository {
	if c.orderRepository == nil {
		c.orderRepository = gormRepo.NewOrderRepository(c.DB())
	}

	return c.orderRepository
}

func (c *Container) OrderItemRepo() repository.OrderItemRepository {
	if c.orderItemRepository == nil {
		c.orderItemRepository = gormRepo.NewOrderItemRepository(c.DB())
	}

	return c.orderItemRepository
}

func (c *Container) UploadRepo() repository.UploadRepository {
	if c.uploadRepository == nil {
		c.uploadRepository = gormRepo.NewUploadRepository(c.DB())
	}

	return c.uploadRepository
}

func (c *Container) WishlistRepo() repository.WishlistRepository {
	if c.wishlistRepository == nil {
		c.wishlistRepository = gormRepo.NewWishlistRepository(c.DB())
	}

	return c.wishlistRepository
}

func (c *Container) WishlistItemRepo() repository.WishlistItemRepository {
	if c.wishlistItemRepository == nil {
		c.wishlistItemRepository = gormRepo.NewWishlistItemRepository(c.DB())
	}

	return c.wishlistItemRepository
}

func (c *Container) UserService() service.UserService {
	if c.userService == nil {
		c.userService = service.NewUserService(
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
	}

	return c.userService
}

func (c *Container) ProductService() service.ProductService {
	if c.productService == nil {
		c.productService = service.NewProductService(c.ProductRepo(), c.UploadManager())
	}

	return c.productService
}

func (c *Container) CategoryService() service.CategoryService {
	if c.categoryService == nil {
		c.categoryService = service.NewCategoryService(c.CategoryRepo())
	}

	return c.categoryService
}

func (c *Container) InventoryService() service.InventoryService {
	if c.inventoryService == nil {
		c.inventoryService = service.NewInventoryService(
			c.ProductRepo(),
			c.TxManager(),
		)
	}

	return c.inventoryService
}

func (c *Container) OrderService() service.OrderService {
	if c.orderService == nil {
		c.orderService = service.NewOrderService(
			c.OrderRepo(),
			c.OrderItemRepo(),
			c.ProductRepo(),
			c.TaskFactory().Orders(),
			c.TxManager(),
			c.Config().OrderCancelDelay,
			c.UploadManager(),
			c.InventoryService(),
		)
	}

	return c.orderService
}

func (c *Container) WishlistService() service.WishlistService {
	if c.wishlistService == nil {
		c.wishlistService = service.NewWishlistService(
			c.WishlistRepo(),
			c.WishlistItemRepo(),
			c.ProductRepo(),
		)
	}

	return c.wishlistService
}

func (c *Container) NotificationService() service.NotificationService {
	if c.notificationService == nil {
		c.notificationService = service.NewNotificationService(
			c.NotificationRegistry(),
			c.TemplateManager(),
		)
	}

	return c.notificationService
}

func (c *Container) PaymentService() service.PaymentService {
	if c.paymentService == nil {
		c.paymentService = service.NewPaymentService(
			c.PaymentProvider(),
			c.OrderRepo(),
			c.OrderService(),
			c.TxManager(),
		)
	}

	return c.paymentService
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
