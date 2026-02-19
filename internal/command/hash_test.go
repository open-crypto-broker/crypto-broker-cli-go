package command

import (
	"context"
	"io"
	"log"
	"testing"

	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/otel"
	cryptobrokerclientgo "github.com/open-crypto-broker/crypto-broker-client-go"
)

func BenchmarkHash_profile_Default_Synchronously(b *testing.B) {
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

	hashCmd, err := NewHash(ctx, lib, logger, tracerProvider)
	if err != nil {
		b.Fatalf("could not instantiate hash, err: %s", err.Error())
	}

	payload := cryptobrokerclientgo.HashDataPayload{
		Profile: "Default",
		Input:   []byte("Hello world"),
	}
	for b.Loop() {
		err := hashCmd.hashBytes(ctx, payload)
		if err != nil {
			b.Fatalf("could not run hash, err: %s", err.Error())
		}
	}
}

func BenchmarkHash_profile_Default_Asynchronously(b *testing.B) {
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

		hashCmd, err := NewHash(ctx, lib, logger, tracerProvider)
		if err != nil {
			b.Fatalf("could not instantiate hash, err: %s", err.Error())
		}

		payload := cryptobrokerclientgo.HashDataPayload{
			Profile: "Default",
			Input:   []byte("Hello world"),
		}

		for p.Next() {
			err := hashCmd.hashBytes(ctx, payload)
			if err != nil {
				b.Fatalf("could not run hash, err: %s", err.Error())
			}
		}
	})
}
