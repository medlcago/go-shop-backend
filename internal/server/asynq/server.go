package asynq

import (
	"context"
	"go-shop-backend/internal/core"
	"go-shop-backend/internal/tasks"
	taskHandlers "go-shop-backend/internal/tasks/handlers"
	"go-shop-backend/pkg/logger"
	"log/slog"

	"github.com/hibiken/asynq"
)

type Server struct {
	srv    *asynq.Server
	mux    *asynq.ServeMux
	deps   *core.Dependencies
	logger *slog.Logger
}

func NewServer(deps *core.Dependencies) *Server {
	log := deps.Logger.With("server", "asynq")

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: deps.Cfg.Redis.Address, Password: deps.Cfg.Redis.Password},
		asynq.Config{
			Concurrency:     10,
			ShutdownTimeout: deps.Cfg.ShutdownTimeout,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
		},
	)

	mux := asynq.NewServeMux()

	return &Server{
		srv:    srv,
		mux:    mux,
		deps:   deps,
		logger: log,
	}
}

func (s *Server) Start(ctx context.Context) error {
	s.Init()

	s.logger.Info(
		"Asynq server starting",
		slog.String("env", s.deps.Cfg.Environment),
	)

	go func() {
		<-ctx.Done()
		s.logger.Info("Asynq shutdown signal received")
		err := s.Stop(context.Background())
		if err != nil {
			s.logger.Error("s.Stop failed", logger.Err(err))
		}
	}()

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

func (s *Server) Init() {
	orderHandler := taskHandlers.NewOrderTaskHandler(s.deps.OrderService, s.logger)

	s.mux.HandleFunc(tasks.TypeCancelOrder, orderHandler.CancelOrder)
}
