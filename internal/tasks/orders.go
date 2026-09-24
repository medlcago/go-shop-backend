package tasks

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type OrderTask interface {
	EnqueueCancelOrder(ctx context.Context, payload CancelOrderPayload, delay time.Duration) error
}

type CancelOrderPayload struct {
	UserID  uuid.UUID `json:"user_id"`
	OrderID uuid.UUID `json:"order_id"`
}

type orderTask struct {
	client *asynq.Client
}

func (o *orderTask) EnqueueCancelOrder(ctx context.Context, payload CancelOrderPayload, delay time.Duration) error {
	task, err := newCancelOrderTask(payload, delay)
	if err != nil {
		return err
	}

	_, err = o.client.EnqueueContext(ctx, task)
	return err
}

func newCancelOrderTask(payload CancelOrderPayload, delay time.Duration) (*asynq.Task, error) {
	data, err := json.Marshal(payload)

	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TypeCancelOrder, data, asynq.ProcessIn(delay)), nil
}
