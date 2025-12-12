package testutils

import "log/slog"

var (
	DiscardSlog = slog.New(slog.DiscardHandler)
)
