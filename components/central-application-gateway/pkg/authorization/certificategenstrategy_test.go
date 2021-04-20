package authorization

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httputil"
	"testing"

	"github.com/kyma-project/kyma/components/application-gateway/pkg/authorization/testconsts"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	proxyStub = &httputil.ReverseProxy{}

	certificate = []byte(testconsts.Certificate)
	privateKey  = []byte(testconsts.PrivateKey)
)

func TestCertificateGenStrategy(t *testing.T) {

	t.Run("should add certificates to proxy", func(t *testing.T) {
		// given
		proxy := &httputil.ReverseProxy{}

		expectedProxy := &httputil.ReverseProxy{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					Certificates: []tls.Certificate{
						{
							Certificate: [][]byte{cert()},
							PrivateKey:  key(),
						},
					},
				},
			},
		}

		certGenStrategy := newCertificateGenStrategy(certificate, privateKey)

		request, err := http.NewRequest("GET", "www.example.com", nil)
		require.NoError(t, err)

		// when
		err = certGenStrategy.AddAuthorization(request, func(transport *http.Transport) {
			proxy.Transport = transport
		})

		// then
		require.NoError(t, err)
		assert.Equal(t, expectedProxy, proxy)
	})

	t.Run("should return error when key is invalid", func(t *testing.T) {
		// given
		certGenStrategy := newCertificateGenStrategy(certificate, []byte("invalid key"))

		request, err := http.NewRequest("GET", "www.example.com", nil)
		require.NoError(t, err)

		// when
		err = certGenStrategy.AddAuthorization(request, nil)

		// then
		require.Error(t, err)
	})

	t.Run("should return error when certificate is invalid", func(t *testing.T) {
		// given

		certGenStrategy := newCertificateGenStrategy([]byte("invalid cert"), privateKey)

		request, err := http.NewRequest("GET", "www.example.com", nil)
		require.NoError(t, err)

		// when
		err = certGenStrategy.AddAuthorization(request, nil)

		// then
		require.Error(t, err)
	})
}

func key() *rsa.PrivateKey {
	pemBlock, _ := pem.Decode(privateKey)
	key, _ := x509.ParsePKCS1PrivateKey(pemBlock.Bytes)
	return key
}

func cert() []byte {
	pemBlock, _ := pem.Decode(certificate)
	return pemBlock.Bytes
}
