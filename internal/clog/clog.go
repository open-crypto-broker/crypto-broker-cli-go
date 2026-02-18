// Package clog stands for crypto broker logger. Contains utilities related with logging.
package clog

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/env"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/sdk/log"
)

// SetupGlobalLogger initializes the crypto broker logger.
// It predefines defaults for logger. If user provides custom values that are not supported by the logger, it panics.
// It sets the logger to the default global logger.
// Supports OTEL_LOGS_EXPORTER with values: "console", "otlp", "otlphttp", "otlpgrpc", or comma-separated combinations
func SetupGlobalLogger(ctx context.Context) *slog.Logger {
	logsExporter := strings.ToLower(strings.TrimSpace(os.Getenv(env.OTEL_LOGS_EXPORTER)))

	exporters := strings.Split(logsExporter, ",")
	for i, exporter := range exporters {
		exporters[i] = strings.TrimSpace(exporter)
	}

	var useOTLPHTTP, useOTLPGRPC, useOTLPAuto, useConsole bool
	for _, exporter := range exporters {
		switch exporter {
		case keywordExporterOTLPHTTP:
			useOTLPHTTP = true
		case keywordExporterOTLPGRPC:
			useOTLPGRPC = true
		case keywordExporterOTLP:
			useOTLPAuto = true
		case keywordExporterConsole:
			useConsole = true
		case "":
			useConsole = true
		default:
			fmt.Fprintf(os.Stderr, "Invalid OTEL_LOGS_EXPORTER value '%s', falling back to console logging\n", exporter)
			useConsole = true
		}
	}

	if !useOTLPHTTP && !useOTLPGRPC && !useOTLPAuto && !useConsole {
		useConsole = true
	}

	var otlpLogger *slog.Logger
	if useOTLPHTTP {
		otlpLogger = setupOTLPLoggerWithProtocol(ctx, "http")
	} else if useOTLPGRPC {
		otlpLogger = setupOTLPLoggerWithProtocol(ctx, "grpc")
	} else if useOTLPAuto {
		otlpLogger = setupOTLPLogger(ctx)
	}

	if (useOTLPHTTP || useOTLPGRPC || useOTLPAuto) && useConsole {
		return setupMultiLoggerWithOTLP(otlpLogger)
	} else if useOTLPHTTP || useOTLPGRPC || useOTLPAuto {
		return otlpLogger
	} else {
		return setupConsoleLogger()
	}
}

// setupConsoleLogger sets up the traditional console-based logging
// this function may panic if the log level or log output is invalid
func setupConsoleLogger() *slog.Logger {
	logLevel := slog.LevelInfo
	userProvidedLogLevel := strings.ToLower(os.Getenv(env.LOG_LEVEL))
	if userProvidedLogLevel != "" {
		switch userProvidedLogLevel {
		case strings.ToLower(logLevelDebug):
			logLevel = slog.LevelDebug
		case strings.ToLower(logLevelInfo):
			logLevel = slog.LevelInfo
		case strings.ToLower(logLevelWarn):
			logLevel = slog.LevelWarn
		case strings.ToLower(logLevelError):
			logLevel = slog.LevelError
		default:
			panic(fmt.Sprintf("invalid log level provided: %s, available levels: %s, %s, %s, %s",
				userProvidedLogLevel, logLevelDebug, logLevelInfo, logLevelWarn, logLevelError))
		}
	}

	var logOutput *os.File
	logOutput = os.Stdout // default
	userProvidedLogOutput := strings.ToLower(os.Getenv(env.LOG_OUTPUT))
	if userProvidedLogOutput != "" {
		switch userProvidedLogOutput {
		case strings.ToLower(logOutputStdout):
			logOutput = os.Stdout
		case strings.ToLower(logOutputStderr):
			logOutput = os.Stderr
		default:
			panic(fmt.Sprintf("invalid log output provided: %s, available outputs: %s, %s",
				userProvidedLogOutput, logOutputStdout, logOutputStderr))
		}
	}

	var logHandler slog.Handler
	logHandler = slog.NewJSONHandler(logOutput, &slog.HandlerOptions{Level: logLevel}) // default
	userProvidedLogFormat := strings.ToLower(os.Getenv(env.LOG_FORMAT))
	if userProvidedLogFormat != "" {
		switch userProvidedLogFormat {
		case strings.ToLower(logFormatJSON):
			logHandler = slog.NewJSONHandler(logOutput, &slog.HandlerOptions{Level: logLevel})
		case strings.ToLower(logFormatText):
			logHandler = slog.NewTextHandler(logOutput, &slog.HandlerOptions{Level: logLevel})
		default:
			panic(fmt.Sprintf("invalid log format provided: %s, available formats: %s, %s",
				userProvidedLogFormat, logFormatJSON, logFormatText))
		}
	}

	logger := slog.New(logHandler)
	fixedLogger := logger.With(slog.String("service", serviceName))
	slog.SetDefault(fixedLogger)

	return fixedLogger
}

// setupMultiLoggerWithOTLP sets up logging to both console and OTLP
func setupMultiLoggerWithOTLP(otlpLogger *slog.Logger) *slog.Logger {
	consoleLogger := setupConsoleLogger()
	consoleHandler := consoleLogger.Handler()
	otlpHandler := otlpLogger.Handler()
	multiHandler := &multiHandler{
		handlers: []slog.Handler{consoleHandler, otlpHandler},
	}

	logger := slog.New(multiHandler)
	fixedLogger := logger.With(slog.String("service", serviceName))
	slog.SetDefault(fixedLogger)

	return fixedLogger
}

// multiHandler implements slog.Handler to write to multiple handlers
type multiHandler struct {
	handlers []slog.Handler
}

func (h *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}

	return false
}

func (h *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	var errs []error
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, r.Level) {
			if err := handler.Handle(ctx, r); err != nil {
				errs = append(errs, err)
			}
		}
	}

	if len(errs) > 0 {
		return errs[0]
	}

	return nil
}

func (h *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithAttrs(attrs)
	}

	return &multiHandler{handlers: newHandlers}
}

func (h *multiHandler) WithGroup(name string) slog.Handler {
	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithGroup(name)
	}

	return &multiHandler{handlers: newHandlers}
}

// setupOTLPLogger sets up OpenTelemetry OTLP logging
// Automatically detects protocol: HTTP for full URLs, gRPC for host:port
func setupOTLPLogger(ctx context.Context) *slog.Logger {
	otlpEndpoint := os.Getenv(env.OTEL_EXPORTER_OTLP_ENDPOINT)
	isHTTP := strings.HasPrefix(otlpEndpoint, "http://") || strings.HasPrefix(otlpEndpoint, "https://")
	if isHTTP {
		return setupOTLPLoggerHTTP(ctx, otlpEndpoint)
	}

	return setupOTLPLoggerGRPC(ctx, otlpEndpoint)
}

// setupOTLPLoggerWithProtocol sets up OpenTelemetry OTLP logging with explicit protocol
func setupOTLPLoggerWithProtocol(ctx context.Context, protocol string) *slog.Logger {
	otlpEndpoint := os.Getenv(env.OTEL_EXPORTER_OTLP_ENDPOINT)

	switch protocol {
	case "http":
		return setupOTLPLoggerHTTP(ctx, otlpEndpoint)
	case "grpc":
		return setupOTLPLoggerGRPC(ctx, otlpEndpoint)
	default:
		return setupOTLPLogger(ctx)
	}
}

// setupOTLPLoggerHTTP sets up OTLP logging via HTTP
func setupOTLPLoggerHTTP(ctx context.Context, otlpEndpoint string) *slog.Logger {
	var endpointHost string
	var urlPath string
	useSecure := true

	if strings.HasPrefix(otlpEndpoint, "http://") {
		otlpEndpoint = strings.TrimPrefix(otlpEndpoint, "http://")
		useSecure = false
	} else if strings.HasPrefix(otlpEndpoint, "https://") {
		otlpEndpoint = strings.TrimPrefix(otlpEndpoint, "https://")
		useSecure = true
	}

	// Split endpoint and path
	parts := strings.SplitN(otlpEndpoint, "/", 2)
	endpointHost = parts[0]
	if len(parts) > 1 {
		urlPath = "/" + parts[1]
	}

	// Remove /v1/logs suffix if present in the path, as otlploghttp will add it automatically
	urlPath = strings.TrimSuffix(urlPath, "/v1/logs")
	headers := make(map[string]string)
	if apiToken := os.Getenv(env.OTEL_EXPORTER_OTLP_HEADERS_AUTHORIZATION); apiToken != "" {
		headers["Authorization"] = apiToken
	}

	opts := []otlploghttp.Option{
		otlploghttp.WithEndpoint(endpointHost),
		otlploghttp.WithHeaders(headers),
	}

	// Use insecure (HTTP) only if explicitly specified with http:// scheme
	if !useSecure {
		opts = append(opts, otlploghttp.WithInsecure())
	}

	// Add custom URL path if present
	if urlPath != "" {
		opts = append(opts, otlploghttp.WithURLPath(urlPath+"/v1/logs"))
	}

	logExporter, err := otlploghttp.New(ctx, opts...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create OTLP HTTP log exporter, falling back to console logging: %v\n", err)
		return setupConsoleLogger()
	}

	logProcessor := log.NewBatchProcessor(logExporter)
	loggerProvider := log.NewLoggerProvider(
		log.WithProcessor(logProcessor),
	)
	global.SetLoggerProvider(loggerProvider)
	handler := otelslog.NewHandler(serviceName, otelslog.WithLoggerProvider(loggerProvider))
	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logger
}

// setupOTLPLoggerGRPC sets up OTLP logging via gRPC
func setupOTLPLoggerGRPC(ctx context.Context, otlpEndpoint string) *slog.Logger {
	var endpointHost string
	useSecure := true

	if strings.HasPrefix(otlpEndpoint, "http://") {
		otlpEndpoint = strings.TrimPrefix(otlpEndpoint, "http://")
		useSecure = false
	} else if strings.HasPrefix(otlpEndpoint, "https://") {
		otlpEndpoint = strings.TrimPrefix(otlpEndpoint, "https://")
		useSecure = true
	}

	// Split endpoint and path - for gRPC we only need the host:port part
	parts := strings.SplitN(otlpEndpoint, "/", 2)
	endpointHost = parts[0]
	headers := make(map[string]string)
	if apiToken := os.Getenv(env.OTEL_EXPORTER_OTLP_HEADERS_AUTHORIZATION); apiToken != "" {
		headers["authorization"] = apiToken
	}

	opts := []otlploggrpc.Option{
		otlploggrpc.WithEndpoint(endpointHost),
		otlploggrpc.WithHeaders(headers),
	}

	// Use insecure (HTTP/2 without TLS) only if explicitly specified with http:// scheme
	if !useSecure {
		opts = append(opts, otlploggrpc.WithInsecure())
	}
	logExporter, err := otlploggrpc.New(ctx, opts...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create OTLP gRPC log exporter, falling back to console logging: %v\n", err)
		return setupConsoleLogger()
	}
	logProcessor := log.NewBatchProcessor(logExporter)
	loggerProvider := log.NewLoggerProvider(
		log.WithProcessor(logProcessor),
	)
	global.SetLoggerProvider(loggerProvider)
	handler := otelslog.NewHandler(serviceName, otelslog.WithLoggerProvider(loggerProvider))
	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logger
}
