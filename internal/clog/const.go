package clog

// predefined keywords representing log level, log format and log output
const (
	logLevelDebug = "debug"
	logLevelInfo  = "info"
	logLevelWarn  = "warn"
	logLevelError = "error"
)

// predefined keywords representing log format and log output
const (
	logFormatJSON = "json"
	logFormatText = "text"
)

// predefined keywords representing log output
const (
	logOutputStdout = "stdout"
	logOutputStderr = "stderr"
)

// predefined keywords representing log exporters
const (
	keywordExporterConsole  = "console"
	keywordExporterOTLPHTTP = "otlphttp"
	keywordExporterOTLPGRPC = "otlpgrpc"
	keywordExporterOTLP     = "otlp"
)

// predefined service name
const (
	serviceName = "crypto-broker-cli-go"
)
