package database

import "time"

type Option interface {
	apply(*option)
}

type option struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

type optionFn func(*option)

func (f optionFn) apply(o *option) { f(o) }

func WithMaxOpenConns(maxOpenConns int) Option {
	return optionFn(func(o *option) {
		o.MaxOpenConns = maxOpenConns
	})
}

func WithMaxIdleConns(maxIdleConns int) Option {
	return optionFn(func(o *option) {
		o.MaxIdleConns = maxIdleConns
	})
}

func WithConnMaxLifetime(connMaxLifetime time.Duration) Option {
	return optionFn(func(o *option) {
		o.ConnMaxLifetime = connMaxLifetime
	})
}

func WithConnMaxIdleTime(connMaxIdleTime time.Duration) Option {
	return optionFn(func(o *option) {
		o.ConnMaxIdleTime = connMaxIdleTime
	})
}

func getOption(opts ...Option) option {
	opt := option{
		MaxOpenConns:    100,
		MaxIdleConns:    50,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 10 * time.Minute,
	}

	for _, o := range opts {
		o.apply(&opt)
	}

	return opt
}
