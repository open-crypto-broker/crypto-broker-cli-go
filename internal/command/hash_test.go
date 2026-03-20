package command

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/otel"
	cryptobrokerclientgo "github.com/open-crypto-broker/crypto-broker-client-go"
)

func BenchmarkHash_profile_Default_Sequential(b *testing.B) {
	ctx := context.Background()
	logger := slog.New(
		slog.NewTextHandler(
			io.Discard, &slog.HandlerOptions{
				AddSource: false,
			},
		),
	)
	tracerProvider, err := otel.NewTracerProvider(ctx, logger)
	if err != nil {
		b.Fatalf("could not instantiate tracer provider, err: %s", err.Error())
	}
	lib, err := cryptobrokerclientgo.NewLibrary(ctx)
	if err != nil {
		b.Fatalf("could not instantiate library, err: %s", err.Error())
	}

	hashCmd, err := NewHash(ctx, lib, logger, tracerProvider)
	if err != nil {
		b.Fatalf("could not instantiate hash, err: %s", err.Error())
	}

	payload := cryptobrokerclientgo.HashDataPayload{
		Profile: "Default",
		Input:   []byte("Hello world"),
		Metadata: &cryptobrokerclientgo.Metadata{
			TraceContext: &cryptobrokerclientgo.TraceContext{
				CorrelationId: uuid.New().String(),
			},
		},
	}
	for b.Loop() {
		err := hashCmd.hashBytes(ctx, payload)
		if err != nil {
			b.Fatalf("could not run hash, err: %s", err.Error())
		}
	}
}

func BenchmarkHash_profile_Default_Parallel(b *testing.B) {
	ctx := context.Background()
	logger := slog.New(
		slog.NewTextHandler(
			io.Discard, &slog.HandlerOptions{
				AddSource: false,
			},
		),
	)
	tracerProvider, err := otel.NewTracerProvider(ctx, logger)
	if err != nil {
		b.Fatalf("could not instantiate tracer provider, err: %s", err.Error())
	}

	b.RunParallel(func(p *testing.PB) {
		lib, err := cryptobrokerclientgo.NewLibrary(ctx)
		if err != nil {
			b.Fatalf("could not instantiate library, err: %s", err.Error())
		}

		hashCmd, err := NewHash(ctx, lib, logger, tracerProvider)
		if err != nil {
			b.Fatalf("could not instantiate hash, err: %s", err.Error())
		}

		payload := cryptobrokerclientgo.HashDataPayload{
			Profile: "Default",
			Input:   []byte("Hello world"),
			Metadata: &cryptobrokerclientgo.Metadata{
				TraceContext: &cryptobrokerclientgo.TraceContext{
					CorrelationId: uuid.New().String(),
				},
			},
		}

		for p.Next() {
			err := hashCmd.hashBytes(ctx, payload)
			if err != nil {
				b.Fatalf("could not run hash, err: %s", err.Error())
			}
		}
	})
}
