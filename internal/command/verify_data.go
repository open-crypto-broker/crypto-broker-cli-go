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

// VerifyData sends arbitrary data and a signature to Crypto Broker for verification.
type VerifyData struct {
	logger              *slog.Logger
	cryptoBrokerLibrary *cryptobrokerclientgo.Library
	tracerProvider      *otel.TracerProvider
}

func NewVerifyData(ctx context.Context, lib *cryptobrokerclientgo.Library, logger *slog.Logger, tracerProvider *otel.TracerProvider) (*VerifyData, error) {
	return &VerifyData{logger: logger, cryptoBrokerLibrary: lib, tracerProvider: tracerProvider}, nil
}

func (command *VerifyData) Run(ctx context.Context, data, signature []byte, profile, keyPath string, loop int) error {
	defer func() { _ = command.gracefulShutdown() }()
	key, err := readPEMKey(keyPath)
	if err != nil {
		return err
	}
	command.logger.Info("Verifying data")

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	if loop < constant.MinLoopFlagValue || loop > constant.MaxLoopFlagValue {
		_, err := command.verifyData(ctx, data, signature, profile, key)
		return err
	}

	delay, err := time.ParseDuration(fmt.Sprintf("%dms", loop))
	if err != nil {
		return err
	}
	for {
		select {
		case <-signals:
			command.logger.Info("Received SIGTERM signal")
			return nil
		default:
			if _, err := command.verifyData(ctx, data, signature, profile, key); err != nil {
				return err
			}
			time.Sleep(delay)
		}
	}
}

func (command *VerifyData) verifyData(ctx context.Context, data, signature []byte, profile string, key []byte) (bool, error) {
	tracer := command.tracerProvider.GetTracer("crypto-broker-cli-go")
	ctx, span := tracer.Start(ctx, "CLI.VerifyData", trace.WithAttributes(
		otel.AttributeRpcMethod.String("VerifyData"),
		otel.AttributeCryptoProfile.String(profile),
		otel.AttributeCryptoInputSize.Int(len(data)),
	))
	defer span.End()

	spanContext := span.SpanContext()
	signatureFormat := cryptobrokerclientgo.SignatureFormatRaw
	response, err := command.cryptoBrokerLibrary.VerifyData(ctx, cryptobrokerclientgo.VerifyDataPayload{
		Profile:         profile,
		SignatureFormat: &signatureFormat,
		KeySource: cryptobrokerclientgo.SignKeySource{Single: &cryptobrokerclientgo.KeySource{
			RawKey: key,
		}},
		Input: data, Signature: signature,
		Metadata: &cryptobrokerclientgo.Metadata{Id: uuid.New().String(), TraceContext: &cryptobrokerclientgo.TraceContext{
			TraceId: spanContext.TraceID().String(), SpanId: spanContext.SpanID().String(), TraceFlags: spanContext.TraceFlags().String(), TraceState: spanContext.TraceState().String(),
		}},
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return false, fmt.Errorf("could not verify data through Crypto Broker: %w", err)
	}

	span.SetStatus(codes.Ok, "VerifyData operation completed successfully")
	command.logger.Info("Verify data response", "valid", response.GetValid())
	return response.GetValid(), nil
}

func (command *VerifyData) gracefulShutdown() error {
	command.logger.Info("Closing crypto broker library connection")
	return command.cryptoBrokerLibrary.Close()
}
