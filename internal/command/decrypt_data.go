package command

import (
	"context"
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

// DecryptData sends AES-GCM decryption requests to Crypto Broker.
type DecryptData struct {
	logger              *slog.Logger
	cryptoBrokerLibrary *cryptobrokerclientgo.Library
	tracerProvider      *otel.TracerProvider
}

// NewDecryptData initializes decrypt command.
func NewDecryptData(ctx context.Context, lib *cryptobrokerclientgo.Library, logger *slog.Logger, tracerProvider *otel.TracerProvider) (*DecryptData, error) {
	return &DecryptData{
		logger:              logger,
		cryptoBrokerLibrary: lib,
		tracerProvider:      tracerProvider,
	}, nil
}

// Run executes command logic.
func (command *DecryptData) Run(ctx context.Context, data []byte, flagProfile, flagKeySource, flagKey, flagNonce, flagAAD, flagTag string, flagLoop int) error {
	defer func() { _ = command.gracefulShutdown() }()

	command.logger.Info("Decrypting data")

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
				if err := command.decryptData(ctx, data, flagProfile, flagKeySource, flagKey, flagNonce, flagAAD, flagTag); err != nil {
					return err
				}

				time.Sleep(toSleep)
			}
		}
	} else {
		if err := command.decryptData(ctx, data, flagProfile, flagKeySource, flagKey, flagNonce, flagAAD, flagTag); err != nil {
			return err
		}
		return nil
	}
}

// decryptData decrypts the supplied ciphertext and logs a UTF-8 plaintext response.
func (command *DecryptData) decryptData(ctx context.Context, data []byte, flagProfile, flagKeySource, flagKey, flagNonce, flagAAD, flagTag string) error {
	keySource, nonce, aad, err := parseEncryptionInputs(flagKeySource, flagKey, flagNonce, flagAAD)
	if err != nil {
		return err
	}
	tag, err := decodeHex("tag", flagTag)
	if err != nil {
		return err
	}

	tracer := command.tracerProvider.GetTracer("crypto-broker-cli-go")
	ctx, span := tracer.Start(ctx, "CLI.DecryptData",
		trace.WithAttributes(
			otel.AttributeRpcMethod.String("DecryptData"),
			otel.AttributeCryptoProfile.String(flagProfile),
			otel.AttributeCryptoInputSize.Int(len(data)),
		))
	defer span.End()

	spanContext := span.SpanContext()
	response, err := command.cryptoBrokerLibrary.DecryptData(ctx, cryptobrokerclientgo.DecryptDataPayload{
		Profile:    flagProfile,
		KeySource:  keySource,
		Ciphertext: data,
		Nonce:      nonce,
		AAD:        aad,
		Tag:        tag,
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
		return fmt.Errorf("could not decrypt data through Crypto Broker: %w", err)
	}

	span.SetStatus(codes.Ok, "DecryptData operation completed successfully")
	command.logger.Info("Decrypt data response", "plaintext", string(response.GetPlaintext()))

	return nil
}

// gracefulShutdown closes library connection.
func (command *DecryptData) gracefulShutdown() error {
	command.logger.Info("Closing crypto broker library connection")
	return command.cryptoBrokerLibrary.Close()
}
