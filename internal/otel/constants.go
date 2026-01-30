package otel

// default values for OTEL configurations
const (
	defaultServiceName    = "crypto-broker-cli-go"
	defaultServiceVersion = "unknown service version"
	defaultTracesExporter = "console"
)

// keys representing OTEL exporters
const (
	keyExporterOTLPGRPC = "otlpgrpc"
	keyExporterConsole  = "console"
	keyExporterOTLPHTTP = "otlphttp"
)

// sampler names
const (
	samplerAlwaysOn                = "always_on"
	samplerAlways                  = "always"
	samplerAlwaysOff               = "always_off"
	samplerNever                   = "never"
	samplerTraceIDRatio            = "traceidratio"
	samplerRatio                   = "ratio"
	samplerParentBasedAlwaysOn     = "parentbased_always_on"
	samplerParentBasedAlwaysOff    = "parentbased_always_off"
	samplerParentBasedTraceIDRatio = "parentbased_traceidratio"
)
