package constant

// constants that represents keywords behind the flags of the CLI.
const (
	KeywordFlagProfile            = "profile"
	KeywordFlagLoop               = "loop"
	KeywordFlagOutputFormat       = "output-format"
	KeywordFlagEncoding           = "encoding"
	KeywordFlagSubject            = "subject"
	KeywordFlagFilePathCSR        = "csr"
	KeywordFlagFilePathCACert     = "caCert"
	KeywordFlagFilePathSigningKey = "caKey"
	KeywordFlagConnections        = "connections"
	KeywordFlagConcurrency        = "concurrency"
	KeywordFlagRequests           = "requests"
	KeywordFlagDuration           = "duration"
	KeywordFlagTimeout            = "timeout"
	KeywordFlagConnectTimeout     = "connect-timeout"
	KeywordFlagInput              = "input"
)

// constants that represents supported encodings.
const (
	EncodingPEM = "pem"
	EncodingDER = "der"
)

// constants that represents supported loop flag values.
const (
	MinLoopFlagValue = 1
	MaxLoopFlagValue = 1000
	NoLoopFlagValue  = -1000001
)

const ClientGoModulePath = "github.com/open-crypto-broker/crypto-broker-client-go"
