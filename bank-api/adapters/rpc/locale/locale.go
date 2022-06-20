package locale

import (
	"context"

	"google.golang.org/grpc/metadata"
)

type key string

const localeKey key = "locale"

// FromContext returns the locale values stored in ctx, if any.
func FromContext(ctx context.Context) []string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		return md.Get(string(localeKey))
	}
	return []string{}
}
