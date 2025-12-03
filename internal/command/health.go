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

	"github.com/open-crypto-broker/crypto-broker-CLI-go/internal/constant"
	cryptobrokerclientgo "github.com/open-crypto-broker/crypto-broker-client-go"
)

// Health represents command that checks broker server health status
type Health struct {
	logger              *log.Logger
	cryptoBrokerLibrary *cryptobrokerclientgo.Library
}

// NewHealth initializes health command
func NewHealth(ctx context.Context, logger *log.Logger) (*Health, error) {
	lib, err := cryptobrokerclientgo.NewLibrary(ctx)
	if err != nil {
		return nil, err
	}

	return &Health{logger: logger, cryptoBrokerLibrary: lib}, nil
}

// Run executes command logic.
func (command *Health) Run(ctx context.Context, flagLoop int) error {
	defer command.gracefulShutdown()

	command.logger.Printf("Checking broker server health\n")

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
	timestampStart := time.Now()
	responseBody, err := command.cryptoBrokerLibrary.HealthData(ctx)
	if err != nil {
		return err
	}

	timestampFinish := time.Now()
	durationElapsed := timestampFinish.Sub(timestampStart)
	marshalledResp, err := json.MarshalIndent(responseBody, " ", "  ")
	if err != nil {
		return err
	}

	command.logger.Println("Health check response:\n", string(marshalledResp))
	command.logger.Printf("Health check took: %fµs\n", float64(durationElapsed.Nanoseconds())/1000.0)

	return nil
}

// gracefulShutdown closes library connection.
func (command *Health) gracefulShutdown() error {
	command.logger.Printf("Closing crypto broker library connection\n")
	return command.cryptoBrokerLibrary.Close()
}
