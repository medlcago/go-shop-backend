package tasks

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const (
	TypeCancelOrder = "cancel:order"
)

type CancelOrderPayload struct {
	UserID  uuid.UUID `json:"user_id"`
	OrderID uuid.UUID `json:"order_id"`
}

func NewCancelOrderTask(userID, orderID uuid.UUID, delay time.Duration) (*asynq.Task, error) {
	payload, err := json.Marshal(CancelOrderPayload{
		UserID:  userID,
		OrderID: orderID,
	})

	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TypeCancelOrder, payload, asynq.ProcessIn(delay)), nil
}
