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

// SignData sends arbitrary data to Crypto Broker for signing.
type SignData struct {
	logger              *slog.Logger
	cryptoBrokerLibrary *cryptobrokerclientgo.Library
	tracerProvider      *otel.TracerProvider
}

func NewSignData(ctx context.Context, lib *cryptobrokerclientgo.Library, logger *slog.Logger, tracerProvider *otel.TracerProvider) (*SignData, error) {
	return &SignData{logger: logger, cryptoBrokerLibrary: lib, tracerProvider: tracerProvider}, nil
}

func (command *SignData) Run(ctx context.Context, data []byte, profile, keyPath string, loop int) error {
	defer func() { _ = command.gracefulShutdown() }()
	key, err := readPEMKey(keyPath)
	if err != nil {
		return err
	}

	command.logger.Info("Signing data")

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	if loop < constant.MinLoopFlagValue || loop > constant.MaxLoopFlagValue {
		_, err := command.signData(ctx, data, profile, key)
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
			if _, err := command.signData(ctx, data, profile, key); err != nil {
				return err
			}
			time.Sleep(delay)
		}
	}
}

func (command *SignData) signData(ctx context.Context, data []byte, profile string, key []byte) ([]byte, error) {
	tracer := command.tracerProvider.GetTracer(otel.ServiceName)
	ctx, span := tracer.Start(ctx, "CLI.SignData", trace.WithAttributes(
		otel.AttributeRpcMethod.String("SignData"),
		otel.AttributeCryptoProfile.String(profile),
		otel.AttributeCryptoInputSize.Int(len(data)),
	))
	defer span.End()

	spanContext := span.SpanContext()
	signatureFormat := cryptobrokerclientgo.SignatureFormatRaw
	response, err := command.cryptoBrokerLibrary.SignData(ctx, cryptobrokerclientgo.SignDataPayload{
		Profile:         profile,
		SignatureFormat: &signatureFormat,
		KeySource: cryptobrokerclientgo.SignKeySource{Single: &cryptobrokerclientgo.KeySource{
			RawKey: key,
		}},
		Input: data,
		Metadata: &cryptobrokerclientgo.Metadata{Id: uuid.New().String(), TraceContext: &cryptobrokerclientgo.TraceContext{
			TraceId: spanContext.TraceID().String(), SpanId: spanContext.SpanID().String(), TraceFlags: spanContext.TraceFlags().String(), TraceState: spanContext.TraceState().String(),
		}},
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return nil, fmt.Errorf("could not sign data through Crypto Broker: %w", err)
	}

	span.SetStatus(codes.Ok, "SignData operation completed successfully")
	command.logger.Info("Sign data response", "signature", hex.EncodeToString(response.GetSignature()), "descriptor", response.GetDescriptor_())

	return response.GetSignature(), nil
}

func (command *SignData) gracefulShutdown() error {
	command.logger.Info("Closing crypto broker library connection")

	return command.cryptoBrokerLibrary.Close()
}

func readPEMKey(path string) ([]byte, error) {
	if path == "" {
		return nil, fmt.Errorf("key is required")
	}

	key, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read key file %q: %w", path, err)
	}

	if len(key) == 0 {
		return nil, fmt.Errorf("key file %q is empty", path)
	}

	return key, nil
}
