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
	KeywordFlagKeySource          = "keySource"
	KeywordFlagKey                = "key"
	KeywordFlagNonce              = "nonce"
	KeywordFlagAAD                = "aad"
	KeywordFlagTag                = "tag"
)

// Constants that represent supported symmetric key source values.
const (
	KeySourceRaw   = "raw"
	KeySourceKeyID = "key-id"
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
