package cache

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type item struct {
	value     string
	expiresAt time.Time
}

type inMemoryCache struct {
	mu      sync.Mutex
	items   map[string]item
	keyFunc func(key string) string
	now     func() time.Time
}

func NewInMemoryCache(prefix string) *inMemoryCache {
	return &inMemoryCache{
		items: make(map[string]item),
		keyFunc: func(key string) string {
			return fmt.Sprintf("%s_%s", prefix, key)
		},
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func (c *inMemoryCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = item{
		value:     value,
		expiresAt: c.now().Add(ttl),
	}

	return nil
}

func (c *inMemoryCache) Get(ctx context.Context, key string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	it, ok := c.items[key]
	if !ok {
		return "", ErrNotFound
	}

	if !it.expiresAt.IsZero() && c.now().After(it.expiresAt) {
		delete(c.items, key)
		return "", ErrNotFound
	}

	return it.value, nil
}

func (c *inMemoryCache) Exists(ctx context.Context, key string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	_, err := c.Get(ctx, key)
	return err
}

func (c *inMemoryCache) Delete(ctx context.Context, key string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
	return nil
}
