package otel

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"strings"

	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/env"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

// TracerProvider holds the OpenTelemetry tracer provider
type TracerProvider struct {
	tp *sdktrace.TracerProvider
}

// GetGlobalTracer returns the global tracer for the service
func GetGlobalTracer() trace.Tracer {
	return otel.Tracer(serviceName)
}

// NewTracerProvider creates and initializes a new OpenTelemetry tracer provider
func NewTracerProvider(ctx context.Context, logger *slog.Logger) (*TracerProvider, error) {
	exporterNames := strings.Split(strings.ToLower(tracesExporter), ",")
	for i, name := range exporterNames {
		exporterNames[i] = strings.TrimSpace(name)
	}

	var batchers []sdktrace.TracerProviderOption
	if slices.Contains(exporterNames, keyExporterOTLPHTTP) {
		if otlpEndpoint == "" {
			return nil, fmt.Errorf("%s is not set", env.OTEL_EXPORTER_OTLP_ENDPOINT)
		}

		batcherHTTP, err := getBatchersHTTP(ctx, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create HTTP OTLP exporter: %w", err)
		}

		batchers = append(batchers, batcherHTTP...)
	}

	if slices.Contains(exporterNames, keyExporterOTLPGRPC) {
		if otlpEndpoint == "" {
			return nil, fmt.Errorf("%s is not set", env.OTEL_EXPORTER_OTLP_ENDPOINT)
		}

		batcherGRPC, err := getBatchersGRPC(ctx, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create gRPC OTLP exporter: %w", err)
		}

		batchers = append(batchers, batcherGRPC...)
	}

	if slices.Contains(exporterNames, keyExporterConsole) {
		batcherConsole, err := getBatchersConsole(logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create console exporter: %w", err)
		}
		batchers = append(batchers, batcherConsole...)
	}

	if len(batchers) == 0 {
		logger.Info("No valid exporters configured, using no-op tracer provider",
			"requested_exporters", tracesExporter)
		tp := sdktrace.NewTracerProvider()
		otel.SetTracerProvider(tp)
		return &TracerProvider{tp: tp}, nil
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(serviceVersion),
			semconv.ServiceNamespaceKey.String("crypto-broker"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	sampler := defineSampler(logger)
	options := append(batchers, sdktrace.WithResource(res), sdktrace.WithSampler(sampler))
	tp := sdktrace.NewTracerProvider(options...)
	otel.SetTracerProvider(tp)

	logger.Info("OpenTelemetry tracer provider initialized",
		"service_version", serviceVersion,
		"exporters", tracesExporter,
		"sampler", sampler.Description())

	return &TracerProvider{tp: tp}, nil
}

// getBatchersHTTP creates a HTTP exporter
func getBatchersHTTP(ctx context.Context, logger *slog.Logger) ([]sdktrace.TracerProviderOption, error) {
	// Parse the endpoint URL to extract host:port and path
	// For Dynatrace: "https://trc17344.live.dynatrace.com/api/v2/otlp"
	// Endpoint should only be "trc17344.live.dynatrace.com"
	// Path should be "/api/v2/otlp"
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

	// Remove /v1/traces suffix if present in the path, as otlptracehttp will add it automatically
	urlPath = strings.TrimSuffix(urlPath, "/v1/traces")

	// Build headers for HTTP OTLP exporter
	headers := make(map[string]string)

	// Check for API token
	if apiToken != "" {
		headers["Authorization"] = apiToken
		logger.Info("Dynatrace API token configured for OTLP HTTP exporter")
	}

	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(endpointHost),
		otlptracehttp.WithHeaders(headers),
	}

	// Use insecure (HTTP) only if explicitly specified with http:// scheme
	// For HTTPS (recommended for production like Dynatrace), don't add WithInsecure()
	if !useSecure {
		// For HTTP endpoints (typically for local development only)
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	// Add custom URL path if present
	if urlPath != "" {
		opts = append(opts, otlptracehttp.WithURLPath(urlPath+"/v1/traces"))
	}

	otlpExporter, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP OTLP exporter: %w", err)
	}
	logger.Info("HTTP OTLP exporter configured", "endpoint", endpointHost, "path", urlPath)

	return []sdktrace.TracerProviderOption{sdktrace.WithBatcher(otlpExporter)}, nil
}

// getBatchersGRPC creates a gRPC exporter
func getBatchersGRPC(ctx context.Context, logger *slog.Logger) ([]sdktrace.TracerProviderOption, error) {
	otlpExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(otlpEndpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC OTLP exporter: %w", err)
	}
	logger.Info("gRPC OTLP exporter configured", "endpoint", otlpEndpoint)

	return []sdktrace.TracerProviderOption{sdktrace.WithBatcher(otlpExporter)}, nil
}

// getBatchersConsole creates a console exporter
func getBatchersConsole(logger *slog.Logger) ([]sdktrace.TracerProviderOption, error) {
	consoleExporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, fmt.Errorf("failed to create console exporter: %w", err)
	}
	logger.Info("Console exporter configured")

	return []sdktrace.TracerProviderOption{sdktrace.WithBatcher(consoleExporter)}, nil
}

// defineSampler defines the sampler for the tracer provider
func defineSampler(logger *slog.Logger) sdktrace.Sampler {
	var sampler sdktrace.Sampler 
	switch samplerName {
	case samplerAlwaysOn, samplerAlways:
		sampler = sdktrace.AlwaysSample()
	case samplerAlwaysOff, samplerNever:
		sampler = sdktrace.NeverSample()
	case samplerTraceIDRatio, samplerRatio:
		sampler = sdktrace.TraceIDRatioBased(samplingRatio)
	case samplerParentBasedAlwaysOn:
		sampler = sdktrace.ParentBased(sdktrace.AlwaysSample())
	case samplerParentBasedAlwaysOff:
		sampler = sdktrace.ParentBased(sdktrace.NeverSample())
	case samplerParentBasedTraceIDRatio:
		sampler = sdktrace.ParentBased(sdktrace.TraceIDRatioBased(samplingRatio))
	default:
		logger.Info("Unknown OTEL_TRACES_SAMPLER value, using always_on", "sampler", samplerName)
		sampler = sdktrace.AlwaysSample()
	}

	return sampler
}

// Shutdown gracefully shuts down the tracer provider
func (tp *TracerProvider) Shutdown(ctx context.Context) error {
	if tp.tp != nil {
		return tp.tp.Shutdown(ctx)
	}
	return nil
}

// GetTracer returns a tracer with the given name
func (tp *TracerProvider) GetTracer(name string, opts ...trace.TracerOption) trace.Tracer {
	return tp.tp.Tracer(name, opts...)
}

// parseFloat64 parses a string to float64
func parseFloat64(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}
