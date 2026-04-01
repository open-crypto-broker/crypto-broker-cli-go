package command

import (
	"context"
	"fmt"
	"log/slog"
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

// Health represents command that checks broker server health status
type Health struct {
	logger              *slog.Logger
	cryptoBrokerLibrary *cryptobrokerclientgo.Library
	tracerProvider      *otel.TracerProvider
}

// NewHealth initializes health command
func NewHealth(ctx context.Context, lib *cryptobrokerclientgo.Library, logger *slog.Logger, tracerProvider *otel.TracerProvider) (*Health, error) {
	return &Health{
		logger:              logger,
		cryptoBrokerLibrary: lib,
		tracerProvider:      tracerProvider,
	}, nil
}

// Run executes command logic.
func (command *Health) Run(ctx context.Context, flagLoop int) error {
	defer command.gracefulShutdown()

	command.logger.Info("Checking broker server health")

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
				if err := command.checkHealth(ctx); err != nil {
					return err
				}

				time.Sleep(toSleep)
			}
		}
	} else {
		if err := command.checkHealth(ctx); err != nil {
			return err
		}
		return nil
	}
}

// checkHealth sends health check request through crypto broker library.
// In case of success it displays response and returns nil error, otherwise it returns non-nil error.
// Internally method measures execution time and prints it through logger.
func (command *Health) checkHealth(ctx context.Context) error {
	tracer := command.tracerProvider.GetTracer("crypto-broker-cli-go")
	ctx, span := tracer.Start(ctx, "CLI.Health",
		trace.WithAttributes(otel.AttributeRpcMethod.String("Health")))
	defer span.End()

	timestampStart := time.Now()
	responseBody := command.cryptoBrokerLibrary.HealthData(ctx)
	timestampFinish := time.Now()
	durationElapsed := timestampFinish.Sub(timestampStart)

	span.SetStatus(codes.Ok, "Health check completed successfully")

	command.logger.Info("Health check response", "response", responseBody)
	command.logger.Info("Health check took", "duration_microseconds", float64(durationElapsed.Nanoseconds())/1000.0)

	return nil
}

// gracefulShutdown closes library connection.
func (command *Health) gracefulShutdown() error {
	command.logger.Info("Closing crypto broker library connection")
	return command.cryptoBrokerLibrary.Close()
}
