package command

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/otel"
	cryptobrokerclientgo "github.com/open-crypto-broker/crypto-broker-client-go"
)

func BenchmarkVerifyData_profile_Default_Sequential(b *testing.B) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{AddSource: false}))
	tracerProvider, err := otel.NewTracerProvider(ctx, logger)
	if err != nil {
		b.Fatalf("could not instantiate tracer provider, err: %s", err.Error())
	}
	privateKeyPEM, publicKeyPEM := benchmarkSigningKeys(b)
	lib, err := cryptobrokerclientgo.NewLibrary(ctx)
	if err != nil {
		b.Fatalf("could not instantiate library, err: %s", err.Error())
	}
	b.Cleanup(func() { _ = lib.Close() })
	signCommand, err := NewSignData(ctx, lib, logger, tracerProvider)
	if err != nil {
		b.Fatalf("could not instantiate sign-data command, err: %s", err.Error())
	}
	signature, err := signCommand.signData(ctx, benchmarkSigningInput, "Default", privateKeyPEM)
	if err != nil {
		b.Fatalf("could not create benchmark signature, err: %s", err.Error())
	}
	verifyCommand, err := NewVerifyData(ctx, lib, logger, tracerProvider)
	if err != nil {
		b.Fatalf("could not instantiate verify-data command, err: %s", err.Error())
	}
	for b.Loop() {
		valid, err := verifyCommand.verifyData(ctx, benchmarkSigningInput, signature, "Default", publicKeyPEM)
		if err != nil || !valid {
			b.Fatalf("could not verify data: valid=%t err=%v", valid, err)
		}
	}
}

func BenchmarkVerifyData_profile_Default_Parallel(b *testing.B) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{AddSource: false}))
	tracerProvider, err := otel.NewTracerProvider(ctx, logger)
	if err != nil {
		b.Fatalf("could not instantiate tracer provider, err: %s", err.Error())
	}
	privateKeyPEM, publicKeyPEM := benchmarkSigningKeys(b)
	setupLibrary, err := cryptobrokerclientgo.NewLibrary(ctx)
	if err != nil {
		b.Fatalf("could not instantiate library, err: %s", err.Error())
	}
	setupCommand, err := NewSignData(ctx, setupLibrary, logger, tracerProvider)
	if err != nil {
		b.Fatalf("could not instantiate sign-data command, err: %s", err.Error())
	}
	signature, err := setupCommand.signData(ctx, benchmarkSigningInput, "Default", privateKeyPEM)
	_ = setupLibrary.Close()
	if err != nil {
		b.Fatalf("could not create benchmark signature, err: %s", err.Error())
	}
	b.RunParallel(func(p *testing.PB) {
		lib, err := cryptobrokerclientgo.NewLibrary(ctx)
		if err != nil {
			b.Fatalf("could not instantiate library, err: %s", err.Error())
		}
		defer func() { _ = lib.Close() }()
		command, err := NewVerifyData(ctx, lib, logger, tracerProvider)
		if err != nil {
			b.Fatalf("could not instantiate verify-data command, err: %s", err.Error())
		}
		for p.Next() {
			valid, err := command.verifyData(ctx, benchmarkSigningInput, signature, "Default", publicKeyPEM)
			if err != nil || !valid {
				b.Fatalf("could not verify data: valid=%t err=%v", valid, err)
			}
		}
	})
}
