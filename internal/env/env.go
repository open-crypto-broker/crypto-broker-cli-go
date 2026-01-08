// Package env stands for environment variables. Contains constants for environment variables used by the Crypto Broker.
package env

const (
	// OTEL_TRACES_EXPORTER is OpenTelemetry environment variable that specifies the trace exporter(s) to use.
	// Can be "otlp", "console", "both", or comma-separated list like "console,otlp".
	OTEL_TRACES_EXPORTER = "OTEL_TRACES_EXPORTER"

	// OTEL_EXPORTER_OTLP_ENDPOINT is OpenTelemetry environment variable that specifies the OTLP endpoint.
	// For gRPC OTLP, use format "host:port". For HTTP OTLP, use "http://host:port".
	OTEL_EXPORTER_OTLP_ENDPOINT = "OTEL_EXPORTER_OTLP_ENDPOINT"

	// OTEL_TRACES_SAMPLER is OpenTelemetry environment variable that specifies the sampling strategy.
	// Valid values: always_on, always_off, traceidratio, parentbased_always_on, parentbased_always_off, parentbased_traceidratio.
	OTEL_TRACES_SAMPLER = "OTEL_TRACES_SAMPLER"

	// OTEL_TRACES_SAMPLER_ARG is OpenTelemetry environment variable that specifies sampling ratio (0.0-1.0)
	// when using ratio-based samplers like traceidratio or parentbased_traceidratio.
	OTEL_TRACES_SAMPLER_ARG = "OTEL_TRACES_SAMPLER_ARG"

	// OTEL_SERVICE_NAME is OpenTelemetry environment variable that specifies the service name for traces.
	OTEL_SERVICE_NAME = "OTEL_SERVICE_NAME"

	// OTEL_SERVICE_VERSION is OpenTelemetry environment variable that specifies the service version for traces.
	OTEL_SERVICE_VERSION = "OTEL_SERVICE_VERSION"
)
