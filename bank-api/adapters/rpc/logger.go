package rpc

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

const startFormat = "02 Jan 2006 15:04:05 -0700"

func UnaryLogger(logger logr.Logger) grpc.UnaryServerInterceptor {
	logger = logger.WithName("grpc")

	return func(ctx context.Context, req interface{},
		info *grpc.UnaryServerInfo, next grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		reply, err := next(ctx, req)
		duration := time.Since(start)

		status, _ := status.FromError(err)
		code := status.Code()

		kvs := []any{
			"method", info.FullMethod,
			"status", code.String(),
			"start", start.Format(startFormat),
			"duration", duration.String(),
			"ip", getIP(ctx),
		}
		if code == codes.OK {
			logger.Info("unary call", kvs...)
		} else {
			kvs = append(kvs, "details", status.Details())
			logger.Error(err, "unary call failed", kvs...)
		}
		return reply, err
	}
}

func StreamLogger(logger logr.Logger) grpc.StreamServerInterceptor {
	logger = logger.WithName("grpc")

	return func(srv interface{}, stream grpc.ServerStream,
		info *grpc.StreamServerInfo, next grpc.StreamHandler) error {
		start := time.Now()

		ip := getIP(stream.Context())

		kvs := []any{
			"method", info.FullMethod,
			"start", start.Format(startFormat),
			"ip", ip,
		}
		logger.Info("stream started", kvs...)

		err := next(srv, stream)
		duration := time.Since(start)

		status, _ := status.FromError(err)
		code := status.Code()

		kvs = []any{
			"method", info.FullMethod,
			"status", code.String(),
			"start", start.Format(startFormat),
			"duration", duration.String(),
			"ip", ip,
		}
		if code == codes.OK {
			logger.Info("stream ended", kvs...)
		} else {
			kvs = append(kvs, "details", status.Details())
			logger.Error(err, "stream failed", kvs...)
		}
		return err
	}
}

func getIP(ctx context.Context) string {
	ip := ""
	peer, _ := peer.FromContext(ctx)
	if peer != nil {
		ip = peer.Addr.String()
	}
	return ip
}
