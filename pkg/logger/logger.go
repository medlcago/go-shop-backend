package logger

import (
	"log/slog"
	"os"
)

type Env string

const (
	EnvProduction  Env = "production"
	EnvDevelopment Env = "development"
)

func NewSlog(env Env) *slog.Logger {
	var level slog.Level

	switch env {
	case EnvProduction:
		level = slog.LevelInfo
	case EnvDevelopment:
		level = slog.LevelDebug
	default:
		level = slog.LevelInfo
	}

	handler := &RequestIDHandler{
		Handler: slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level:     level,
			AddSource: env == EnvDevelopment,
		}).WithAttrs([]slog.Attr{
			slog.String("environment", string(env)),
		}),
	}

	return slog.New(handler)
}

func Err(err error) slog.Attr {
	return slog.Any("error", err)
}

func Op(op string) slog.Attr {
	return slog.String("op", op)
}

func Fatal(l *slog.Logger, msg string, err error) {
	l.Error(msg, Err(err))
	os.Exit(1)
}
