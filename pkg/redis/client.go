package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	rdb redis.UniversalClient
}

func New(addr, password string, opts ...Option) (*Client, error) {
	opt := getOption(opts...)

	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           opt.DB,
		DialTimeout:  opt.DialTimeout,
		ReadTimeout:  opt.ReadTimeout,
		WriteTimeout: opt.WriteTimeout,
		PoolSize:     opt.PoolSize,
		MinIdleConns: opt.MinIdleConns,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &Client{
		rdb: rdb,
	}, nil
}

func (c *Client) RDB() redis.UniversalClient {
	return c.rdb
}

func (c *Client) Close() error {
	return c.rdb.Close()
}
