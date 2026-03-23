package worker

import (
	"context"
	"errors"
	"go-shop-backend/config"
	"go-shop-backend/pkg/database"
	"go-shop-backend/pkg/logger"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

type OrderCanceler interface {
	CancelExpiredPending(ctx context.Context, now time.Time, batchSize int) ([]uuid.UUID, error)
}

type ExpiredOrderCanceler struct {
	orderCanceler OrderCanceler
	txManager     database.TxManager
	logger        *slog.Logger

	cfg config.Worker

	started atomic.Bool
	stopCh  chan struct{}
}

func NewExpiredOrderCanceler(
	orderCanceler OrderCanceler,
	txManager database.TxManager,
	logger *slog.Logger,
	cfg config.Worker,
) *ExpiredOrderCanceler {
	logger = logger.With(
		slog.String("worker", "expired-order-canceler"),
		slog.Duration("interval", cfg.Interval),
		slog.Duration("timeout", cfg.Timeout),
		slog.Int("batch_size", cfg.BatchSize),
	)

	return &ExpiredOrderCanceler{
		orderCanceler: orderCanceler,
		txManager:     txManager,
		logger:        logger,
		cfg:           cfg,
		stopCh:        make(chan struct{}),
	}
}

func (w *ExpiredOrderCanceler) Start(ctx context.Context) {
	if !w.started.CompareAndSwap(false, true) {
		w.logger.Warn("worker already started")
		return
	}
	go w.loop(ctx)
}

func (w *ExpiredOrderCanceler) Close() {
	if !w.started.Load() {
		return
	}

	close(w.stopCh)
	w.logger.Info("worker fully stopped")
}

func (w *ExpiredOrderCanceler) loop(ctx context.Context) {
	ticker := time.NewTicker(w.cfg.Interval)
	defer ticker.Stop()

	w.logger.Info("expired order canceler started")

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("worker stopped by context")
			return
		case <-w.stopCh:
			w.logger.Info("worker stopped manually")
			return
		case <-ticker.C:
			w.runOnce(ctx)
		}
	}
}

func (w *ExpiredOrderCanceler) runOnce(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, w.cfg.Timeout)
	defer cancel()

	now := time.Now().UTC()
	var totalCanceled int

	w.logger.Debug("worker tick", slog.Time("now", now))

	for {
		var ids []uuid.UUID

		err := w.txManager.Wrap(ctx, func(txCtx context.Context) error {
			var err error
			ids, err = w.orderCanceler.CancelExpiredPending(txCtx, now, w.cfg.BatchSize)
			return err
		})

		if err != nil {
			if !errors.Is(err, context.Canceled) {
				w.logger.Error(
					"failed to cancel expired orders",
					slog.Int("total_canceled", totalCanceled),
					logger.Err(err),
				)
			}

			if totalCanceled > 0 {
				w.logger.Info(
					"worker stopped with partial results",
					slog.Int("total_canceled", totalCanceled),
					logger.Err(err),
				)
			}
			return
		}

		if len(ids) == 0 {
			if totalCanceled > 0 {
				w.logger.Info(
					"worker completed successfully",
					slog.Int("total_canceled", totalCanceled),
				)
			}
			return
		}

		totalCanceled += len(ids)

		w.logger.Info(
			"canceled orders batch",
			slog.Int("batch_count", len(ids)),
			slog.Int("total_canceled", totalCanceled),
		)

		select {
		case <-ctx.Done():
			w.logger.Info(
				"worker stopped due to context cancellation",
				slog.Int("total_canceled", totalCanceled),
				logger.Err(ctx.Err()),
			)
			return
		case <-time.After(50 * time.Millisecond):
		}
	}
}
