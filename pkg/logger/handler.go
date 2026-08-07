package logger

import (
	"context"
	"log/slog"
)

type contextKey string

const requestIDKey contextKey = "request_id"

type RequestIDHandler struct {
	slog.Handler
}

func (h *RequestIDHandler) Handle(ctx context.Context, r slog.Record) error {
	if requestID, ok := RequestIDFromContext(ctx); ok {
		r.AddAttrs(slog.String("request_id", requestID))
	}

	return h.Handler.Handle(ctx, r)
}

func (h *RequestIDHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &RequestIDHandler{
		Handler: h.Handler.WithAttrs(attrs),
	}
}

func (h *RequestIDHandler) WithGroup(name string) slog.Handler {
	return &RequestIDHandler{
		Handler: h.Handler.WithGroup(name),
	}
}

func (h *RequestIDHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.Handler.Enabled(ctx, level)
}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

func RequestIDFromContext(ctx context.Context) (string, bool) {
	requestID, ok := ctx.Value(requestIDKey).(string)
	return requestID, ok
}
