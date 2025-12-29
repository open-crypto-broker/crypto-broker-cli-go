package otel

import "go.opentelemetry.io/otel/attribute"

var (
	AttributeRpcMethod                  = attribute.Key("rpc.method")
	AttributeCryptoProfile              = attribute.Key("crypto.profile")
	AttributeCryptoInputSize            = attribute.Key("crypto.input_size")
	AttributeCryptoHashAlgorithm        = attribute.Key("crypto.hash_algorithm")
	AttributeCryptoHashOutputSize       = attribute.Key("crypto.hash_output_size")
	AttributeCryptoSignedCertSize       = attribute.Key("crypto.signed_cert_size")
	AttributeCryptoBenchmarkResultsSize = attribute.Key("crypto.benchmark_results_size")
	AttributeCryptoCsrSize              = attribute.Key("crypto.csr_size")
	AttributeCryptoCaCertSize           = attribute.Key("crypto.ca_cert_size")
	AttributeCryptoCaKeySize            = attribute.Key("crypto.ca_key_size")
)
