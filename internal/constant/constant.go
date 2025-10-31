package constant

// constants that represents keywords behind the flags of the CLI.
const (
	KeywordFlagProfile            = "profile"
	KeywordFlagLoop               = "loop"
	KeywordFlagEncoding           = "encoding"
	KeywordFlagSubject            = "subject"
	KeywordFlagFilePathCSR        = "csr"
	KeywordFlagFilePathCACert     = "caCert"
	KeywordFlagFilePathSigningKey = "caKey"
)

// constants that represents supported encodings.
const (
	EncodingPEM = "pem"
	EncodingB64 = "b64"
)

// constants that represents supported loop flag values.
const (
	MinLoopFlagValue = 1
	MaxLoopFlagValue = 1000
	NoLoopFlagValue  = -1000001
)
