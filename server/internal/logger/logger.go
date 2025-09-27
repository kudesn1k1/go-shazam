// logger/logger.go
package logger

import (
	"context"
	"log/slog"
	"os"
)

var defaultLogger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
	Level: slog.LevelInfo,
}))

func FromContext(ctx context.Context) *slog.Logger {
	if correlationID, ok := ctx.Value("correlation_id").(string); ok {
		return defaultLogger.With("correlation_id", correlationID)
	}
	return defaultLogger
}

func Global() *slog.Logger {
	return defaultLogger
}
