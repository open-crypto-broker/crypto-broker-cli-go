package command

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/constant"
	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/otel"
	cryptobrokerclientgo "github.com/open-crypto-broker/crypto-broker-client-go"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type SignCertificate struct {
	logger              *slog.Logger
	cryptoBrokerLibrary *cryptobrokerclientgo.Library
	tracerProvider      *otel.TracerProvider
}

// NewSignCertificate initializes sign command. This may panic in case of failure.
func NewSignCertificate(ctx context.Context, lib *cryptobrokerclientgo.Library, logger *slog.Logger, tracerProvider *otel.TracerProvider) (*SignCertificate, error) {
	return &SignCertificate{
		logger:              logger,
		cryptoBrokerLibrary: lib,
		tracerProvider:      tracerProvider,
	}, nil
}

// Run executes command logic.
func (command *SignCertificate) Run(ctx context.Context, filePathCSR, filePathCACert, filePathSigningKey, flagProfile, flagEncoding, flagSubject string, flagLoop int) error {
	defer func() { _ = command.gracefulShutdown() }()

	rawContentCSR, err := command.readFileBytes(filePathCSR)
	if err != nil {
		return fmt.Errorf("could not read certificate signing request file, err: %w", err)
	}

	rawContentCACert, err := command.readFileBytes(filePathCACert)
	if err != nil {
		return fmt.Errorf("could not read CA Certificate file, err: %w", err)
	}

	rawContentSigningKey, err := command.readFileBytes(filePathSigningKey)
	if err != nil {
		return fmt.Errorf("could not read signing key file, err: %w", err)
	}

	var subject *string
	if flagSubject != "" {
		subject = &flagSubject
	} else {
		subject = nil
	}

	payload := cryptobrokerclientgo.SignCertificatePayload{
		Profile:      flagProfile,
		CSR:          rawContentCSR,
		CAPrivateKey: rawContentSigningKey,
		CACert:       rawContentCACert,
		Subject:      subject,
		Metadata:     nil, // Will be set in signCertificate with trace context
	}

	command.logger.Info(
		fmt.Sprintf("Signing certificate using %s profile", flagProfile),
	)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	if flagLoop >= constant.MinLoopFlagValue && flagLoop <= constant.MaxLoopFlagValue {
		toSleep, err := time.ParseDuration(fmt.Sprintf("%dms", flagLoop))
		if err != nil {
			return fmt.Errorf("could not parse duration, err: %w", err)
		}

		for {
			select {
			case <-c:
				command.logger.Info("Received SIGTERM signal")
				return nil
			default:
				if err := command.signCertificate(ctx, payload, flagEncoding); err != nil {
					return err
				}

				time.Sleep(toSleep)
			}
		}
	} else {
		if err := command.signCertificate(ctx, payload, flagEncoding); err != nil {
			return err
		}
		return nil
	}
}

func (command *SignCertificate) signCertificate(ctx context.Context, payload cryptobrokerclientgo.SignCertificatePayload, flagEncoding string) error {
	tracer := command.tracerProvider.GetTracer("crypto-broker-cli-go")
	correlationId := ""
	if payload.Metadata != nil && payload.Metadata.TraceContext != nil {
		correlationId = payload.Metadata.TraceContext.CorrelationId
	}
	ctx, span := tracer.Start(ctx, "CLI.SignCertificate",
		trace.WithAttributes(
			otel.AttributeRpcMethod.String("SignCertificate"),
			otel.AttributeCryptoProfile.String(payload.Profile),
			otel.AttributeCryptoCsrSize.Int(len(payload.CSR)),
			otel.AttributeCryptoCaCertSize.Int(len(payload.CACert)),
			otel.AttributeCryptoCaKeySize.Int(len(payload.CAPrivateKey)),
			otel.AttributeCorrelationId.String(correlationId),
		))
	defer span.End()

	// Inject trace context into payload metadata
	spanContext := span.SpanContext()
	if payload.Metadata == nil {
		payload.Metadata = &cryptobrokerclientgo.Metadata{
			Id: uuid.New().String(),
		}
	}
	payload.Metadata.TraceContext = &cryptobrokerclientgo.TraceContext{
		TraceId:       spanContext.TraceID().String(),
		SpanId:        spanContext.SpanID().String(),
		TraceFlags:    spanContext.TraceFlags().String(),
		TraceState:    spanContext.TraceState().String(),
		CorrelationId: correlationId,
	}

	timestampSignCertificateStart := time.Now()
	payload.OutputFormat = cryptobrokerclientgo.OutputFormatPem // default output format
	if strings.ToLower(flagEncoding) == constant.EncodingDER {
		payload.OutputFormat = cryptobrokerclientgo.OutputFormatDer
	}

	responseBody, err := command.cryptoBrokerLibrary.SignCertificate(ctx, payload)
	if err != nil && !errors.Is(err, cryptobrokerclientgo.ErrCircuitOpen) {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("failed to obtain signed certificate through CryptoBroker library, err: %w", err)
	}

	if responseBody != nil {
		timestampSignCertificateFinish := time.Now()
		durationElapsedSignCertificate := timestampSignCertificateFinish.Sub(timestampSignCertificateStart)

		span.SetAttributes(otel.AttributeCryptoSignedCertSize.Int(len(responseBody.GetDer()) + len(responseBody.GetPem())))
		span.SetStatus(codes.Ok, "Certificate signing completed successfully")

		command.logger.Info("Sign certificate response", "response", responseBody)
		command.logger.Info(
			fmt.Sprintf("Certificate Signing took %d µs", durationElapsedSignCertificate.Microseconds()),
		)
	}

	return nil
}

// gracefulShutdown closes library connection.
func (command *SignCertificate) gracefulShutdown() error {
	command.logger.Info("Closing crypto broker library connection")
	return command.cryptoBrokerLibrary.Close()
}

// readFileBytes opens a file and reads its bytes
func (command *SignCertificate) readFileBytes(filePath string) ([]byte, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not open %s file, err: %w", filePath, err)
	}

	defer func() { _ = f.Close() }()

	rawContent, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("could not read %s file, err: %w", filePath, err)
	}

	return rawContent, nil
}
