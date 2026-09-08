package cache

import (
	"errors"

	"github.com/redis/go-redis/v9"
)

var (
	ErrCacheMiss = errors.New("cache: key not found")
)

func HandleError(err error) error {
	switch {
	case errors.Is(err, redis.Nil):
		return ErrCacheMiss
	case err != nil:
		return err
	default:
		return nil
	}
}
