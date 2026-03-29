package redis

import "time"

type Option interface {
	apply(*option)
}

type option struct {
	DB           int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	PoolSize     int
	MinIdleConns int
}

type optionFn func(*option)

func (f optionFn) apply(o *option) { f(o) }

func WithDB(db int) Option {
	return optionFn(func(c *option) {
		c.DB = db
	})
}

func WithDialTimeout(dialTimeout time.Duration) Option {
	return optionFn(func(o *option) {
		o.DialTimeout = dialTimeout
	})
}

func WithReadTimeout(readTimeout time.Duration) Option {
	return optionFn(func(o *option) {
		o.ReadTimeout = readTimeout
	})
}

func WithWriteTimeout(writeTimeout time.Duration) Option {
	return optionFn(func(o *option) {
		o.WriteTimeout = writeTimeout
	})
}

func WithPoolSize(poolSize int) Option {
	return optionFn(func(o *option) {
		o.PoolSize = poolSize
	})
}

func WithMinIdleConns(minIdleConns int) Option {
	return optionFn(func(o *option) {
		o.MinIdleConns = minIdleConns
	})
}

func getOption(opts ...Option) option {
	opt := option{
		DB:           0,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 5,
	}

	for _, o := range opts {
		o.apply(&opt)
	}

	return opt
}
