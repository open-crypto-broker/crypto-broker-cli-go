package command

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/constant"
	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/otel"
	cryptobrokerclientgo "github.com/open-crypto-broker/crypto-broker-client-go"
)

func TestDecryptDataDecodeTag(t *testing.T) {
	tag, err := decodeHex("tag", "0e8a5ed855c3f96c2c35db398b714ffa")
	if err != nil {
		t.Fatalf("decodeHex() error = %v", err)
	}

	want := []byte{0x0e, 0x8a, 0x5e, 0xd8, 0x55, 0xc3, 0xf9, 0x6c, 0x2c, 0x35, 0xdb, 0x39, 0x8b, 0x71, 0x4f, 0xfa}
	if !bytes.Equal(tag, want) {
		t.Fatalf("tag = %x, want %x", tag, want)
	}
}

func TestDecryptDataRejectsInvalidTag(t *testing.T) {
	if _, err := decodeHex("tag", "not-hex"); err == nil {
		t.Fatal("decodeHex() error = nil, want non-nil")
	}
}

func BenchmarkDecryptData_profile_Default_Sequential(b *testing.B) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{AddSource: false}))
	tracerProvider, err := otel.NewTracerProvider(ctx, logger)
	if err != nil {
		b.Fatalf("could not instantiate tracer provider, err: %s", err.Error())
	}
	lib, err := cryptobrokerclientgo.NewLibrary(ctx)
	if err != nil {
		b.Fatalf("could not instantiate library, err: %s", err.Error())
	}
	b.Cleanup(func() { _ = lib.Close() })

	decryptCmd, err := NewDecryptData(ctx, lib, logger, tracerProvider)
	if err != nil {
		b.Fatalf("could not instantiate decrypt command, err: %s", err.Error())
	}

	for b.Loop() {
		if err := decryptCmd.decryptData(ctx, benchmarkEncryptionCiphertext,
			"Default", constant.KeySourceRaw, benchmarkEncryptionKey, "a83f89b37c90f937b8df5011", benchmarkEncryptionAAD, "a77b42e960d89683140cae283a87466e"); err != nil {
			b.Fatalf("could not run decrypt, err: %s", err.Error())
		}
	}
}

func BenchmarkDecryptData_profile_Default_Parallel(b *testing.B) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{AddSource: false}))
	tracerProvider, err := otel.NewTracerProvider(ctx, logger)
	if err != nil {
		b.Fatalf("could not instantiate tracer provider, err: %s", err.Error())
	}

	b.RunParallel(func(p *testing.PB) {
		lib, err := cryptobrokerclientgo.NewLibrary(ctx)
		if err != nil {
			b.Fatalf("could not instantiate library, err: %s", err.Error())
		}
		defer func() { _ = lib.Close() }()

		decryptCmd, err := NewDecryptData(ctx, lib, logger, tracerProvider)
		if err != nil {
			b.Fatalf("could not instantiate decrypt command, err: %s", err.Error())
		}

		for p.Next() {
			if err := decryptCmd.decryptData(ctx, benchmarkEncryptionCiphertext,
				"Default", constant.KeySourceRaw, benchmarkEncryptionKey, "a83f89b37c90f937b8df5011", benchmarkEncryptionAAD, "a77b42e960d89683140cae283a87466e"); err != nil {
				b.Fatalf("could not run decrypt, err: %s", err.Error())
			}
		}
	})
}
