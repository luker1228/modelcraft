package middleware

import (
	"context"

	"go.uber.org/zap"
)

type zapLoggerKey struct{}

// WithLogger stores a *zap.Logger in the context.
func WithLogger(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, zapLoggerKey{}, logger)
}

// LoggerFromCtx retrieves the *zap.Logger from the context.
// Falls back to the global zap.L() when none is stored.
func LoggerFromCtx(ctx context.Context) *zap.Logger {
	if l, ok := ctx.Value(zapLoggerKey{}).(*zap.Logger); ok && l != nil {
		return l
	}
	return zap.L()
}
