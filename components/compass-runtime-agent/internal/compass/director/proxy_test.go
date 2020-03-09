package director

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestSingleJoiningSlash(t *testing.T) {
	testCases := map[string]struct {
		proxyPath      string
		requestPath    string
		expectedResult string
	}{
		"Request path is added to proxy with adding additional slash at the end": {
			proxyPath:      "director/graphql",
			requestPath:    "foo/bar",
			expectedResult: "director/graphql/foo/bar",
		},
		"Request path is added to proxy and keeping adding additional slash at the end": {
			proxyPath:      "director/graphql",
			requestPath:    "foo/bar/",
			expectedResult: "director/graphql/foo/bar/",
		},
		"Trim trailing slash if request path is empty with slash": {
			proxyPath:      "director/graphql",
			requestPath:    "/",
			expectedResult: "director/graphql",
		},
		"Do not add trailing slash if request path is empty without slash": {
			proxyPath:      "director/graphql",
			requestPath:    "",
			expectedResult: "director/graphql",
		},
	}

	for tN, tC := range testCases {
		t.Run(tN, func(t *testing.T) {
			result := singleJoiningSlash(tC.proxyPath, tC.requestPath)
			assert.Equal(t, tC.expectedResult, result)
		})
	}
}

func TestProxyServeHTTP(t *testing.T) {
	t.Run("return service unavailable error when proxy to director is not yet configured", func(t *testing.T) {
		// given
		var (
			expectedBody = "Proxy to Director is not initialized. Try again later.\n"
			expectedCode = http.StatusServiceUnavailable

			fixRequest  = httptest.NewRequest(http.MethodPost, "http://hakuna.com/matata", nil)
			spyResponse = httptest.NewRecorder()
		)

		proxy := NewProxy(ProxyConfig{})

		// when
		proxy.ServeHTTP(spyResponse, fixRequest)

		// then
		assert.Equal(t, expectedCode, spyResponse.Code)
		assert.Equal(t, expectedBody, spyResponse.Body.String())
	})

	t.Run("proxying to set target URL with the given certs", func(t *testing.T) {
		// given
		var (
			expectedBody = "Buenos Aires from Director!"
			expectedCode = http.StatusOK
			spyResponse  = httptest.NewRecorder()
			fixCert      = MustLoadFixCert(t)
		)

		var gotPeerCertificates []*x509.Certificate
		ts := NewTestTLSServer(func(w http.ResponseWriter, r *http.Request) {
			gotPeerCertificates = r.TLS.PeerCertificates
			w.Write([]byte(expectedBody))
		})
		fixRequest := httptest.NewRequest(http.MethodPost, ts.URL, nil)

		proxy := NewProxy(ProxyConfig{InsecureSkipVerify: true})

		// when
		err := proxy.SetURLAndCerts(ts.URL, fixCert)
		proxy.ServeHTTP(spyResponse, fixRequest)

		// then
		require.NoError(t, err)
		assert.Equal(t, expectedCode, spyResponse.Code)
		assert.Equal(t, expectedBody, spyResponse.Body.String())

		require.Len(t, gotPeerCertificates, 1)
		assert.Equal(t, fixCert.Certificate[0], gotPeerCertificates[0].Raw)

	})

}

var (
	// LocalhostCert is a PEM-encoded TLS cert with SAN IPs
	// "127.0.0.1" and "[::1]", expiring at Jan 29 16:00:00 2084 GMT.
	// generated from src/crypto/tls:
	// go run "$(go env GOROOT)/src/crypto/tls/generate_cert.go" --rsa-bits 1024 --host 127.0.0.1,::1,hakuna.com --ca --start-date "Jan 1 00:00:00 1970" --duration=1000000h
	LocalhostCert = []byte(`-----BEGIN CERTIFICATE-----
MIICEjCCAXugAwIBAgIQLUIDxCBL0Ach8QWQmeHMbzANBgkqhkiG9w0BAQsFADAS
MRAwDgYDVQQKEwdBY21lIENvMCAXDTcwMDEwMTAwMDAwMFoYDzIwODQwMTI5MTYw
MDAwWjASMRAwDgYDVQQKEwdBY21lIENvMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCB
iQKBgQDl/rFHdQLdq3ZKJytvciRdFVRC9DBHcRlepuOh6pySab8Ry0x2rDkBkjxF
KaRj22J2/v3l7Gb1rJ3FLd6WIXloFXDEBONLLpJGfFRPiLtZyLLL9HiQKaSZeOEZ
fX8jym1qRmhNW77woWKI9HPxeIaFuyhUdBWIzCwbGRcXXFHtAwIDAQABo2cwZTAO
BgNVHQ8BAf8EBAMCAqQwEwYDVR0lBAwwCgYIKwYBBQUHAwEwDwYDVR0TAQH/BAUw
AwEB/zAtBgNVHREEJjAkggpoYWt1bmEuY29thwR/AAABhxAAAAAAAAAAAAAAAAAA
AAABMA0GCSqGSIb3DQEBCwUAA4GBADG0q61EmPU+XvQ17jXckQ2/Z7+4eUAoK4YA
nipmw7Wa5OtQI5GNs8fkXx9VQikkqCNj4o0BhLRzuyGw9PCSlOG+WVp6SG1mxXUv
ZJz77gRf14cKVrrm6Sb1in4zLLCofFizFn4Cqxkp/omrBKDrgDtLIjZs7LJ7jAd0
S4rc/28b
-----END CERTIFICATE-----`)

	// LocalhostKey is the private key for localhostCert.
	LocalhostKey = []byte(`-----BEGIN PRIVATE KEY-----
MIICeAIBADANBgkqhkiG9w0BAQEFAASCAmIwggJeAgEAAoGBAOX+sUd1At2rdkon
K29yJF0VVEL0MEdxGV6m46HqnJJpvxHLTHasOQGSPEUppGPbYnb+/eXsZvWsncUt
3pYheWgVcMQE40sukkZ8VE+Iu1nIssv0eJAppJl44Rl9fyPKbWpGaE1bvvChYoj0
c/F4hoW7KFR0FYjMLBsZFxdcUe0DAgMBAAECgYEAugEzSoEdYjzrG6l1/Vmogwde
8A8ghIa5Z808x5RAMEEJX9C09DzwlY644457/q5MgcRTfoGj+wgxSGiCXZSQ483e
8TP4h5Xa6tnYl3db2EDAtu5vy4qiETX6xHERYJS+Zgde57fLrZaI9irgu8sitG5o
MJf/kQ+W0xQTceQE0aECQQD5oxt51YTB3GEdj6wvYKiCgbzWm3QDkeyRZKTFXubz
ufqsmMNP3/wg4U8wopU85vsF+QrXDPKyzqq7mH8hCuDLAkEA69trMCKTQ2vUSaAC
uPtocchYlcyGGQE9ZgNbIXMuIrN/9NKk/pJucu4IWs8inwgw2sdK7R+6AmnR4PII
Lxy1qQJBALQzdYYBB5AZUVFRgO3CTGHI3VPda2WYVLivefGvi++r9LPaokJqYUoq
2ks1UZ1g7xtkptqN0jQY004PytVDUPkCQAqOLlTgJ0EMMVr+K0EGF12IPtataY7y
7EGFgu2TTwxEhkR5rOKrwP+pwXv26zC82Bricmr8UYHMJJVxn4YkPckCQQCy/bgU
Cwl3216i8PjILFepa9Qlwr8/NlZVs+WXMu7MgqKish9S4JQRKKH2hTt8g1YyOdCk
zaVipCMSMaallH3u
-----END PRIVATE KEY-----`)
)

func MustLoadFixCert(t *testing.T) *tls.Certificate {
	t.Helper()
	cert, err := tls.X509KeyPair(LocalhostCert, LocalhostKey)
	require.NoError(t, err)
	return &cert
}

func NewTestTLSServer(handler func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	ts := httptest.NewUnstartedServer(http.HandlerFunc(handler))
	ts.TLS = &tls.Config{ClientAuth: tls.RequireAnyClientCert}
	ts.StartTLS()

	return ts
}
