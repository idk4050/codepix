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

var startFormat = "02 Jan 2006 15:04:05 -0700"

func UnaryLogger(logger logr.Logger) grpc.UnaryServerInterceptor {
	log := logger.WithName("grpc")

	return func(ctx context.Context, req interface{},
		info *grpc.UnaryServerInfo, next grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		reply, err := next(ctx, req)
		duration := time.Since(start)

		status := status.Code(err)

		ip := ""
		peer, _ := peer.FromContext(ctx)
		if peer != nil {
			ip = peer.Addr.String()
		}

		kvs := []any{
			"method", info.FullMethod,
			"status", status.String(),
			"start", start.Format(startFormat),
			"duration", duration.String(),
			"ip", ip,
		}
		if status == codes.OK {
			log.Info("unary call", kvs...)
		} else {
			log.Error(err, "unary call failed", kvs...)
		}
		return reply, err
	}
}

func StreamLogger(logger logr.Logger) grpc.StreamServerInterceptor {
	log := logger.WithName("grpc")

	return func(srv interface{}, stream grpc.ServerStream,
		info *grpc.StreamServerInfo, next grpc.StreamHandler) error {
		start := time.Now()

		ip := ""
		peer, _ := peer.FromContext(stream.Context())
		if peer != nil {
			ip = peer.Addr.String()
		}

		kvs := []any{
			"method", info.FullMethod,
			"start", start.Format(startFormat),
			"ip", ip,
		}
		log.Info("stream started", kvs...)

		err := next(srv, stream)
		duration := time.Since(start)

		status := status.Code(err)

		kvs = []any{
			"method", info.FullMethod,
			"status", status.String(),
			"start", start.Format(startFormat),
			"duration", duration.String(),
			"ip", ip,
		}
		if status == codes.OK {
			log.Info("stream ended", kvs...)
		} else {
			log.Error(err, "stream failed", kvs...)
		}
		return err
	}
}
