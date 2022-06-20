package rpc

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func UnaryPanicHandler(logger logr.Logger) grpc.UnaryServerInterceptor {
	logger = logger.WithName("grpc.panic")

	return func(ctx context.Context, req interface{},
		info *grpc.UnaryServerInfo, next grpc.UnaryHandler) (reply interface{}, err error) {
		start := time.Now()

		panicked := true
		defer func() {
			if panicked {
				duration := time.Since(start)

				switch r := recover().(type) {
				case error:
					err = fmt.Errorf("internal error: %w", r)
				default:
					err = fmt.Errorf("internal error: %v", r)
				}
				kvs := []any{
					"method", info.FullMethod,
					"start", start.Format(startFormat),
					"duration", duration.String(),
					"ip", getIP(ctx),
				}
				logger.Error(err, "unary call panic", kvs...)
				err = status.Error(codes.Internal, err.Error())
			}
		}()

		reply, err = next(ctx, req)
		panicked = false
		return
	}
}

func StreamPanicHandler(logger logr.Logger) grpc.StreamServerInterceptor {
	logger = logger.WithName("grpc.panic")

	return func(srv interface{}, stream grpc.ServerStream,
		info *grpc.StreamServerInfo, next grpc.StreamHandler) (err error) {
		start := time.Now()

		panicked := true
		defer func() {
			if panicked {
				duration := time.Since(start)

				switch r := recover().(type) {
				case error:
					err = fmt.Errorf("internal error: %w", r)
				default:
					err = fmt.Errorf("internal error: %v", r)
				}
				kvs := []any{
					"method", info.FullMethod,
					"start", start.Format(startFormat),
					"duration", duration.String(),
					"ip", getIP(stream.Context()),
				}
				logger.Error(err, "stream panic", kvs...)
				err = status.Error(codes.Internal, err.Error())
			}
		}()

		err = next(srv, stream)
		panicked = false
		return
	}
}
