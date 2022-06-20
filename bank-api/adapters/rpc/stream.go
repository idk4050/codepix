package rpc

import (
	"context"

	"google.golang.org/grpc"
)

type StreamWithCtx struct {
	grpc.ServerStream
	Ctx context.Context
}

func (h *StreamWithCtx) Context() context.Context {
	return h.Ctx
}
