package middleware

import (
	"go-shop-backend/internal/metrics"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
)

func Prometheus(httpMetrics metrics.HTTPMetrics) fiber.Handler {
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

		route := ctx.Route().Path
		duration := time.Since(start)
		status := ctx.Response().StatusCode()
		method := ctx.Method()

		httpMetrics.RequestTotal(method, route, status)
		httpMetrics.RequestDuration(method, route, duration)

		return err
	}
}
