package command

import (
	"context"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/constant"
	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/otel"
	cryptobrokerclientgo "github.com/open-crypto-broker/crypto-broker-client-go"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// EncryptData sends AES-GCM encryption requests to Crypto Broker.
type EncryptData struct {
	logger              *slog.Logger
	cryptoBrokerLibrary *cryptobrokerclientgo.Library
	tracerProvider      *otel.TracerProvider
}

// NewEncryptData initializes encrypt command
func NewEncryptData(ctx context.Context, lib *cryptobrokerclientgo.Library, logger *slog.Logger, tracerProvider *otel.TracerProvider) (*EncryptData, error) {
	return &EncryptData{
		logger:              logger,
		cryptoBrokerLibrary: lib,
		tracerProvider:      tracerProvider,
	}, nil
}

// Run executes command logic.
func (command *EncryptData) Run(ctx context.Context, data []byte, flagProfile, flagKeyRaw, flagKeyID, flagNonce, flagAAD string, flagLoop int) error {
	defer func() { _ = command.gracefulShutdown() }()

	command.logger.Info("Encrypting data")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	if flagLoop >= constant.MinLoopFlagValue && flagLoop <= constant.MaxLoopFlagValue {
		toSleep, err := time.ParseDuration(fmt.Sprintf("%dms", flagLoop))
		if err != nil {
			panic(err)
		}

		for {
			select {
			case <-c:
				command.logger.Info("Received SIGTERM signal")
				return nil
			default:
				if err := command.encryptData(ctx, data, flagProfile, flagKeyRaw, flagKeyID, flagNonce, flagAAD); err != nil {
					return err
				}

				time.Sleep(toSleep)
			}
		}
	} else {
		if err := command.encryptData(ctx, data, flagProfile, flagKeyRaw, flagKeyID, flagNonce, flagAAD); err != nil {
			return err
		}

		return nil
	}
}

// encryptData encrypts the supplied plaintext and logs a hex-encoded response.
func (command *EncryptData) encryptData(ctx context.Context, data []byte, flagProfile, flagKeyRaw, flagKeyID, flagNonce, flagAAD string) error {
	keySource, nonce, aad, err := parseEncryptionInputs(flagKeyRaw, flagKeyID, flagNonce, flagAAD)
	if err != nil {
		return err
	}

	tracer := command.tracerProvider.GetTracer("crypto-broker-cli-go")
	ctx, span := tracer.Start(ctx, "CLI.EncryptData",
		trace.WithAttributes(
			otel.AttributeRpcMethod.String("EncryptData"),
			otel.AttributeCryptoProfile.String(flagProfile),
			otel.AttributeCryptoInputSize.Int(len(data)),
		))
	defer span.End()

	spanContext := span.SpanContext()
	response, err := command.cryptoBrokerLibrary.EncryptData(ctx, cryptobrokerclientgo.EncryptDataPayload{
		Profile:   flagProfile,
		KeySource: keySource,
		Plaintext: data,
		Nonce:     nonce,
		AAD:       aad,
		Metadata: &cryptobrokerclientgo.Metadata{
			Id: uuid.New().String(),
			TraceContext: &cryptobrokerclientgo.TraceContext{
				TraceId:    spanContext.TraceID().String(),
				SpanId:     spanContext.SpanID().String(),
				TraceFlags: spanContext.TraceFlags().String(),
				TraceState: spanContext.TraceState().String(),
			},
		},
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return fmt.Errorf("could not encrypt data through Crypto Broker: %w", err)
	}

	span.SetStatus(codes.Ok, "EncryptData operation completed successfully")
	command.logger.Info("Encrypt data response",
		"ciphertext", hex.EncodeToString(response.GetCiphertext()),
		"tag", hex.EncodeToString(response.GetCipherMetadata().GetTag()),
	)

	return nil
}

// gracefulShutdown closes library connection.
func (command *EncryptData) gracefulShutdown() error {
	command.logger.Info("Closing crypto broker library connection")
	return command.cryptoBrokerLibrary.Close()
}

func parseEncryptionInputs(keyRaw, keyID, nonce, aad string) (cryptobrokerclientgo.KeySource, []byte, []byte, error) {
	nonceBytes, err := decodeHex("nonce", nonce)
	if err != nil {
		return cryptobrokerclientgo.KeySource{}, nil, nil, err
	}

	aadBytes, err := decodeHex("aad", aad)
	if err != nil {
		return cryptobrokerclientgo.KeySource{}, nil, nil, err
	}

	if keyRaw != "" {
		rawKey, err := decodeHex("key", keyRaw)
		if err != nil {
			return cryptobrokerclientgo.KeySource{}, nil, nil, err
		}
		return cryptobrokerclientgo.KeySource{RawKey: rawKey}, nonceBytes, aadBytes, nil
	}

	if keyID != "" {
		return cryptobrokerclientgo.KeySource{KeyID: keyID}, nonceBytes, aadBytes, nil
	}

	return cryptobrokerclientgo.KeySource{}, nil, nil, fmt.Errorf("either keyRaw or keyID must be provided")
}

func decodeHex(name, value string) ([]byte, error) {
	decoded, err := hex.DecodeString(value)
	if err != nil {
		return nil, fmt.Errorf("invalid hexadecimal %s: %w", name, err)
	}

	return decoded, nil
}
