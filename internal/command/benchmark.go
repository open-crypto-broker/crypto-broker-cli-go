package command

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/constant"
	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/otel"
	cryptobrokerclientgo "github.com/open-crypto-broker/crypto-broker-client-go"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Benchmark represents command that runs server-side cryptographic benchmarks
type Benchmark struct {
	logger              *log.Logger
	cryptoBrokerLibrary *cryptobrokerclientgo.Library
	tracerProvider      *otel.TracerProvider
}

// NewBenchmark initializes benchmark command
func NewBenchmark(ctx context.Context, logger *log.Logger, tracerProvider *otel.TracerProvider) (*Benchmark, error) {
	lib, err := cryptobrokerclientgo.NewLibrary(ctx)
	if err != nil {
		return nil, err
	}

	return &Benchmark{
		logger:              logger,
		cryptoBrokerLibrary: lib,
		tracerProvider:      tracerProvider,
	}, nil
}

// Run executes command logic.
func (command *Benchmark) Run(ctx context.Context, flagLoop int) error {
	defer command.gracefulShutdown()

	command.logger.Printf("Running server-side benchmarks\n")

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
				command.logger.Printf("Received SIGTERM signal\n")
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

	timestampStart := time.Now()
	responseBody, err := command.cryptoBrokerLibrary.BenchmarkData(ctx, cryptobrokerclientgo.BenchmarkDataPayload{})
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

	command.logger.Println("Benchmark results:\n", string(marshalledResp))
	command.logger.Printf("Benchmark execution took: %fµs\n", float64(durationElapsed.Nanoseconds())/1000.0)

	return nil
}

// gracefulShutdown closes library connection.
func (command *Benchmark) gracefulShutdown() error {
	command.logger.Printf("Closing crypto broker library connection\n")
	return command.cryptoBrokerLibrary.Close()
}
