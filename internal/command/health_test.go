package command

import (
	"context"
	"io"
	"log"
	"testing"

	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/otel"
	cryptobrokerclientgo "github.com/open-crypto-broker/crypto-broker-client-go"
)

func BenchmarkHealth_Synchronously(b *testing.B) {
	ctx := context.Background()
	logger := log.New(io.Discard, "TEST: ", log.Ldate|log.Lmicroseconds)
	tracerProvider, err := otel.NewTracerProvider(ctx, logger, "crypto-broker-cli-go", "0.0.0")
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

func BenchmarkHealth_Asynchronously(b *testing.B) {
	ctx := context.Background()
	logger := log.New(io.Discard, "TEST: ", log.Ldate|log.Lmicroseconds)
	tracerProvider, err := otel.NewTracerProvider(ctx, logger, "crypto-broker-cli-go", "0.0.0")
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