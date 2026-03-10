package command

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/otel"
	cryptobrokerclientgo "github.com/open-crypto-broker/crypto-broker-client-go"
)

func BenchmarkSign_profile_Default_CSR_SECP256R1_CA_RSA4096_Sequential(b *testing.B) {
	ctx := context.Background()
	logger := slog.New(
		slog.NewTextHandler(
			io.Discard, &slog.HandlerOptions{
				AddSource: false,
			},
		),
	)
	tracerProvider, err := otel.NewTracerProvider(ctx, logger, "", "")
	if err != nil {
		b.Fatalf("could not instantiate tracer provider, err: %s", err.Error())
	}
	lib, err := cryptobrokerclientgo.NewLibrary(ctx)
	if err != nil {
		b.Fatalf("could not instantiate library, err: %s", err.Error())
	}
	signCmd, err := NewSign(ctx, lib, logger, tracerProvider)
	if err != nil {
		b.Fatalf("could not instantiate sign, err: %s", err.Error())
	}

	payload := cryptobrokerclientgo.SignCertificatePayload{
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
		CAPrivateKey: []byte(`-----BEGIN EC PRIVATE KEY-----
MIHcAgEBBEIAsaSvwGS0nfPXCBX7MY0nt2VYYkOrf1dygvH8oIxyDE9LyWJ7eDBx
T77tKXW71fO1Kq0WOcocNp89wg6PMsUFZxWgBwYFK4EEACOhgYkDgYYABAERlddb
QZRNFQU21lb8jJUpjaS2UG2TH3CFdFxmCwFo66LI7NF6KgAbculBz4++FbD7fcb0
DtjzHrdJ+nj4OUaRYwD4jjv8Z7gEiQ9GYM8hPsyvAXJbbMsiUo+lcsXNWa4a7ZmG
YPEvJDRcZOaQELgCfS90jAPT45yefLkIsgEWq45bKA==
-----END EC PRIVATE KEY-----`),
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
		err := signCmd.signCertificate(ctx, payload, "PEM")
		if err != nil {
			b.Fatalf("could not run sign, err: %s", err.Error())
		}
	}
}

func BenchmarkSign_profile_Default_CSR_SECP256R1_CA_RSA4096_Parallel(b *testing.B) {
	ctx := context.Background()
	logger := slog.New(
		slog.NewTextHandler(
			io.Discard, &slog.HandlerOptions{
				AddSource: false,
			},
		),
	)
	tracerProvider, err := otel.NewTracerProvider(ctx, logger, "", "")
	if err != nil {
		b.Fatalf("could not instantiate tracer provider, err: %s", err.Error())
	}

	b.RunParallel(func(p *testing.PB) {
		lib, err := cryptobrokerclientgo.NewLibrary(ctx)
		if err != nil {
			b.Fatalf("could not instantiate library, err: %s", err.Error())
		}

		signCmd, err := NewSign(ctx, lib, logger, tracerProvider)
		if err != nil {
			b.Fatalf("could not instantiate sign, err: %s", err.Error())
		}

		payload := cryptobrokerclientgo.SignCertificatePayload{
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
			CAPrivateKey: []byte(`-----BEGIN EC PRIVATE KEY-----
MIHcAgEBBEIAsaSvwGS0nfPXCBX7MY0nt2VYYkOrf1dygvH8oIxyDE9LyWJ7eDBx
T77tKXW71fO1Kq0WOcocNp89wg6PMsUFZxWgBwYFK4EEACOhgYkDgYYABAERlddb
QZRNFQU21lb8jJUpjaS2UG2TH3CFdFxmCwFo66LI7NF6KgAbculBz4++FbD7fcb0
DtjzHrdJ+nj4OUaRYwD4jjv8Z7gEiQ9GYM8hPsyvAXJbbMsiUo+lcsXNWa4a7ZmG
YPEvJDRcZOaQELgCfS90jAPT45yefLkIsgEWq45bKA==
-----END EC PRIVATE KEY-----`),
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
			err := signCmd.signCertificate(ctx, payload, "PEM")
			if err != nil {
				b.Fatalf("could not run sign, err: %s", err.Error())
			}
		}
	})
}

func BenchmarkSign_profile_Default_CSR_SECP521R1_CA_SECP521R1_Sequential(b *testing.B) {
	ctx := context.Background()
	logger := slog.New(
		slog.NewTextHandler(
			io.Discard, &slog.HandlerOptions{
				AddSource: false,
			},
		),
	)
	tracerProvider, err := otel.NewTracerProvider(ctx, logger, "", "")
	if err != nil {
		b.Fatalf("could not instantiate tracer provider, err: %s", err.Error())
	}
	lib, err := cryptobrokerclientgo.NewLibrary(ctx)
	if err != nil {
		b.Fatalf("could not instantiate library, err: %s", err.Error())
	}
	signCmd, err := NewSign(ctx, lib, logger, tracerProvider)
	if err != nil {
		b.Fatalf("could not instantiate sign, err: %s", err.Error())
	}

	payload := cryptobrokerclientgo.SignCertificatePayload{
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
		CAPrivateKey: []byte(`-----BEGIN EC PRIVATE KEY-----
MIHcAgEBBEIAsaSvwGS0nfPXCBX7MY0nt2VYYkOrf1dygvH8oIxyDE9LyWJ7eDBx
T77tKXW71fO1Kq0WOcocNp89wg6PMsUFZxWgBwYFK4EEACOhgYkDgYYABAERlddb
QZRNFQU21lb8jJUpjaS2UG2TH3CFdFxmCwFo66LI7NF6KgAbculBz4++FbD7fcb0
DtjzHrdJ+nj4OUaRYwD4jjv8Z7gEiQ9GYM8hPsyvAXJbbMsiUo+lcsXNWa4a7ZmG
YPEvJDRcZOaQELgCfS90jAPT45yefLkIsgEWq45bKA==
-----END EC PRIVATE KEY-----`),
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
		err := signCmd.signCertificate(ctx, payload, "PEM")
		if err != nil {
			b.Fatalf("could not run sign, err: %s", err.Error())
		}
	}
}

func BenchmarkSign_profile_Default_CSR_SECP521R1_CA_SECP521R1_Parallel(b *testing.B) {
	ctx := context.Background()
	logger := slog.New(
		slog.NewTextHandler(
			io.Discard, &slog.HandlerOptions{
				AddSource: false,
			},
		),
	)
	tracerProvider, err := otel.NewTracerProvider(ctx, logger, "", "")
	if err != nil {
		b.Fatalf("could not instantiate tracer provider, err: %s", err.Error())
	}

	b.RunParallel(func(p *testing.PB) {
		lib, err := cryptobrokerclientgo.NewLibrary(ctx)
		if err != nil {
			b.Fatalf("could not instantiate library, err: %s", err.Error())
		}

		signCmd, err := NewSign(ctx, lib, logger, tracerProvider)
		if err != nil {
			b.Fatalf("could not instantiate sign, err: %s", err.Error())
		}

		payload := cryptobrokerclientgo.SignCertificatePayload{
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
			CAPrivateKey: []byte(`-----BEGIN EC PRIVATE KEY-----
MIHcAgEBBEIAsaSvwGS0nfPXCBX7MY0nt2VYYkOrf1dygvH8oIxyDE9LyWJ7eDBx
T77tKXW71fO1Kq0WOcocNp89wg6PMsUFZxWgBwYFK4EEACOhgYkDgYYABAERlddb
QZRNFQU21lb8jJUpjaS2UG2TH3CFdFxmCwFo66LI7NF6KgAbculBz4++FbD7fcb0
DtjzHrdJ+nj4OUaRYwD4jjv8Z7gEiQ9GYM8hPsyvAXJbbMsiUo+lcsXNWa4a7ZmG
YPEvJDRcZOaQELgCfS90jAPT45yefLkIsgEWq45bKA==
-----END EC PRIVATE KEY-----`),
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
			err := signCmd.signCertificate(ctx, payload, "PEM")
			if err != nil {
				b.Fatalf("could not run sign, err: %s", err.Error())
			}
		}
	})
}

func BenchmarkSign_profile_Default_CSR_SECP256R1_CA_SECP384R1_Sequential(b *testing.B) {
	ctx := context.Background()
	logger := slog.New(
		slog.NewTextHandler(
			io.Discard, &slog.HandlerOptions{
				AddSource: false,
			},
		),
	)
	tracerProvider, err := otel.NewTracerProvider(ctx, logger, "", "")
	if err != nil {
		b.Fatalf("could not instantiate tracer provider, err: %s", err.Error())
	}
	lib, err := cryptobrokerclientgo.NewLibrary(ctx)
	if err != nil {
		b.Fatalf("could not instantiate library, err: %s", err.Error())
	}
	signCmd, err := NewSign(ctx, lib, logger, tracerProvider)
	if err != nil {
		b.Fatalf("could not instantiate sign, err: %s", err.Error())
	}

	payload := cryptobrokerclientgo.SignCertificatePayload{
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
		CAPrivateKey: []byte(`-----BEGIN EC PRIVATE KEY-----
MIGkAgEBBDD+E65pqiUQUcKZCjrLOlpg+EMDUU+RIQmDIIilUzTim94OrhKKB/z4
OM25YzcvwQ6gBwYFK4EEACKhZANiAATS59LZWhfYdy/WpKGVk/xNfyzHh8GYTx1r
tXtrzLrNz8vpvYxfayUUDyhVV+J/aoY4tSUAVj+x3yAM2RXZLtJJihW6UiTyXEXF
+azQXNRDVkit8IQi53+KZDR0ECdsRBI=
-----END EC PRIVATE KEY-----`),
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
		err := signCmd.signCertificate(ctx, payload, "PEM")
		if err != nil {
			b.Fatalf("could not run sign, err: %s", err.Error())
		}
	}
}

func BenchmarkSign_profile_Default_CSR_SECP256R1_CA_SECP384R1_Parallel(b *testing.B) {
	ctx := context.Background()
	logger := slog.New(
		slog.NewTextHandler(
			io.Discard, &slog.HandlerOptions{
				AddSource: false,
			},
		),
	)
	tracerProvider, err := otel.NewTracerProvider(ctx, logger, "", "")
	if err != nil {
		b.Fatalf("could not instantiate tracer provider, err: %s", err.Error())
	}

	b.RunParallel(func(p *testing.PB) {
		lib, err := cryptobrokerclientgo.NewLibrary(ctx)
		if err != nil {
			b.Fatalf("could not instantiate library, err: %s", err.Error())
		}

		signCmd, err := NewSign(ctx, lib, logger, tracerProvider)
		if err != nil {
			b.Fatalf("could not instantiate sign, err: %s", err.Error())
		}

		payload := cryptobrokerclientgo.SignCertificatePayload{
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
			CAPrivateKey: []byte(`-----BEGIN EC PRIVATE KEY-----
MIGkAgEBBDD+E65pqiUQUcKZCjrLOlpg+EMDUU+RIQmDIIilUzTim94OrhKKB/z4
OM25YzcvwQ6gBwYFK4EEACKhZANiAATS59LZWhfYdy/WpKGVk/xNfyzHh8GYTx1r
tXtrzLrNz8vpvYxfayUUDyhVV+J/aoY4tSUAVj+x3yAM2RXZLtJJihW6UiTyXEXF
+azQXNRDVkit8IQi53+KZDR0ECdsRBI=
-----END EC PRIVATE KEY-----`),
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
			err := signCmd.signCertificate(ctx, payload, "PEM")
			if err != nil {
				b.Fatalf("could not run sign, err: %s", err.Error())
			}
		}
	})
}
