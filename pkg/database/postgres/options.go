package postgres

import "time"

type Option interface {
	apply(*option)
}

type optionFunc func(*option)

func (f optionFunc) apply(o *option) { f(o) }

type option struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

func defaultOptions() *option {
	return &option{
		MaxOpenConns:    100,
		MaxIdleConns:    50,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 10 * time.Minute,
	}
}

func WithMaxOpenConns(maxOpenConns int) Option {
	return optionFunc(func(o *option) {
		o.MaxOpenConns = maxOpenConns
	})
}

func WithMaxIdleConns(maxIdleConns int) Option {
	return optionFunc(func(o *option) {
		o.MaxIdleConns = maxIdleConns
	})
}

func WithConnMaxLifetime(connMaxLifetime time.Duration) Option {
	return optionFunc(func(o *option) {
		o.ConnMaxLifetime = connMaxLifetime
	})
}

func WithConnMaxIdleTime(connMaxIdleTime time.Duration) Option {
	return optionFunc(func(o *option) {
		o.ConnMaxIdleTime = connMaxIdleTime
	})
}
