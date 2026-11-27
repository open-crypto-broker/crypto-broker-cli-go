package command

import (
	"context"
	"crypto/x509"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/otel"
	cryptobroker "github.com/open-crypto-broker/crypto-broker-client-go"
)

var benchmarkCAKeyP521 = []byte(
	"\x30\x81\xee\x02\x01\x00\x30\x10\x06\x07\x2a\x86\x48\xce\x3d\x02" +
		"\x01\x06\x05\x2b\x81\x04\x00\x23\x04\x81\xd6\x30\x81\xd3\x02\x01" +
		"\x01\x04\x42\x00\xb1\xa4\xaf\xc0\x64\xb4\x9d\xf3\xd7\x08\x15\xfb" +
		"\x31\x8d\x27\xb7\x65\x58\x62\x43\xab\x7f\x57\x72\x82\xf1\xfc\xa0" +
		"\x8c\x72\x0c\x4f\x4b\xc9\x62\x7b\x78\x30\x71\x4f\xbe\xed\x29\x75" +
		"\xbb\xd5\xf3\xb5\x2a\xad\x16\x39\xca\x1c\x36\x9f\x3d\xc2\x0e\x8f" +
		"\x32\xc5\x05\x67\x15\xa1\x81\x89\x03\x81\x86\x00\x04\x01\x11\x95" +
		"\xd7\x5b\x41\x94\x4d\x15\x05\x36\xd6\x56\xfc\x8c\x95\x29\x8d\xa4" +
		"\xb6\x50\x6d\x93\x1f\x70\x85\x74\x5c\x66\x0b\x01\x68\xeb\xa2\xc8" +
		"\xec\xd1\x7a\x2a\x00\x1b\x72\xe9\x41\xcf\x8f\xbe\x15\xb0\xfb\x7d" +
		"\xc6\xf4\x0e\xd8\xf3\x1e\xb7\x49\xfa\x78\xf8\x39\x46\x91\x63\x00" +
		"\xf8\x8e\x3b\xfc\x67\xb8\x04\x89\x0f\x46\x60\xcf\x21\x3e\xcc\xaf" +
		"\x01\x72\x5b\x6c\xcb\x22\x52\x8f\xa5\x72\xc5\xcd\x59\xae\x1a\xed" +
		"\x99\x86\x60\xf1\x2f\x24\x34\x5c\x64\xe6\x90\x10\xb8\x02\x7d\x2f" +
		"\x74\x8c\x03\xd3\xe3\x9c\x9e\x7c\xb9\x08\xb2\x01\x16\xab\x8e\x5b" +
		"\x28",
)

var benchmarkCAKeyP384 = []byte(
	"\x30\x81\xb6\x02\x01\x00\x30\x10\x06\x07\x2a\x86\x48\xce\x3d\x02" +
		"\x01\x06\x05\x2b\x81\x04\x00\x22\x04\x81\x9e\x30\x81\x9b\x02\x01" +
		"\x01\x04\x30\xfe\x13\xae\x69\xaa\x25\x10\x51\xc2\x99\x0a\x3a\xcb" +
		"\x3a\x5a\x60\xf8\x43\x03\x51\x4f\x91\x21\x09\x83\x20\x88\xa5\x53" +
		"\x34\xe2\x9b\xde\x0e\xae\x12\x8a\x07\xfc\xf8\x38\xcd\xb9\x63\x37" +
		"\x2f\xc1\x0e\xa1\x64\x03\x62\x00\x04\xd2\xe7\xd2\xd9\x5a\x17\xd8" +
		"\x77\x2f\xd6\xa4\xa1\x95\x93\xfc\x4d\x7f\x2c\xc7\x87\xc1\x98\x4f" +
		"\x1d\x6b\xb5\x7b\x6b\xcc\xba\xcd\xcf\xcb\xe9\xbd\x8c\x5f\x6b\x25" +
		"\x14\x0f\x28\x55\x57\xe2\x7f\x6a\x86\x38\xb5\x25\x00\x56\x3f\xb1" +
		"\xdf\x20\x0c\xd9\x15\xd9\x2e\xd2\x49\x8a\x15\xba\x52\x24\xf2\x5c" +
		"\x45\xc5\xf9\xac\xd0\x5c\xd4\x43\x56\x48\xad\xf0\x84\x22\xe7\x7f" +
		"\x8a\x64\x34\x74\x10\x27\x6c\x44\x12",
)

func TestBenchmarkCertificateKeysPKCS8DER(t *testing.T) {
	for _, key := range [][]byte{benchmarkCAKeyP521, benchmarkCAKeyP384} {
		if _, err := x509.ParsePKCS8PrivateKey(key); err != nil {
			t.Fatalf("invalid benchmark PKCS#8 DER key: %v", err)
		}
	}
}

func BenchmarkSignCertificate_profile_Default_CSR_SECP256R1_CA_RSA4096_Sequential(b *testing.B) {
	ctx := context.Background()
	logger := slog.New(
		slog.NewTextHandler(
			io.Discard, &slog.HandlerOptions{
				AddSource: false,
			},
		),
	)
	tracerProvider, err := otel.NewTracerProvider(ctx, logger)
	if err != nil {
		b.Fatalf("could not instantiate tracer provider, err: %s", err.Error())
	}
	lib, err := cryptobroker.NewLibrary(ctx)
	if err != nil {
		b.Fatalf("could not instantiate library, err: %s", err.Error())
	}
	signCrtCmd, err := NewSignCertificate(ctx, lib, logger, tracerProvider)
	if err != nil {
		b.Fatalf("could not instantiate sign, err: %s", err.Error())
	}

	payload := cryptobroker.SignCertificatePayload{
		Metadata: &cryptobroker.Metadata{
			TraceContext: &cryptobroker.TraceContext{
				CorrelationId: uuid.New().String(),
			},
		},
		Profile: "Default",
		CSR: []byte(`-----BEGIN CERTIFICATE REQUEST-----
MIIBXzCCAQUCAQAwgaIxCzAJBgNVBAYTAkRFMREwDwYDVQQKDAhUZXN0IE9yZzEl
MCMGA1UECwwcVGVzdCBPcmcgQ2VydGlmaWNhdGUgU2VydmljZTEMMAoGA1UECwwD
RGV2MSEwHwYDVQQLDBhzdGFnaW5nLWNlcnRpZmljYXRlcy0xMDExDTALBgNVBAcM
BHRlc3QxGTAXBgNVBAMMEHRlc3QtY29tbW9uLW5hbWUwWTATBgcqhkjOPQIBBggq
hkjOPQMBBwNCAATznFqF7j2Gbsvmv96hkY6WLYC4V0A/AHmxxaHYNyJlu5mJLHC0
b9jEuGIuifnfJjMUqPXbXVNkGp8lSIgyfJIJoAAwCgYIKoZIzj0EAwIDSAAwRQIh
ALeIB2/wKr3HtLRsmlYYoUAJPkw2vAXj9kiBUwhGw2hFAiBP9PTPCcOIZN50n9C0
NrPbJOOC/7QNdsuxmDFGEapyZg==
-----END CERTIFICATE REQUEST-----`),
		KeySource: cryptobroker.SignKeySource{Single: &cryptobroker.KeySource{RawKey: benchmarkCAKeyP521}},
		CACert: []byte(`-----BEGIN CERTIFICATE-----
MIIC7DCCAk2gAwIBAgIUcy7fW7YwJWYg5YC1VIK+27ly8yYwCgYIKoZIzj0EAwQw
fjELMAkGA1UEBhMCREUxEDAOBgNVBAgMB0JhdmFyaWExGjAYBgNVBAoMEVRlc3Qt
T3JnYW5pemF0aW9uMR0wGwYDVQQLDBRUZXN0LU9yZ2FuaXphdGlvbi1DQTEiMCAG
A1UEAwwZVGVzdC1Pcmdhbml6YXRpb24tUm9vdC1DQTAeFw0yMzAxMDEwMTAxMDFa
Fw0zMzAxMDEwMTAxMDFaMH4xCzAJBgNVBAYTAkRFMRAwDgYDVQQIDAdCYXZhcmlh
MRowGAYDVQQKDBFUZXN0LU9yZ2FuaXphdGlvbjEdMBsGA1UECwwUVGVzdC1Pcmdh
bml6YXRpb24tQ0ExIjAgBgNVBAMMGVRlc3QtT3JnYW5pemF0aW9uLVJvb3QtQ0Ew
gZswEAYHKoZIzj0CAQYFK4EEACMDgYYABAERlddbQZRNFQU21lb8jJUpjaS2UG2T
H3CFdFxmCwFo66LI7NF6KgAbculBz4++FbD7fcb0DtjzHrdJ+nj4OUaRYwD4jjv8
Z7gEiQ9GYM8hPsyvAXJbbMsiUo+lcsXNWa4a7ZmGYPEvJDRcZOaQELgCfS90jAPT
45yefLkIsgEWq45bKKNmMGQwEgYDVR0TAQH/BAgwBgEB/wIBATAOBgNVHQ8BAf8E
BAMCAYYwHQYDVR0OBBYEFCYxHAX0Wr6I9FIybAP6+p2xnPRyMB8GA1UdIwQYMBaA
FCYxHAX0Wr6I9FIybAP6+p2xnPRyMAoGCCqGSM49BAMEA4GMADCBiAJCAUgiYrF4
H6K3+1vqastXKjfhnv12eNOZuv+Awo0Q1RPqYHhZxF5x5gykw0clhgy6wfmqB+Km
dAHEn4LToNX0cl1oAkIB8Cv/F/7TJ0tJn0FpwtCBbNWzlUpz6TJj2wz5e4t80dzi
DKXl/HVVm/pvigXURZC+DzE90ztDcthH55yHm+sMhuE=
-----END CERTIFICATE-----`),
	}
	for b.Loop() {
		err := signCrtCmd.signCertificate(ctx, payload, "PEM")

		if err != nil && !errors.Is(err, cryptobroker.ErrCircuitOpen) {
			b.Fatalf("could not run hash, err: %s", err.Error())
		}
	}
}

func BenchmarkSignCertificate_profile_Default_CSR_SECP256R1_CA_RSA4096_Parallel(b *testing.B) {
	setTestConcurrency(b)

	ctx := context.Background()
	logger := slog.New(
		slog.NewTextHandler(
			io.Discard, &slog.HandlerOptions{
				AddSource: false,
			},
		),
	)
	tracerProvider, err := otel.NewTracerProvider(ctx, logger)
	if err != nil {
		b.Fatalf("could not instantiate tracer provider, err: %s", err.Error())
	}

	b.RunParallel(func(p *testing.PB) {
		lib, err := cryptobroker.NewLibrary(ctx)
		if err != nil {
			b.Fatalf("could not instantiate library, err: %s", err.Error())
		}

		signCrtCmd, err := NewSignCertificate(ctx, lib, logger, tracerProvider)
		if err != nil {
			b.Fatalf("could not instantiate sign, err: %s", err.Error())
		}

		payload := cryptobroker.SignCertificatePayload{
			Metadata: &cryptobroker.Metadata{
				TraceContext: &cryptobroker.TraceContext{
					CorrelationId: uuid.New().String(),
				},
			},
			Profile: "Default",
			CSR: []byte(`-----BEGIN CERTIFICATE REQUEST-----
MIIBXzCCAQUCAQAwgaIxCzAJBgNVBAYTAkRFMREwDwYDVQQKDAhUZXN0IE9yZzEl
MCMGA1UECwwcVGVzdCBPcmcgQ2VydGlmaWNhdGUgU2VydmljZTEMMAoGA1UECwwD
RGV2MSEwHwYDVQQLDBhzdGFnaW5nLWNlcnRpZmljYXRlcy0xMDExDTALBgNVBAcM
BHRlc3QxGTAXBgNVBAMMEHRlc3QtY29tbW9uLW5hbWUwWTATBgcqhkjOPQIBBggq
hkjOPQMBBwNCAATznFqF7j2Gbsvmv96hkY6WLYC4V0A/AHmxxaHYNyJlu5mJLHC0
b9jEuGIuifnfJjMUqPXbXVNkGp8lSIgyfJIJoAAwCgYIKoZIzj0EAwIDSAAwRQIh
ALeIB2/wKr3HtLRsmlYYoUAJPkw2vAXj9kiBUwhGw2hFAiBP9PTPCcOIZN50n9C0
NrPbJOOC/7QNdsuxmDFGEapyZg==
-----END CERTIFICATE REQUEST-----`),
			KeySource: cryptobroker.SignKeySource{Single: &cryptobroker.KeySource{RawKey: benchmarkCAKeyP521}},
			CACert: []byte(`-----BEGIN CERTIFICATE-----
MIIC7DCCAk2gAwIBAgIUcy7fW7YwJWYg5YC1VIK+27ly8yYwCgYIKoZIzj0EAwQw
fjELMAkGA1UEBhMCREUxEDAOBgNVBAgMB0JhdmFyaWExGjAYBgNVBAoMEVRlc3Qt
T3JnYW5pemF0aW9uMR0wGwYDVQQLDBRUZXN0LU9yZ2FuaXphdGlvbi1DQTEiMCAG
A1UEAwwZVGVzdC1Pcmdhbml6YXRpb24tUm9vdC1DQTAeFw0yMzAxMDEwMTAxMDFa
Fw0zMzAxMDEwMTAxMDFaMH4xCzAJBgNVBAYTAkRFMRAwDgYDVQQIDAdCYXZhcmlh
MRowGAYDVQQKDBFUZXN0LU9yZ2FuaXphdGlvbjEdMBsGA1UECwwUVGVzdC1Pcmdh
bml6YXRpb24tQ0ExIjAgBgNVBAMMGVRlc3QtT3JnYW5pemF0aW9uLVJvb3QtQ0Ew
gZswEAYHKoZIzj0CAQYFK4EEACMDgYYABAERlddbQZRNFQU21lb8jJUpjaS2UG2T
H3CFdFxmCwFo66LI7NF6KgAbculBz4++FbD7fcb0DtjzHrdJ+nj4OUaRYwD4jjv8
Z7gEiQ9GYM8hPsyvAXJbbMsiUo+lcsXNWa4a7ZmGYPEvJDRcZOaQELgCfS90jAPT
45yefLkIsgEWq45bKKNmMGQwEgYDVR0TAQH/BAgwBgEB/wIBATAOBgNVHQ8BAf8E
BAMCAYYwHQYDVR0OBBYEFCYxHAX0Wr6I9FIybAP6+p2xnPRyMB8GA1UdIwQYMBaA
FCYxHAX0Wr6I9FIybAP6+p2xnPRyMAoGCCqGSM49BAMEA4GMADCBiAJCAUgiYrF4
H6K3+1vqastXKjfhnv12eNOZuv+Awo0Q1RPqYHhZxF5x5gykw0clhgy6wfmqB+Km
dAHEn4LToNX0cl1oAkIB8Cv/F/7TJ0tJn0FpwtCBbNWzlUpz6TJj2wz5e4t80dzi
DKXl/HVVm/pvigXURZC+DzE90ztDcthH55yHm+sMhuE=
-----END CERTIFICATE-----`),
		}
		for p.Next() {
			err := signCrtCmd.signCertificate(ctx, payload, "PEM")

			if err != nil && !errors.Is(err, cryptobroker.ErrCircuitOpen) {
				b.Fatalf("could not run hash, err: %s", err.Error())
			}
		}
	})
}

func BenchmarkSignCertificate_profile_Default_CSR_SECP521R1_CA_SECP521R1_Sequential(b *testing.B) {
	ctx := context.Background()
	logger := slog.New(
		slog.NewTextHandler(
			io.Discard, &slog.HandlerOptions{
				AddSource: false,
			},
		),
	)
	tracerProvider, err := otel.NewTracerProvider(ctx, logger)
	if err != nil {
		b.Fatalf("could not instantiate tracer provider, err: %s", err.Error())
	}
	lib, err := cryptobroker.NewLibrary(ctx)
	if err != nil {
		b.Fatalf("could not instantiate library, err: %s", err.Error())
	}
	signCrtCmd, err := NewSignCertificate(ctx, lib, logger, tracerProvider)
	if err != nil {
		b.Fatalf("could not instantiate sign, err: %s", err.Error())
	}

	payload := cryptobroker.SignCertificatePayload{
		Metadata: &cryptobroker.Metadata{
			TraceContext: &cryptobroker.TraceContext{
				CorrelationId: uuid.New().String(),
			},
		},
		Profile: "Default",
		CSR: []byte(`-----BEGIN CERTIFICATE REQUEST-----
MIIB5zCCAUgCAQAwgaIxCzAJBgNVBAYTAkRFMREwDwYDVQQKDAhUZXN0IE9yZzEl
MCMGA1UECwwcVGVzdCBPcmcgQ2VydGlmaWNhdGUgU2VydmljZTEMMAoGA1UECwwD
RGV2MSEwHwYDVQQLDBhzdGFnaW5nLWNlcnRpZmljYXRlcy0xMDExDTALBgNVBAcM
BHRlc3QxGTAXBgNVBAMMEHRlc3QtY29tbW9uLW5hbWUwgZswEAYHKoZIzj0CAQYF
K4EEACMDgYYABAG4J9e+RaevFWbipbgZTrdvVWgjc11uGM/XTODgHZf3W08OnL3i
c91AC6m+ul7iRUKV7Feyf7jGuvR9xiqghfMR+wCaI9S0SoOff/M7JCDIDAcB6TVl
wY9xlUF9z25XXnGHq6v18AQ+kKGPZQJ8eZYQWqMo48hzbmAV8M7dzEmIaGcltqAA
MAoGCCqGSM49BAMEA4GMADCBiAJCAbck1OvqQkWqeRcBRQRwXDs2EEtLWMZJCGsO
gab0fPVu7Kh8nMW9pdk5/P1z5UpgpcZkSNQDduCxSDr1pnsTXtI3AkIBBRaUW2og
xz4as/yt+3tVfrJa9Yaf3TjDqlTlncA8kJ3hhsRX5U/dwEJv2/ZMO7MWh12XUrQL
8rifvki0agFlvWQ=
-----END CERTIFICATE REQUEST-----`),
		KeySource: cryptobroker.SignKeySource{Single: &cryptobroker.KeySource{RawKey: benchmarkCAKeyP521}},
		CACert: []byte(`-----BEGIN CERTIFICATE-----
MIIC7DCCAk2gAwIBAgIUcy7fW7YwJWYg5YC1VIK+27ly8yYwCgYIKoZIzj0EAwQw
fjELMAkGA1UEBhMCREUxEDAOBgNVBAgMB0JhdmFyaWExGjAYBgNVBAoMEVRlc3Qt
T3JnYW5pemF0aW9uMR0wGwYDVQQLDBRUZXN0LU9yZ2FuaXphdGlvbi1DQTEiMCAG
A1UEAwwZVGVzdC1Pcmdhbml6YXRpb24tUm9vdC1DQTAeFw0yMzAxMDEwMTAxMDFa
Fw0zMzAxMDEwMTAxMDFaMH4xCzAJBgNVBAYTAkRFMRAwDgYDVQQIDAdCYXZhcmlh
MRowGAYDVQQKDBFUZXN0LU9yZ2FuaXphdGlvbjEdMBsGA1UECwwUVGVzdC1Pcmdh
bml6YXRpb24tQ0ExIjAgBgNVBAMMGVRlc3QtT3JnYW5pemF0aW9uLVJvb3QtQ0Ew
gZswEAYHKoZIzj0CAQYFK4EEACMDgYYABAERlddbQZRNFQU21lb8jJUpjaS2UG2T
H3CFdFxmCwFo66LI7NF6KgAbculBz4++FbD7fcb0DtjzHrdJ+nj4OUaRYwD4jjv8
Z7gEiQ9GYM8hPsyvAXJbbMsiUo+lcsXNWa4a7ZmGYPEvJDRcZOaQELgCfS90jAPT
45yefLkIsgEWq45bKKNmMGQwEgYDVR0TAQH/BAgwBgEB/wIBATAOBgNVHQ8BAf8E
BAMCAYYwHQYDVR0OBBYEFCYxHAX0Wr6I9FIybAP6+p2xnPRyMB8GA1UdIwQYMBaA
FCYxHAX0Wr6I9FIybAP6+p2xnPRyMAoGCCqGSM49BAMEA4GMADCBiAJCAUgiYrF4
H6K3+1vqastXKjfhnv12eNOZuv+Awo0Q1RPqYHhZxF5x5gykw0clhgy6wfmqB+Km
dAHEn4LToNX0cl1oAkIB8Cv/F/7TJ0tJn0FpwtCBbNWzlUpz6TJj2wz5e4t80dzi
DKXl/HVVm/pvigXURZC+DzE90ztDcthH55yHm+sMhuE=
-----END CERTIFICATE-----`),
	}
	for b.Loop() {
		err := signCrtCmd.signCertificate(ctx, payload, "PEM")

		if err != nil && !errors.Is(err, cryptobroker.ErrCircuitOpen) {
			b.Fatalf("could not run hash, err: %s", err.Error())
		}
	}
}

func BenchmarkSignCertificate_profile_Default_CSR_SECP521R1_CA_SECP521R1_Parallel(b *testing.B) {
	setTestConcurrency(b)

	ctx := context.Background()
	logger := slog.New(
		slog.NewTextHandler(
			io.Discard, &slog.HandlerOptions{
				AddSource: false,
			},
		),
	)
	tracerProvider, err := otel.NewTracerProvider(ctx, logger)
	if err != nil {
		b.Fatalf("could not instantiate tracer provider, err: %s", err.Error())
	}

	b.RunParallel(func(p *testing.PB) {
		lib, err := cryptobroker.NewLibrary(ctx)
		if err != nil {
			b.Fatalf("could not instantiate library, err: %s", err.Error())
		}

		signCrtCmd, err := NewSignCertificate(ctx, lib, logger, tracerProvider)
		if err != nil {
			b.Fatalf("could not instantiate sign, err: %s", err.Error())
		}

		payload := cryptobroker.SignCertificatePayload{
			Metadata: &cryptobroker.Metadata{
				TraceContext: &cryptobroker.TraceContext{
					CorrelationId: uuid.New().String(),
				},
			},
			Profile: "Default",
			CSR: []byte(`-----BEGIN CERTIFICATE REQUEST-----
MIIB5zCCAUgCAQAwgaIxCzAJBgNVBAYTAkRFMREwDwYDVQQKDAhUZXN0IE9yZzEl
MCMGA1UECwwcVGVzdCBPcmcgQ2VydGlmaWNhdGUgU2VydmljZTEMMAoGA1UECwwD
RGV2MSEwHwYDVQQLDBhzdGFnaW5nLWNlcnRpZmljYXRlcy0xMDExDTALBgNVBAcM
BHRlc3QxGTAXBgNVBAMMEHRlc3QtY29tbW9uLW5hbWUwgZswEAYHKoZIzj0CAQYF
K4EEACMDgYYABAG4J9e+RaevFWbipbgZTrdvVWgjc11uGM/XTODgHZf3W08OnL3i
c91AC6m+ul7iRUKV7Feyf7jGuvR9xiqghfMR+wCaI9S0SoOff/M7JCDIDAcB6TVl
wY9xlUF9z25XXnGHq6v18AQ+kKGPZQJ8eZYQWqMo48hzbmAV8M7dzEmIaGcltqAA
MAoGCCqGSM49BAMEA4GMADCBiAJCAbck1OvqQkWqeRcBRQRwXDs2EEtLWMZJCGsO
gab0fPVu7Kh8nMW9pdk5/P1z5UpgpcZkSNQDduCxSDr1pnsTXtI3AkIBBRaUW2og
xz4as/yt+3tVfrJa9Yaf3TjDqlTlncA8kJ3hhsRX5U/dwEJv2/ZMO7MWh12XUrQL
8rifvki0agFlvWQ=
-----END CERTIFICATE REQUEST-----`),
			KeySource: cryptobroker.SignKeySource{Single: &cryptobroker.KeySource{RawKey: benchmarkCAKeyP521}},
			CACert: []byte(`-----BEGIN CERTIFICATE-----
MIIC7DCCAk2gAwIBAgIUcy7fW7YwJWYg5YC1VIK+27ly8yYwCgYIKoZIzj0EAwQw
fjELMAkGA1UEBhMCREUxEDAOBgNVBAgMB0JhdmFyaWExGjAYBgNVBAoMEVRlc3Qt
T3JnYW5pemF0aW9uMR0wGwYDVQQLDBRUZXN0LU9yZ2FuaXphdGlvbi1DQTEiMCAG
A1UEAwwZVGVzdC1Pcmdhbml6YXRpb24tUm9vdC1DQTAeFw0yMzAxMDEwMTAxMDFa
Fw0zMzAxMDEwMTAxMDFaMH4xCzAJBgNVBAYTAkRFMRAwDgYDVQQIDAdCYXZhcmlh
MRowGAYDVQQKDBFUZXN0LU9yZ2FuaXphdGlvbjEdMBsGA1UECwwUVGVzdC1Pcmdh
bml6YXRpb24tQ0ExIjAgBgNVBAMMGVRlc3QtT3JnYW5pemF0aW9uLVJvb3QtQ0Ew
gZswEAYHKoZIzj0CAQYFK4EEACMDgYYABAERlddbQZRNFQU21lb8jJUpjaS2UG2T
H3CFdFxmCwFo66LI7NF6KgAbculBz4++FbD7fcb0DtjzHrdJ+nj4OUaRYwD4jjv8
Z7gEiQ9GYM8hPsyvAXJbbMsiUo+lcsXNWa4a7ZmGYPEvJDRcZOaQELgCfS90jAPT
45yefLkIsgEWq45bKKNmMGQwEgYDVR0TAQH/BAgwBgEB/wIBATAOBgNVHQ8BAf8E
BAMCAYYwHQYDVR0OBBYEFCYxHAX0Wr6I9FIybAP6+p2xnPRyMB8GA1UdIwQYMBaA
FCYxHAX0Wr6I9FIybAP6+p2xnPRyMAoGCCqGSM49BAMEA4GMADCBiAJCAUgiYrF4
H6K3+1vqastXKjfhnv12eNOZuv+Awo0Q1RPqYHhZxF5x5gykw0clhgy6wfmqB+Km
dAHEn4LToNX0cl1oAkIB8Cv/F/7TJ0tJn0FpwtCBbNWzlUpz6TJj2wz5e4t80dzi
DKXl/HVVm/pvigXURZC+DzE90ztDcthH55yHm+sMhuE=
-----END CERTIFICATE-----`),
		}
		for p.Next() {
			err := signCrtCmd.signCertificate(ctx, payload, "PEM")

			if err != nil && !errors.Is(err, cryptobroker.ErrCircuitOpen) {
				b.Fatalf("could not run hash, err: %s", err.Error())
			}
		}
	})
}

func BenchmarkSignCertificate_profile_Default_CSR_SECP256R1_CA_SECP384R1_Sequential(b *testing.B) {
	ctx := context.Background()
	logger := slog.New(
		slog.NewTextHandler(
			io.Discard, &slog.HandlerOptions{
				AddSource: false,
			},
		),
	)
	tracerProvider, err := otel.NewTracerProvider(ctx, logger)
	if err != nil {
		b.Fatalf("could not instantiate tracer provider, err: %s", err.Error())
	}
	lib, err := cryptobroker.NewLibrary(ctx)
	if err != nil {
		b.Fatalf("could not instantiate library, err: %s", err.Error())
	}
	signCrtCmd, err := NewSignCertificate(ctx, lib, logger, tracerProvider)
	if err != nil {
		b.Fatalf("could not instantiate sign certificate, err: %s", err.Error())
	}

	payload := cryptobroker.SignCertificatePayload{
		Metadata: &cryptobroker.Metadata{
			TraceContext: &cryptobroker.TraceContext{
				CorrelationId: uuid.New().String(),
			},
		},
		Profile: "Default",
		CSR: []byte(`-----BEGIN CERTIFICATE REQUEST-----
MIIBXzCCAQUCAQAwgaIxCzAJBgNVBAYTAkRFMREwDwYDVQQKDAhUZXN0IE9yZzEl
MCMGA1UECwwcVGVzdCBPcmcgQ2VydGlmaWNhdGUgU2VydmljZTEMMAoGA1UECwwD
RGV2MSEwHwYDVQQLDBhzdGFnaW5nLWNlcnRpZmljYXRlcy0xMDExDTALBgNVBAcM
BHRlc3QxGTAXBgNVBAMMEHRlc3QtY29tbW9uLW5hbWUwWTATBgcqhkjOPQIBBggq
hkjOPQMBBwNCAATznFqF7j2Gbsvmv96hkY6WLYC4V0A/AHmxxaHYNyJlu5mJLHC0
b9jEuGIuifnfJjMUqPXbXVNkGp8lSIgyfJIJoAAwCgYIKoZIzj0EAwIDSAAwRQIh
ALeIB2/wKr3HtLRsmlYYoUAJPkw2vAXj9kiBUwhGw2hFAiBP9PTPCcOIZN50n9C0
NrPbJOOC/7QNdsuxmDFGEapyZg==
-----END CERTIFICATE REQUEST-----`),
		KeySource: cryptobroker.SignKeySource{Single: &cryptobroker.KeySource{RawKey: benchmarkCAKeyP384}},
		CACert: []byte(`-----BEGIN CERTIFICATE-----
MIICoTCCAiegAwIBAgIUWu/H/WSqYIKoo23VGupKI7txz+4wCgYIKoZIzj0EAwMw
fjELMAkGA1UEBhMCREUxEDAOBgNVBAgMB0JhdmFyaWExGjAYBgNVBAoMEVRlc3Qt
T3JnYW5pemF0aW9uMR0wGwYDVQQLDBRUZXN0LU9yZ2FuaXphdGlvbi1DQTEiMCAG
A1UEAwwZVGVzdC1Pcmdhbml6YXRpb24tUm9vdC1DQTAeFw0yMzAxMDEwMTAxMDFa
Fw0zMzAxMDEwMTAxMDFaMH4xCzAJBgNVBAYTAkRFMRAwDgYDVQQIDAdCYXZhcmlh
MRowGAYDVQQKDBFUZXN0LU9yZ2FuaXphdGlvbjEdMBsGA1UECwwUVGVzdC1Pcmdh
bml6YXRpb24tQ0ExIjAgBgNVBAMMGVRlc3QtT3JnYW5pemF0aW9uLVJvb3QtQ0Ew
djAQBgcqhkjOPQIBBgUrgQQAIgNiAATS59LZWhfYdy/WpKGVk/xNfyzHh8GYTx1r
tXtrzLrNz8vpvYxfayUUDyhVV+J/aoY4tSUAVj+x3yAM2RXZLtJJihW6UiTyXEXF
+azQXNRDVkit8IQi53+KZDR0ECdsRBKjZjBkMBIGA1UdEwEB/wQIMAYBAf8CAQEw
DgYDVR0PAQH/BAQDAgGGMB0GA1UdDgQWBBRv8eTneQUlEYFwISn4NaIzjm3wYTAf
BgNVHSMEGDAWgBRv8eTneQUlEYFwISn4NaIzjm3wYTAKBggqhkjOPQQDAwNoADBl
AjBdUV/yHjHq90/swrXl5DvfK2vQssqAAgfD6VvhpzKWlOwULmCIdjzd0DJ9BtF6
VqUCMQClUxcW/Pvl4+nj1WwGa9YdQY1qXAhRSUJBcRw6y7Ejr2NQ2zTN2KMM4FV2
f/KE4vY=
-----END CERTIFICATE-----`),
	}
	for b.Loop() {
		err := signCrtCmd.signCertificate(ctx, payload, "PEM")

		if err != nil && !errors.Is(err, cryptobroker.ErrCircuitOpen) {
			b.Fatalf("could not run hash, err: %s", err.Error())
		}
	}
}

func BenchmarkSignCertificate_profile_Default_CSR_SECP256R1_CA_SECP384R1_Parallel(b *testing.B) {
	setTestConcurrency(b)

	ctx := context.Background()
	logger := slog.New(
		slog.NewTextHandler(
			io.Discard, &slog.HandlerOptions{
				AddSource: false,
			},
		),
	)
	tracerProvider, err := otel.NewTracerProvider(ctx, logger)
	if err != nil {
		b.Fatalf("could not instantiate tracer provider, err: %s", err.Error())
	}

	b.RunParallel(func(p *testing.PB) {
		lib, err := cryptobroker.NewLibrary(ctx)
		if err != nil {
			b.Fatalf("could not instantiate library, err: %s", err.Error())
		}

		signCrtCmd, err := NewSignCertificate(ctx, lib, logger, tracerProvider)
		if err != nil {
			b.Fatalf("could not instantiate sign certificate, err: %s", err.Error())
		}

		payload := cryptobroker.SignCertificatePayload{
			Metadata: &cryptobroker.Metadata{
				TraceContext: &cryptobroker.TraceContext{
					CorrelationId: uuid.New().String(),
				},
			},
			Profile: "Default",
			CSR: []byte(`-----BEGIN CERTIFICATE REQUEST-----
MIIBXzCCAQUCAQAwgaIxCzAJBgNVBAYTAkRFMREwDwYDVQQKDAhUZXN0IE9yZzEl
MCMGA1UECwwcVGVzdCBPcmcgQ2VydGlmaWNhdGUgU2VydmljZTEMMAoGA1UECwwD
RGV2MSEwHwYDVQQLDBhzdGFnaW5nLWNlcnRpZmljYXRlcy0xMDExDTALBgNVBAcM
BHRlc3QxGTAXBgNVBAMMEHRlc3QtY29tbW9uLW5hbWUwWTATBgcqhkjOPQIBBggq
hkjOPQMBBwNCAATznFqF7j2Gbsvmv96hkY6WLYC4V0A/AHmxxaHYNyJlu5mJLHC0
b9jEuGIuifnfJjMUqPXbXVNkGp8lSIgyfJIJoAAwCgYIKoZIzj0EAwIDSAAwRQIh
ALeIB2/wKr3HtLRsmlYYoUAJPkw2vAXj9kiBUwhGw2hFAiBP9PTPCcOIZN50n9C0
NrPbJOOC/7QNdsuxmDFGEapyZg==
-----END CERTIFICATE REQUEST-----`),
			KeySource: cryptobroker.SignKeySource{Single: &cryptobroker.KeySource{RawKey: benchmarkCAKeyP384}},
			CACert: []byte(`-----BEGIN CERTIFICATE-----
MIICoTCCAiegAwIBAgIUWu/H/WSqYIKoo23VGupKI7txz+4wCgYIKoZIzj0EAwMw
fjELMAkGA1UEBhMCREUxEDAOBgNVBAgMB0JhdmFyaWExGjAYBgNVBAoMEVRlc3Qt
T3JnYW5pemF0aW9uMR0wGwYDVQQLDBRUZXN0LU9yZ2FuaXphdGlvbi1DQTEiMCAG
A1UEAwwZVGVzdC1Pcmdhbml6YXRpb24tUm9vdC1DQTAeFw0yMzAxMDEwMTAxMDFa
Fw0zMzAxMDEwMTAxMDFaMH4xCzAJBgNVBAYTAkRFMRAwDgYDVQQIDAdCYXZhcmlh
MRowGAYDVQQKDBFUZXN0LU9yZ2FuaXphdGlvbjEdMBsGA1UECwwUVGVzdC1Pcmdh
bml6YXRpb24tQ0ExIjAgBgNVBAMMGVRlc3QtT3JnYW5pemF0aW9uLVJvb3QtQ0Ew
djAQBgcqhkjOPQIBBgUrgQQAIgNiAATS59LZWhfYdy/WpKGVk/xNfyzHh8GYTx1r
tXtrzLrNz8vpvYxfayUUDyhVV+J/aoY4tSUAVj+x3yAM2RXZLtJJihW6UiTyXEXF
+azQXNRDVkit8IQi53+KZDR0ECdsRBKjZjBkMBIGA1UdEwEB/wQIMAYBAf8CAQEw
DgYDVR0PAQH/BAQDAgGGMB0GA1UdDgQWBBRv8eTneQUlEYFwISn4NaIzjm3wYTAf
BgNVHSMEGDAWgBRv8eTneQUlEYFwISn4NaIzjm3wYTAKBggqhkjOPQQDAwNoADBl
AjBdUV/yHjHq90/swrXl5DvfK2vQssqAAgfD6VvhpzKWlOwULmCIdjzd0DJ9BtF6
VqUCMQClUxcW/Pvl4+nj1WwGa9YdQY1qXAhRSUJBcRw6y7Ejr2NQ2zTN2KMM4FV2
f/KE4vY=
-----END CERTIFICATE-----`),
		}
		for p.Next() {
			err := signCrtCmd.signCertificate(ctx, payload, "PEM")

			if err != nil && !errors.Is(err, cryptobroker.ErrCircuitOpen) {
				b.Fatalf("could not run sign certificate, err: %s", err.Error())
			}
		}
	})
}
