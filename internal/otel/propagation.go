package otel

import (
	"context"

	otelapi "go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc/metadata"
)

// InjectGRPCTraceContext adds the active trace context to outgoing gRPC
// metadata.
func InjectGRPCTraceContext(ctx context.Context) context.Context {
	md, _ := metadata.FromOutgoingContext(ctx)
	md = md.Copy()

	otelapi.GetTextMapPropagator().Inject(ctx, grpcMetadataCarrier(md))
	return metadata.NewOutgoingContext(ctx, md)
}

type grpcMetadataCarrier metadata.MD

func (c grpcMetadataCarrier) Get(key string) string {
	values := metadata.MD(c).Get(key)
	if len(values) == 0 {
		return ""
	}

	return values[0]
}

func (c grpcMetadataCarrier) Set(key, value string) {
	metadata.MD(c).Set(key, value)
}

func (c grpcMetadataCarrier) Keys() []string {
	keys := make([]string, 0, len(c))
	for key := range c {
		keys = append(keys, key)
	}

	return keys
}

var _ propagation.TextMapCarrier = grpcMetadataCarrier{}
