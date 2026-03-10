package command

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/otel"
	cryptobrokerclientgo "github.com/open-crypto-broker/crypto-broker-client-go"
)

func BenchmarkHealth_Sequential(b *testing.B) {
	ctx := context.Background()
	logger := slog.New(
		slog.NewTextHandler(
			io.Discard, &slog.HandlerOptions{
				AddSource: false,
			},
		),
	)
	tracerProvider, err := otel.NewTracerProvider(ctx, logger, "", "")
	if err != nil {
		b.Fatalf("could not instantiate tracer provider, err: %s", err.Error())
	}
	lib, err := cryptobrokerclientgo.NewLibrary(ctx)
	if err != nil {
		b.Fatalf("could not instantiate library, err: %s", err.Error())
	}
	healthCmd, err := NewHealth(ctx, lib, logger, tracerProvider)
	if err != nil {
		b.Fatalf("could not instantiate health, err: %s", err.Error())
	}

	for b.Loop() {
		err := healthCmd.checkHealth(ctx)
		if err != nil {
			b.Fatalf("could not run health, err: %s", err.Error())
		}
	}
}

func BenchmarkHealth_Parallel(b *testing.B) {
	ctx := context.Background()
	logger := slog.New(
		slog.NewTextHandler(
			io.Discard, &slog.HandlerOptions{
				AddSource: false,
			},
		),
	)
	tracerProvider, err := otel.NewTracerProvider(ctx, logger, "", "")
	if err != nil {
		b.Fatalf("could not instantiate tracer provider, err: %s", err.Error())
	}
	b.RunParallel(func(p *testing.PB) {
		lib, err := cryptobrokerclientgo.NewLibrary(ctx)
		if err != nil {
			b.Fatalf("could not instantiate library, err: %s", err.Error())
		}
		healthCmd, err := NewHealth(ctx, lib, logger, tracerProvider)
		if err != nil {
			b.Fatalf("could not instantiate health, err: %s", err.Error())
		}
		for p.Next() {
			err := healthCmd.checkHealth(ctx)
			if err != nil {
				b.Fatalf("could not run health, err: %s", err.Error())
			}
		}
	})
}
