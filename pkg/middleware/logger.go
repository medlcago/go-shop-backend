package middleware

import (
	"log/slog"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
)

func Logger(log *slog.Logger) fiber.Handler {
	logger := log.With(slog.String("middleware", "logger"))

	var (
		once       sync.Once
		errHandler fiber.ErrorHandler
	)

	return func(ctx fiber.Ctx) error {
		once.Do(func() {
			errHandler = ctx.App().ErrorHandler
		})

		start := time.Now()

		err := ctx.Next()
		if err != nil {
			if err := errHandler(ctx, err); err != nil {
				_ = ctx.SendStatus(fiber.StatusInternalServerError)
			}
		}

		entry := logger.With(
			slog.String("method", ctx.Method()),
			slog.String("path", ctx.Path()),
			slog.String("remote_addr", ctx.IP()),
			slog.String("user_agent", ctx.Get("User-Agent")),
		)

		entry.Info("request completed",
			slog.Int("bytes", len(ctx.Response().Body())),
			slog.Int("status_code", ctx.Response().StatusCode()),
			slog.Duration("duration", time.Since(start)),
		)

		return nil
	}
}
