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

	"github.com/google/uuid"

	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/constant"
	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/otel"
	cryptobrokerclientgo "github.com/open-crypto-broker/crypto-broker-client-go"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Hash represents command that repeatedly sends hash request to crypto broker and displays its response
type Hash struct {
	logger              *log.Logger
	cryptoBrokerLibrary *cryptobrokerclientgo.Library
	tracerProvider      *otel.TracerProvider
}

// NewHash initializes hash command
func NewHash(ctx context.Context, logger *log.Logger, tracerProvider *otel.TracerProvider) (*Hash, error) {
	lib, err := cryptobrokerclientgo.NewLibrary(ctx)
	if err != nil {
		return nil, err
	}

	return &Hash{
		logger:              logger,
		cryptoBrokerLibrary: lib,
		tracerProvider:      tracerProvider,
	}, nil
}

// Run executes command logic.
func (command *Hash) Run(ctx context.Context, input []byte, flagProfile string, flagLoop int) error {
	defer command.gracefulShutdown()

	payload := cryptobrokerclientgo.HashDataPayload{
		Input:   input,
		Profile: flagProfile,
		Metadata: &cryptobrokerclientgo.Metadata{
			Id:        uuid.New().String(),
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
		},
	}

	command.logger.Printf("Hashing \"%s\" using %s profile \n", string(input), flagProfile)

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
				if err := command.hashBytes(ctx, payload); err != nil {
					return err
				}

				time.Sleep(toSleep)
			}
		}
	} else {
		if err := command.hashBytes(ctx, payload); err != nil {
			return err
		}
		return nil
	}
}

// hashBytes sends hash request through crypto broker library.
// In case of success it displays response and returns nil error, otherwise it returns non-nil error.
// Internally method measures execution time and prints it through logger.
func (command *Hash) hashBytes(ctx context.Context, payload cryptobrokerclientgo.HashDataPayload) error {
	tracer := command.tracerProvider.GetTracer("crypto-broker-cli-go")
	ctx, span := tracer.Start(ctx, "CLI.Hash",
		trace.WithAttributes(
			otel.AttributeRpcMethod.String("Hash"),
			otel.AttributeCryptoProfile.String(payload.Profile),
			otel.AttributeCryptoInputSize.Int(len(payload.Input)),
		))
	defer span.End()

	timestampHashingStart := time.Now()
	responseBody, err := command.cryptoBrokerLibrary.HashData(ctx, payload)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	timestampHashingFinish := time.Now()
	durationElapsedHashing := timestampHashingFinish.Sub(timestampHashingStart)
	marshalledResp, err := json.MarshalIndent(responseBody, " ", "  ")
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetAttributes(
		otel.AttributeCryptoHashAlgorithm.String(responseBody.HashAlgorithm),
		otel.AttributeCryptoHashOutputSize.Int(len(responseBody.HashValue)),
	)
	span.SetStatus(codes.Ok, "Hash operation completed successfully")

	command.logger.Println("Hashed response:\n", string(marshalledResp))
	command.logger.Printf("Data Hashing took: %fµs\n", float64(durationElapsedHashing.Nanoseconds())/1000.0)

	return nil
}

// gracefulShutdown closes library connection.
func (command *Hash) gracefulShutdown() error {
	command.logger.Printf("Closing crypto broker library connection\n")
	return command.cryptoBrokerLibrary.Close()
}
