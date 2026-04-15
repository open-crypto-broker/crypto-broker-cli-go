package command

import (
	"context"
	"encoding/json"
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

// Benchmark represents command that runs server-side cryptographic benchmarks
type Benchmark struct {
	logger              *slog.Logger
	cryptoBrokerLibrary *cryptobrokerclientgo.Library
	tracerProvider      *otel.TracerProvider
}

// NewBenchmark initializes benchmark command
func NewBenchmark(ctx context.Context, lib *cryptobrokerclientgo.Library, logger *slog.Logger, tracerProvider *otel.TracerProvider) (*Benchmark, error) {
	return &Benchmark{
		logger:              logger,
		cryptoBrokerLibrary: lib,
		tracerProvider:      tracerProvider,
	}, nil
}

// Run executes command logic.
func (command *Benchmark) Run(ctx context.Context, flagLoop int) error {
	defer func() { _ = command.gracefulShutdown() }()

	command.logger.Info("Running server-side benchmarks")

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
				if err := command.runBenchmark(ctx); err != nil {
					return err
				}

				time.Sleep(toSleep)
			}
		}
	} else {
		if err := command.runBenchmark(ctx); err != nil {
			return err
		}
		return nil
	}
}

// runBenchmark sends benchmark request through crypto broker library.
// In case of success it displays response and returns nil error, otherwise it returns non-nil error.
// Internally method measures execution time and prints it through logger.
func (command *Benchmark) runBenchmark(ctx context.Context) error {
	tracer := command.tracerProvider.GetTracer("crypto-broker-cli-go")
	ctx, span := tracer.Start(ctx, "CLI.Benchmark",
		trace.WithAttributes(otel.AttributeRpcMethod.String("Benchmark")))
	defer span.End()

	// Inject trace context into payload metadata
	spanContext := span.SpanContext()
	payload := cryptobrokerclientgo.BenchmarkDataPayload{
		Metadata: &cryptobrokerclientgo.Metadata{
			Id:        uuid.New().String(),
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
			TraceContext: &cryptobrokerclientgo.TraceContext{
				TraceId:    spanContext.TraceID().String(),
				SpanId:     spanContext.SpanID().String(),
				TraceFlags: spanContext.TraceFlags().String(),
				TraceState: spanContext.TraceState().String(),
			},
		},
	}

	timestampStart := time.Now()
	responseBody, err := command.cryptoBrokerLibrary.BenchmarkData(ctx, payload)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	timestampFinish := time.Now()
	durationElapsed := timestampFinish.Sub(timestampStart)
	marshalledResp, err := json.MarshalIndent(responseBody, " ", "  ")
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetAttributes(otel.AttributeCryptoBenchmarkResultsSize.Int(len(marshalledResp)))
	span.SetStatus(codes.Ok, "Benchmark operation completed successfully")

	command.logger.Info("Benchmark results", "results", responseBody)
	command.logger.Info(
		fmt.Sprintf("Server-side Benchmarking took %d µs", durationElapsed.Microseconds()),
	)
	return nil
}

// gracefulShutdown closes library connection.
func (command *Benchmark) gracefulShutdown() error {
	command.logger.Info("Closing crypto broker library connection")
	return command.cryptoBrokerLibrary.Close()
}
