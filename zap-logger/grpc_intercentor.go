package logger

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func LoggerUnary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (res interface{}, err error) {
		ctxLogger := Get(ctx)

		defer func() {
			if err != nil {
				ctxLogger.Info("Request details", zap.Error(err))
			} else {
				ctxLogger.Info("Request details")
			}
		}()

		res, err = handler(ctx, req)
		return
	}
}

func RecoverUnary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (res interface{}, err error) {
		start := time.Now()
		defer func() {
			dur := zap.Duration("duration", time.Since(start))
			if rec := recover(); rec != nil {
				res = nil
				err = status.New(codes.Internal, "Internal server error.").Err()
				Get(ctx).DPanic("Unhandled panic.", zap.Any("panic", rec), dur)
				return
			}

			if err == nil {
				Get(ctx).Info("Done.", dur)
				return
			}

			if _, ok := status.FromError(err); ok {
				Get(ctx).Warn("Done with gRPC error.", dur, zap.Error(err))
			} else {
				Get(ctx).Error("Done with unknown error.", dur, zap.Error(err))
			}
		}()

		res, err = handler(ctx, req)
		return
	}
}

func LoggerStream(lg LoggerFromContextGetter) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		ctxLogger := lg(ss.Context())

		defer func() {
			if err != nil {
				ctxLogger.Info("Stream request details", zap.Error(err))
			} else {
				ctxLogger.Info("Stream request details")
			}

		}()
		err = handler(srv, ss)
		return
	}
}

func RecoverStream(lg LoggerFromContextGetter) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		ctx := ss.Context()
		start := time.Now()
		defer func() {
			dur := zap.Duration("duration", time.Since(start))

			if rec := recover(); rec != nil {
				lg(ctx).DPanic("Unhandled panic.", zap.Any("panic", rec))
				err = status.New(codes.Internal, "Internal server error.").Err()
				return
			}

			if err == nil {
				lg(ctx).Info("Done.", dur)
				return
			}

			if _, ok := status.FromError(err); ok {
				lg(ctx).Warn("Done with gRPC error.", dur, zap.Error(err))
			} else {
				lg(ctx).Error("Done with unknown error.", dur, zap.Error(err))
			}
		}()

		err = handler(srv, ss)
		return
	}
}

type LoggerFromContextGetter func(ctx context.Context) *zap.Logger
