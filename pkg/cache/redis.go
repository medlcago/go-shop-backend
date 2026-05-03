package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisCache struct {
	rdb     *redis.Client
	keyFunc func(key string) string
}

func NewRedisCache(rdb *redis.Client, prefix string) *redisCache {
	return &redisCache{
		rdb: rdb,
		keyFunc: func(key string) string {
			return fmt.Sprintf("%s_%s", prefix, key)
		},
	}
}

func (c *redisCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	key = c.keyFunc(key)
	return HandleError(c.rdb.Set(ctx, key, value, ttl).Err())
}

func (c *redisCache) Get(ctx context.Context, key string) (string, error) {
	key = c.keyFunc(key)
	value, err := c.rdb.Get(ctx, key).Result()

	if err != nil {
		return "", HandleError(err)
	}

	return value, nil
}

func (c *redisCache) Exists(ctx context.Context, key string) error {
	key = c.keyFunc(key)
	value, err := c.rdb.Exists(ctx, key).Result()
	if err != nil {
		return HandleError(err)
	}

	if value == 0 {
		return ErrNotFound
	}

	return nil
}

func (c *redisCache) Delete(ctx context.Context, key string) error {
	key = c.keyFunc(key)
	return HandleError(c.rdb.Del(ctx, key).Err())
}
