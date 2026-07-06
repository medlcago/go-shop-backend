package tasks

import (
	"context"
	"time"

	"github.com/hibiken/asynq"
)

type OrderTask interface {
	EnqueueCancelOrder(ctx context.Context, payload CancelOrderPayload, delay time.Duration) error
}

type orderTask struct {
	client *asynq.Client
}

func (o *orderTask) EnqueueCancelOrder(ctx context.Context, payload CancelOrderPayload, delay time.Duration) error {
	task, err := NewCancelOrderTask(payload, delay)
	if err != nil {
		return err
	}

	_, err = o.client.EnqueueContext(ctx, task)
	return err
}
