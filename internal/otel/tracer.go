package otel

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/env"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

var ServiceName string

// default values for OTEL configurations
const (
	defaultServiceName    = "crypto-broker-cli-go"
	defaultServiceVersion = "unknown service version"
	defaultTracesExporter = "console"
)

// keys representing OTEL exporters
const (
	keyExporterOTLP    = "otlp"
	keyExporterConsole = "console"
)

// TracerProvider holds the OpenTelemetry tracer provider
type TracerProvider struct {
	tp *sdktrace.TracerProvider
}

// GetGlobalTracer returns the global tracer for the service
func GetGlobalTracer(serviceName string) trace.Tracer {
	return otel.Tracer(serviceName)
}

// NewTracerProvider creates and initializes a new OpenTelemetry tracer provider
func NewTracerProvider(ctx context.Context, serviceName, serviceVersion string) (*TracerProvider, error) {
	if serviceName == "" {
		ServiceName = os.Getenv(env.OTEL_SERVICE_NAME)
		if ServiceName == "" {
			ServiceName = defaultServiceName
		}
	} else {
		ServiceName = serviceName
	}

	if serviceVersion == "" {
		serviceVersion = os.Getenv(env.OTEL_SERVICE_VERSION)
		if serviceVersion == "" {
			serviceVersion = defaultServiceVersion
		}
	}

	tracesExporter := os.Getenv(env.OTEL_TRACES_EXPORTER)
	if tracesExporter == "" {
		tracesExporter = defaultTracesExporter
	}

	exporterNames := strings.Split(strings.ToLower(tracesExporter), ",")
	for i, name := range exporterNames {
		exporterNames[i] = strings.TrimSpace(name)
	}

	var batchers []sdktrace.TracerProviderOption
	if slices.Contains(exporterNames, keyExporterOTLP) {
		otlpEndpoint := os.Getenv(env.OTEL_EXPORTER_OTLP_ENDPOINT)
		if otlpEndpoint == "" {
			return nil, fmt.Errorf("%s is not set", env.OTEL_EXPORTER_OTLP_ENDPOINT)
		}

		otlpExporter, err := otlptracegrpc.New(ctx,
			otlptracegrpc.WithEndpoint(otlpEndpoint),
			otlptracegrpc.WithInsecure(), // Use TLS in production
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
		}
		batchers = append(batchers, sdktrace.WithBatcher(otlpExporter))
		slog.Info("OTLP exporter configured", slog.String("endpoint", otlpEndpoint))
	}

	if slices.Contains(exporterNames, keyExporterConsole) {
		consoleExporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			return nil, fmt.Errorf("failed to create console exporter: %w", err)
		}
		batchers = append(batchers, sdktrace.WithBatcher(consoleExporter))
		slog.Info("Console exporter configured")
	}

	if len(batchers) == 0 {
		slog.Info("No valid exporters configured, using no-op tracer provider",
			slog.String("requested_exporters", tracesExporter))
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

	sampler := sdktrace.AlwaysSample()
	samplerName := os.Getenv(env.OTEL_TRACES_SAMPLER)
	if samplerName == "" {
		samplerName = "always_on"
	}

	switch samplerName {
	case "always_on", "always":
		sampler = sdktrace.AlwaysSample()
	case "always_off", "never":
		sampler = sdktrace.NeverSample()
	case "traceidratio", "ratio":
		samplingRatio := 1.0 // Default to 100%
		if ratio := os.Getenv(env.OTEL_TRACES_SAMPLER_ARG); ratio != "" {
			if parsedRatio, err := parseFloat64(ratio); err == nil && parsedRatio >= 0.0 && parsedRatio <= 1.0 {
				samplingRatio = parsedRatio
			}
		}
		sampler = sdktrace.TraceIDRatioBased(samplingRatio)
	case "parentbased_always_on":
		sampler = sdktrace.ParentBased(sdktrace.AlwaysSample())
	case "parentbased_always_off":
		sampler = sdktrace.ParentBased(sdktrace.NeverSample())
	case "parentbased_traceidratio":
		samplingRatio := 1.0 // Default to 100%
		if ratio := os.Getenv(env.OTEL_TRACES_SAMPLER_ARG); ratio != "" {
			if parsedRatio, err := parseFloat64(ratio); err == nil && parsedRatio >= 0.0 && parsedRatio <= 1.0 {
				samplingRatio = parsedRatio
			}
		}
		sampler = sdktrace.ParentBased(sdktrace.TraceIDRatioBased(samplingRatio))
	default:
		slog.Warn("Unknown OTEL_TRACES_SAMPLER value, using always_on", slog.String("sampler", samplerName))
		sampler = sdktrace.AlwaysSample()
	}

	options := append(batchers, sdktrace.WithResource(res), sdktrace.WithSampler(sampler))
	tp := sdktrace.NewTracerProvider(options...)
	otel.SetTracerProvider(tp)

	slog.Info("OpenTelemetry tracer provider initialized",
		slog.String("service_name", serviceName),
		slog.String("service_version", serviceVersion),
		slog.String("exporters", tracesExporter),
		slog.String("sampler", samplerName))

	return &TracerProvider{tp: tp}, nil
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
