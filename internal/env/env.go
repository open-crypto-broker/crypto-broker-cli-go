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

	// OTEL_EXPORTER_OTLP_HEADERS_AUTHORIZATION is OpenTelemetry environment variable for OTLP HTTP exporter authorization.
	// For Dynatrace, use format: "Api-Token YOUR_API_TOKEN"
	// Example: OTEL_EXPORTER_OTLP_HEADERS_AUTHORIZATION="Api-Token dt0c01.xxx..."
	OTEL_EXPORTER_OTLP_HEADERS_AUTHORIZATION = "OTEL_EXPORTER_OTLP_HEADERS_AUTHORIZATION"

	// OTEL_LOGS_EXPORTER is OpenTelemetry environment variable that specifies the log exporter(s) to use.
	// Supports comma-separated values for multiple exporters.
	// Valid values: "console", "otlp", "otlphttp", "otlpgrpc", or combinations like "otlp,console".
	// If not set or empty, console logging will be used as default.
	// Examples: "console", "otlp", "otlphttp", "otlpgrpc", "otlphttp,console"
	OTEL_LOGS_EXPORTER = "OTEL_LOGS_EXPORTER"

	// LOG_LEVEL is environment variable that should contain log level.
	// Valid values are denoted in internal/clog package
	LOG_LEVEL = "CRYPTO_BROKER_LOG_LEVEL"

	// LOG_FORMAT is environment variable that should contain log format.
	// Valid values are denoted in internal/clog package
	LOG_FORMAT = "CRYPTO_BROKER_LOG_FORMAT"

	// LOG_OUTPUT is environment variable that should contain log output.
	// Valid values are denoted in internal/clog package
	LOG_OUTPUT = "CRYPTO_BROKER_LOG_OUTPUT"
)
