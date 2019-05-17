package logger

import (
	"context"

	"go.uber.org/zap"
)

type ctxType int

const (
	loggerCtxKey ctxType = iota
)

func Set(ctx context.Context, l *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerCtxKey, l)
}

func Get(ctx context.Context) *zap.Logger {
	return ctx.Value(loggerCtxKey).(*zap.Logger)
}
