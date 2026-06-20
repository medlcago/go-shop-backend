package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisCache struct {
	rdb     redis.UniversalClient
	keyFunc func(key string) string
}

func NewRedisCache(rdb redis.UniversalClient, prefix string) *redisCache {
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

func (c *redisCache) Cache(ctx context.Context, key string, value string, ttl time.Duration) (bool, error) {
	key = c.keyFunc(key)
	ok, err := c.rdb.SetNX(ctx, key, value, ttl).Result()
	if err != nil {
		return false, HandleError(err)
	}

	return ok, nil
}

func (c *redisCache) Get(ctx context.Context, key string) (string, error) {
	key = c.keyFunc(key)
	value, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		return "", HandleError(err)
	}

	return value, nil
}

func (c *redisCache) Exists(ctx context.Context, key string) (bool, error) {
	key = c.keyFunc(key)
	value, err := c.rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, HandleError(err)
	}

	return value > 0, nil
}

func (c *redisCache) Delete(ctx context.Context, key string) error {
	key = c.keyFunc(key)
	return HandleError(c.rdb.Del(ctx, key).Err())
}
