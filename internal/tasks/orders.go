package tasks

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type OrderTask interface {
	EnqueueCancelOrder(ctx context.Context, userID uuid.UUID, orderID uuid.UUID, delay time.Duration) error
}

type orderTask struct {
	client *asynq.Client
}

func (o *orderTask) EnqueueCancelOrder(ctx context.Context, userID uuid.UUID, orderID uuid.UUID, delay time.Duration) error {
	task, err := NewCancelOrderTask(userID, orderID, delay)
	if err != nil {
		return err
	}

	_, err = o.client.EnqueueContext(ctx, task)
	return err
}
