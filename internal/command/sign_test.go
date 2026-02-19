package command

import (
	"context"
	"io"
	"log"
	"testing"

	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/otel"
	cryptobrokerclientgo "github.com/open-crypto-broker/crypto-broker-client-go"
)

func BenchmarkSign_Synchronously(b *testing.B) {
	ctx := context.Background()
	logger := log.New(io.Discard, "TEST: ", log.Ldate|log.Lmicroseconds)
	tracerProvider, err := otel.NewTracerProvider(ctx, logger, "crypto-broker-cli-go", "0.0.0")
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
hkjOPQMBBwNCAAQ48h5W8DkBTRbwfB2tHPKi3I4kzgcPuMPcOlh7C8vSiV13UszH
BiOloPCcl7+0hz1D8difRsdeya9sKLK2qR2soAAwCgYIKoZIzj0EAwIDSAAwRQIg
T2sYmyQws9zTgPv0HJcD/q5Uds5DmFoAM5D0LANNU8sCIQDT05wfvy7UEjKO2nX5
Bg9SEosO1TISv45Llcl4m7wkFQ==
-----END CERTIFICATE REQUEST-----`),
		CAPrivateKey: []byte(`-----BEGIN PRIVATE KEY-----
MIG2AgEAMBAGByqGSM49AgEGBSuBBAAiBIGeMIGbAgEBBDBGW8UiwRuSxxS/Rj5u
FRQvQo7miZG+e/f8veaUcMv5JM5mNi61GtzzQ1hiVArskxChZANiAATidJfbi35A
m+uXmcYKRsOOoi7YqqpQAI+RI8hMn66l2qVaTDWRlAI87u9iw1pvRoGH3nNrsiig
8nCxDr7mPzitAmMeBkFBZaTCFBstVZIDgrv3oZifwRvIaUY8Ppv7ntg=
-----END PRIVATE KEY-----`),
		CACert: []byte(`-----BEGIN CERTIFICATE-----
MIICoTCCAiegAwIBAgIUZv687AKMDfhBzPhtYqY841Zshf0wCgYIKoZIzj0EAwQw
fjELMAkGA1UEBhMCREUxEDAOBgNVBAgMB0JhdmFyaWExGjAYBgNVBAoMEVRlc3Qt
T3JnYW5pemF0aW9uMR0wGwYDVQQLDBRUZXN0LU9yZ2FuaXphdGlvbi1DQTEiMCAG
A1UEAwwZVGVzdC1Pcmdhbml6YXRpb24tUm9vdC1DQTAeFw0yMzAxMDEwMTAxMDFa
Fw0zMzAxMDEwMTAxMDFaMH4xCzAJBgNVBAYTAkRFMRAwDgYDVQQIDAdCYXZhcmlh
MRowGAYDVQQKDBFUZXN0LU9yZ2FuaXphdGlvbjEdMBsGA1UECwwUVGVzdC1Pcmdh
bml6YXRpb24tQ0ExIjAgBgNVBAMMGVRlc3QtT3JnYW5pemF0aW9uLVJvb3QtQ0Ew
djAQBgcqhkjOPQIBBgUrgQQAIgNiAATidJfbi35Am+uXmcYKRsOOoi7YqqpQAI+R
I8hMn66l2qVaTDWRlAI87u9iw1pvRoGH3nNrsiig8nCxDr7mPzitAmMeBkFBZaTC
FBstVZIDgrv3oZifwRvIaUY8Ppv7ntijZjBkMBIGA1UdEwEB/wQIMAYBAf8CAQEw
DgYDVR0PAQH/BAQDAgGGMB0GA1UdDgQWBBTiB5J+O82fGVW8oYbKI2lxR9yqfjAf
BgNVHSMEGDAWgBTiB5J+O82fGVW8oYbKI2lxR9yqfjAKBggqhkjOPQQDBANoADBl
AjAaaXME5CL0R65/hD+f5Zn5zRbzsIw1w88EnkgIw44kRd7M5N0HORiEGh+6jlt5
PsUCMQDEiwry2XAcLFZvxLfCmia4Qobs/EkaZVQ1fCcs6j3Z/mnslUJyobaIkDPa
G5MLQWA=
-----END CERTIFICATE-----`),
	}
	for b.Loop() {
		err := signCmd.signCertificate(ctx, payload, "PEM")
		if err != nil {
			b.Fatalf("could not run sign, err: %s", err.Error())
		}
	}
}

func BenchmarkSign_Asynchronously(b *testing.B) {
	ctx := context.Background()
	logger := log.New(io.Discard, "TEST: ", log.Ldate|log.Lmicroseconds)
	tracerProvider, err := otel.NewTracerProvider(ctx, logger, "crypto-broker-cli-go", "0.0.0")
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
hkjOPQMBBwNCAAQ48h5W8DkBTRbwfB2tHPKi3I4kzgcPuMPcOlh7C8vSiV13UszH
BiOloPCcl7+0hz1D8difRsdeya9sKLK2qR2soAAwCgYIKoZIzj0EAwIDSAAwRQIg
T2sYmyQws9zTgPv0HJcD/q5Uds5DmFoAM5D0LANNU8sCIQDT05wfvy7UEjKO2nX5
Bg9SEosO1TISv45Llcl4m7wkFQ==
-----END CERTIFICATE REQUEST-----
			`),
			CAPrivateKey: []byte(`-----BEGIN PRIVATE KEY-----
MIG2AgEAMBAGByqGSM49AgEGBSuBBAAiBIGeMIGbAgEBBDBGW8UiwRuSxxS/Rj5u
FRQvQo7miZG+e/f8veaUcMv5JM5mNi61GtzzQ1hiVArskxChZANiAATidJfbi35A
m+uXmcYKRsOOoi7YqqpQAI+RI8hMn66l2qVaTDWRlAI87u9iw1pvRoGH3nNrsiig
8nCxDr7mPzitAmMeBkFBZaTCFBstVZIDgrv3oZifwRvIaUY8Ppv7ntg=
-----END PRIVATE KEY-----
			`),
			CACert: []byte(`-----BEGIN CERTIFICATE-----
MIICoTCCAiegAwIBAgIUZv687AKMDfhBzPhtYqY841Zshf0wCgYIKoZIzj0EAwQw
fjELMAkGA1UEBhMCREUxEDAOBgNVBAgMB0JhdmFyaWExGjAYBgNVBAoMEVRlc3Qt
T3JnYW5pemF0aW9uMR0wGwYDVQQLDBRUZXN0LU9yZ2FuaXphdGlvbi1DQTEiMCAG
A1UEAwwZVGVzdC1Pcmdhbml6YXRpb24tUm9vdC1DQTAeFw0yMzAxMDEwMTAxMDFa
Fw0zMzAxMDEwMTAxMDFaMH4xCzAJBgNVBAYTAkRFMRAwDgYDVQQIDAdCYXZhcmlh
MRowGAYDVQQKDBFUZXN0LU9yZ2FuaXphdGlvbjEdMBsGA1UECwwUVGVzdC1Pcmdh
bml6YXRpb24tQ0ExIjAgBgNVBAMMGVRlc3QtT3JnYW5pemF0aW9uLVJvb3QtQ0Ew
djAQBgcqhkjOPQIBBgUrgQQAIgNiAATidJfbi35Am+uXmcYKRsOOoi7YqqpQAI+R
I8hMn66l2qVaTDWRlAI87u9iw1pvRoGH3nNrsiig8nCxDr7mPzitAmMeBkFBZaTC
FBstVZIDgrv3oZifwRvIaUY8Ppv7ntijZjBkMBIGA1UdEwEB/wQIMAYBAf8CAQEw
DgYDVR0PAQH/BAQDAgGGMB0GA1UdDgQWBBTiB5J+O82fGVW8oYbKI2lxR9yqfjAf
BgNVHSMEGDAWgBTiB5J+O82fGVW8oYbKI2lxR9yqfjAKBggqhkjOPQQDBANoADBl
AjAaaXME5CL0R65/hD+f5Zn5zRbzsIw1w88EnkgIw44kRd7M5N0HORiEGh+6jlt5
PsUCMQDEiwry2XAcLFZvxLfCmia4Qobs/EkaZVQ1fCcs6j3Z/mnslUJyobaIkDPa
G5MLQWA=
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