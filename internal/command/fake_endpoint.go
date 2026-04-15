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

// FakeEndpoint represents command that repeatedly sends fake endpoint request to crypto broker and displays its response
type FakeEndpoint struct {
	logger              *slog.Logger
	cryptoBrokerLibrary *cryptobrokerclientgo.Library
	tracerProvider      *otel.TracerProvider
}

// NewFakeEndpoint initializes fake endpoint command
func NewFakeEndpoint(ctx context.Context, lib *cryptobrokerclientgo.Library, logger *slog.Logger, tracerProvider *otel.TracerProvider) (*FakeEndpoint, error) {
	return &FakeEndpoint{
		logger:              logger,
		cryptoBrokerLibrary: lib,
		tracerProvider:      tracerProvider,
	}, nil
}

// Run executes command logic.
func (command *FakeEndpoint) Run(ctx context.Context, flagLoop int) error {
	defer func() { _ = command.gracefulShutdown() }()

	payload := cryptobrokerclientgo.FakeEndpointPayload{
		Metadata: nil,
	}

	command.logger.Info("Calling fake endpoint")

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
				if err := command.callFakeEndpoint(ctx, payload); err != nil {
					return err
				}

				time.Sleep(toSleep)
			}
		}
	} else {
		if err := command.callFakeEndpoint(ctx, payload); err != nil {
			return err
		}
		return nil
	}
}

// callFakeEndpoint sends fake endpoint request through crypto broker library.
// In case of success it displays response and returns nil error, otherwise it returns non-nil error.
// Internally method measures execution time and prints it through logger.
func (command *FakeEndpoint) callFakeEndpoint(ctx context.Context, payload cryptobrokerclientgo.FakeEndpointPayload) error {
	tracer := command.tracerProvider.GetTracer("crypto-broker-cli-go")
	correlationId := ""
	if payload.Metadata != nil && payload.Metadata.TraceContext != nil {
		correlationId = payload.Metadata.TraceContext.CorrelationId
	}

	ctx, span := tracer.Start(ctx, "CLI.FakeEndpoint",
		trace.WithAttributes(
			otel.AttributeRpcMethod.String("FakeEndpoint"),
			otel.AttributeCorrelationId.String(correlationId),
		))
	defer span.End()

	spanContext := span.SpanContext()
	if payload.Metadata == nil {
		payload.Metadata = &cryptobrokerclientgo.Metadata{
			Id:        uuid.New().String(),
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
		}
	}
	payload.Metadata.TraceContext = &cryptobrokerclientgo.TraceContext{
		TraceId:       spanContext.TraceID().String(),
		SpanId:        spanContext.SpanID().String(),
		TraceFlags:    spanContext.TraceFlags().String(),
		TraceState:    spanContext.TraceState().String(),
		CorrelationId: correlationId,
	}

	timestampFakeEndpointStart := time.Now()
	responseBody, err := command.cryptoBrokerLibrary.FakeEndpoint(ctx, payload)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	timestampFakeEndpointFinish := time.Now()
	durationElapsedFakeEndpoint := timestampFakeEndpointFinish.Sub(timestampFakeEndpointStart)

	span.SetStatus(codes.Ok, "Fake endpoint operation completed successfully")

	command.logger.Info("Fake endpoint response", "response", responseBody)
	command.logger.Info("Fake endpoint call took", "duration_microseconds", float64(durationElapsedFakeEndpoint.Nanoseconds())/1000.0)

	return nil
}

// gracefulShutdown closes library connection.
func (command *FakeEndpoint) gracefulShutdown() error {
	command.logger.Info("Closing crypto broker library connection")
	return command.cryptoBrokerLibrary.Close()
}
