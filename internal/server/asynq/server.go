package asynq

import (
	"context"
	"go-shop-backend/internal/core"
	"go-shop-backend/internal/tasks"
	taskHandlers "go-shop-backend/internal/tasks/handlers"
	"log/slog"

	"github.com/hibiken/asynq"
)

type Server struct {
	srv       *asynq.Server
	mux       *asynq.ServeMux
	container *core.Container
	logger    *slog.Logger
}

func NewServer(container *core.Container) *Server {
	log := container.Logger().With("server", "asynq")

	srv := asynq.NewServerFromRedisClient(
		container.RedisClient().RDB(),
		asynq.Config{
			Concurrency:     10,
			ShutdownTimeout: container.Config().ShutdownTimeout,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			Logger: newLogger(log),
		})

	mux := asynq.NewServeMux()

	return &Server{
		srv:       srv,
		mux:       mux,
		container: container,
		logger:    log,
	}
}

func (s *Server) Start(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	s.Init()

	s.logger.Info("Asynq server starting")

	return s.srv.Run(s.mux)
}

func (s *Server) Stop(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	s.logger.Info("Stopping Asynq server")
	s.srv.Shutdown()
	return nil
}

func (s *Server) Name() string {
	return "asynq"
}

func (s *Server) Init() {
	orderHandler := taskHandlers.NewOrderTaskHandler(s.container.OrderService(), s.logger)
	s.mux.HandleFunc(tasks.TypeCancelOrder, orderHandler.CancelOrder)

	notificationHandler := taskHandlers.NewNotificationTaskHandler(s.container.NotificationService(), s.logger)
	s.mux.HandleFunc(tasks.TypeSendEmailConfirmationCode, notificationHandler.SendEmailConfirmationCode)
}
