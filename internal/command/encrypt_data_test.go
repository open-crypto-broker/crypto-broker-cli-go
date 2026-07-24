package command

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"io"
	"log/slog"
	"sync/atomic"
	"testing"

	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/constant"
	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/otel"
	cryptobrokerclientgo "github.com/open-crypto-broker/crypto-broker-client-go"
)

const (
	benchmarkEncryptionKey = "67e1befda6538f162a308bb37b922441bed6e27039a26a48ad8e11c568c4cda0"
	benchmarkEncryptionAAD = "36e02c2a81c60eca849739d52dea95f7"
)

var benchmarkEncryptionNonceCounter atomic.Uint64

var (
	benchmarkEncryptionPlaintext  = []byte("Welcome CryptoBroker")
	benchmarkEncryptionCiphertext = []byte{0x71, 0x41, 0x6b, 0x87, 0x6f, 0xb0, 0xd6, 0x5c, 0x48, 0x4e, 0xc2, 0x01, 0x06, 0xaf, 0x15, 0xa3, 0x64, 0x54, 0x74, 0x3b}
)

func TestParseEncryptionInputs_RawKey(t *testing.T) {
	keySource, nonce, aad, err := parseEncryptionInputs(
		constant.KeySourceRaw,
		"1619159426d9ac45243d3da9eed51899",
		"08edd4dd6bfd0d69275ef2d0",
		"1af8fbdc64693a719f26baa04f0ce8be",
	)
	if err != nil {
		t.Fatalf("parseEncryptionInputs() error = %v", err)
	}

	if !bytes.Equal(keySource.RawKey, []byte{0x16, 0x19, 0x15, 0x94, 0x26, 0xd9, 0xac, 0x45, 0x24, 0x3d, 0x3d, 0xa9, 0xee, 0xd5, 0x18, 0x99}) {
		t.Fatalf("RawKey = %x, want fixture key", keySource.RawKey)
	}
	if keySource.KeyID != "" {
		t.Fatalf("KeyID = %q, want empty", keySource.KeyID)
	}
	if !bytes.Equal(nonce, []byte{0x08, 0xed, 0xd4, 0xdd, 0x6b, 0xfd, 0x0d, 0x69, 0x27, 0x5e, 0xf2, 0xd0}) {
		t.Fatalf("nonce = %x, want fixture nonce", nonce)
	}
	if !bytes.Equal(aad, []byte{0x1a, 0xf8, 0xfb, 0xdc, 0x64, 0x69, 0x3a, 0x71, 0x9f, 0x26, 0xba, 0xa0, 0x4f, 0x0c, 0xe8, 0xbe}) {
		t.Fatalf("aad = %x, want fixture AAD", aad)
	}
}

func TestParseEncryptionInputs_KeyID(t *testing.T) {
	keySource, nonce, aad, err := parseEncryptionInputs(constant.KeySourceKeyID, "managed-key", "00", "")
	if err != nil {
		t.Fatalf("parseEncryptionInputs() error = %v", err)
	}

	if keySource.KeyID != "managed-key" {
		t.Fatalf("KeyID = %q, want managed-key", keySource.KeyID)
	}
	if len(keySource.RawKey) != 0 {
		t.Fatalf("RawKey = %x, want empty", keySource.RawKey)
	}
	if !bytes.Equal(nonce, []byte{0x00}) || len(aad) != 0 {
		t.Fatalf("unexpected decoded metadata: nonce=%x aad=%x", nonce, aad)
	}
}

func TestParseEncryptionInputs_InvalidInput(t *testing.T) {
	tests := []struct {
		name      string
		keySource string
		key       string
		nonce     string
		aad       string
	}{
		{name: "invalid key source", keySource: "unknown", key: "00", nonce: "00"},
		{name: "invalid raw key hex", keySource: constant.KeySourceRaw, key: "invalid", nonce: "00"},
		{name: "invalid nonce hex", keySource: constant.KeySourceRaw, key: "00", nonce: "invalid"},
		{name: "missing key ID", keySource: constant.KeySourceKeyID, nonce: "00"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, _, _, err := parseEncryptionInputs(test.keySource, test.key, test.nonce, test.aad)
			if err == nil {
				t.Fatal("parseEncryptionInputs() error = nil, want non-nil")
			}
		})
	}
}

func BenchmarkEncryptData_profile_Default_Sequential(b *testing.B) {
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

	encryptCmd, err := NewEncryptData(ctx, lib, logger, tracerProvider)
	if err != nil {
		b.Fatalf("could not instantiate encrypt command, err: %s", err.Error())
	}

	for b.Loop() {
		if err := encryptCmd.encryptData(ctx, benchmarkEncryptionPlaintext, "Default", constant.KeySourceRaw,
			benchmarkEncryptionKey, benchmarkEncryptionNonce(), benchmarkEncryptionAAD); err != nil {
			b.Fatalf("could not run encrypt, err: %s", err.Error())
		}
	}
}

func BenchmarkEncryptData_profile_Default_Parallel(b *testing.B) {
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

		encryptCmd, err := NewEncryptData(ctx, lib, logger, tracerProvider)
		if err != nil {
			b.Fatalf("could not instantiate encrypt command, err: %s", err.Error())
		}

		for p.Next() {
			if err := encryptCmd.encryptData(ctx, benchmarkEncryptionPlaintext, "Default", constant.KeySourceRaw,
				benchmarkEncryptionKey, benchmarkEncryptionNonce(), benchmarkEncryptionAAD); err != nil {
				b.Fatalf("could not run encrypt, err: %s", err.Error())
			}
		}
	})
}

func benchmarkEncryptionNonce() string {
	nonce := make([]byte, 12)
	copy(nonce, "bench")
	binary.BigEndian.PutUint64(nonce[4:], benchmarkEncryptionNonceCounter.Add(1))
	
	return hex.EncodeToString(nonce)
}
