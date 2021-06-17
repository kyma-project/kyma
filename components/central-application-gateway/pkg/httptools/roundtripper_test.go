package httptools

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"io/ioutil"
	"math/big"
	mathrand "math/rand"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization/clientcert"

	"github.com/stretchr/testify/require"
)

func TestRoundTripper(t *testing.T) {
	tests := []struct {
		name         string
		transport    http.RoundTripper
		startTLS     bool
		requestError bool
	}{
		{
			name:      "Default round tripper",
			transport: NewRoundTripper(),
		},
		{
			name:      "TLSSkipVerify false",
			transport: NewRoundTripper(WithTLSSkipVerify(false)),
		},
		{
			name:      "TLSSkipVerify true",
			transport: NewRoundTripper(WithTLSSkipVerify(true)),
		},
		{
			name:         "TLSSkipVerify false with TLS",
			transport:    NewRoundTripper(WithTLSSkipVerify(false)),
			startTLS:     true,
			requestError: true,
		},
		{
			name:      "TLSSkipVerify true with TLS\"",
			transport: NewRoundTripper(WithTLSSkipVerify(true)),
			startTLS:  true,
		},
		{
			name: "Empty GetClientCertificate response",
			transport: NewRoundTripper(WithGetClientCertificate(func(info *tls.CertificateRequestInfo) (*tls.Certificate, error) {
				return nil, nil
			})),
		},
		{
			name: "Empty GetClientCertificate response with TLS",
			transport: NewRoundTripper(WithTLSSkipVerify(true), WithGetClientCertificate(func(info *tls.CertificateRequestInfo) (*tls.Certificate, error) {
				return nil, nil
			})),
			startTLS: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// given
			ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			defer ts.Close()

			if tc.startTLS {
				ts.StartTLS()
			} else {
				ts.Start()
			}

			httpClient := &http.Client{
				Transport: tc.transport,
			}
			req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
			require.NoError(t, err)

			res, err := httpClient.Do(req)
			if tc.requestError {
				require.NotNil(t, err)
				return
			}
			require.NoError(t, err)

			_, err = ioutil.ReadAll(res.Body)
			_ = res.Body.Close()
			require.NoError(t, err)
			require.Equal(t, res.StatusCode, http.StatusOK)

		})
	}
}

func TestRoundTripperMTLS(t *testing.T) {
	caTLSCert, caX509Cert, err := newCA()
	require.NoError(t, err)

	serverCert, err := newCert(caTLSCert)
	require.NoError(t, err)

	clientCert, err := newCert(caTLSCert)
	require.NoError(t, err)

	certpool := x509.NewCertPool()
	certpool.AddCert(caX509Cert)

	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	ts.TLS = &tls.Config{
		Certificates: []tls.Certificate{*serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certpool,
	}
	ts.StartTLS()
	defer ts.Close()

	transport := NewRoundTripper(
		WithTLSConfig(&tls.Config{RootCAs: certpool}),
		WithGetClientCertificate(clientcert.NewClientCertificate(clientCert).GetClientCertificate))

	httpClient := &http.Client{
		Transport: transport,
	}

	req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
	require.NoError(t, err)

	res, err := httpClient.Do(req)
	require.NoError(t, err)

	_ = res.Body.Close()
	require.Equal(t, res.StatusCode, http.StatusOK)
}

func newCert(caCert *tls.Certificate) (*tls.Certificate, error) {
	certificate := &x509.Certificate{
		SerialNumber: big.NewInt(mathrand.Int63()),
		Subject: pkix.Name{
			CommonName: "localhost",
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.IP([]byte{127, 0, 0, 1})},
	}
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	ca, err := x509.ParseCertificate(caCert.Certificate[0])
	if err != nil {
		return nil, err
	}
	cert_b, err := x509.CreateCertificate(rand.Reader, certificate, ca, &priv.PublicKey, caCert.PrivateKey)
	if err != nil {
		return nil, err
	}
	return &tls.Certificate{
		Certificate: [][]byte{cert_b},
		PrivateKey:  priv,
	}, nil

}
func newCA() (*tls.Certificate, *x509.Certificate, error) {
	certificate := &x509.Certificate{
		SerialNumber: big.NewInt(20210616),
		Subject: pkix.Name{
			CommonName: "ca-cert",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	caPriv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	cert_b, err := x509.CreateCertificate(rand.Reader, certificate, certificate, &caPriv.PublicKey, caPriv)
	if err != nil {
		return nil, nil, err
	}
	x509Cert, err := x509.ParseCertificate(cert_b)
	if err != nil {
		return nil, nil, err
	}
	return &tls.Certificate{
		Certificate: [][]byte{cert_b},
		PrivateKey:  caPriv,
	}, x509Cert, nil
}
