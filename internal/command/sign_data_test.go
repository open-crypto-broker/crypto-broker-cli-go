package command

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/otel"
	cryptobrokerclientgo "github.com/open-crypto-broker/crypto-broker-client-go"
)

func BenchmarkSignData_profile_Default_Sequential(b *testing.B) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{AddSource: false}))
	tracerProvider, err := otel.NewTracerProvider(ctx, logger)
	if err != nil {
		b.Fatalf("could not instantiate tracer provider, err: %s", err.Error())
	}
	privateKeyPEM, _ := benchmarkSigningKeys(b)
	lib, err := cryptobrokerclientgo.NewLibrary(ctx)
	if err != nil {
		b.Fatalf("could not instantiate library, err: %s", err.Error())
	}
	b.Cleanup(func() { _ = lib.Close() })
	command, err := NewSignData(ctx, lib, logger, tracerProvider)
	if err != nil {
		b.Fatalf("could not instantiate sign-data command, err: %s", err.Error())
	}

	for b.Loop() {
		if _, err := command.signData(ctx, benchmarkSigningInput, "Default", privateKeyPEM); err != nil {
			b.Fatalf("could not sign data, err: %s", err.Error())
		}
	}
}

func BenchmarkSignData_profile_Default_Parallel(b *testing.B) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{AddSource: false}))
	tracerProvider, err := otel.NewTracerProvider(ctx, logger)
	if err != nil {
		b.Fatalf("could not instantiate tracer provider, err: %s", err.Error())
	}
	privateKeyPEM, _ := benchmarkSigningKeys(b)
	b.RunParallel(func(p *testing.PB) {
		lib, err := cryptobrokerclientgo.NewLibrary(ctx)
		if err != nil {
			b.Fatalf("could not instantiate library, err: %s", err.Error())
		}
		defer func() { _ = lib.Close() }()
		command, err := NewSignData(ctx, lib, logger, tracerProvider)
		if err != nil {
			b.Fatalf("could not instantiate sign-data command, err: %s", err.Error())
		}
		for p.Next() {
			if _, err := command.signData(ctx, benchmarkSigningInput, "Default", privateKeyPEM); err != nil {
				b.Fatalf("could not sign data, err: %s", err.Error())
			}
		}
	})
}