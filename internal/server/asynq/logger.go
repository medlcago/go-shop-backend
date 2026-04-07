package asynq

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/hibiken/asynq"
)

var _ asynq.Logger = (*asynqLogger)(nil)

type asynqLogger struct {
	logger *slog.Logger
}

func newLogger(logger *slog.Logger) *asynqLogger {
	return &asynqLogger{logger: logger}
}

func (l *asynqLogger) formatArgs(args ...interface{}) string {
	return fmt.Sprint(args...)
}

func (l *asynqLogger) Debug(args ...interface{}) {
	l.logger.Debug(l.formatArgs(args...))
}

func (l *asynqLogger) Info(args ...interface{}) {
	l.logger.Info(l.formatArgs(args...))
}

func (l *asynqLogger) Warn(args ...interface{}) {
	l.logger.Warn(l.formatArgs(args...))
}

func (l *asynqLogger) Error(args ...interface{}) {
	l.logger.Error(l.formatArgs(args...))
}

func (l *asynqLogger) Fatal(args ...interface{}) {
	l.logger.Error(l.formatArgs(args...))
	os.Exit(1)
}
