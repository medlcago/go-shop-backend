package tasks

import (
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

type Factory interface {
	Orders() OrderTask
}

type factory struct {
	client *asynq.Client
}

func NewFactory(rdb redis.UniversalClient) Factory {
	client := asynq.NewClientFromRedisClient(rdb)

	return &factory{
		client: client,
	}
}

func (f *factory) Orders() OrderTask {
	return &orderTask{f.client}
}
