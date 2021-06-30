package authorization

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"testing"

	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization/clientcert"

	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization/testconsts"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	certificate = []byte(testconsts.Certificate)
	privateKey  = []byte(testconsts.PrivateKey)
)

func TestCertificateGenStrategy(t *testing.T) {

	t.Run("should add certificates to proxy", func(t *testing.T) {
		// given
		clientCert := clientcert.NewClientCertificate(nil)
		certGenStrategy := newCertificateGenStrategy(certificate, privateKey)

		request, err := http.NewRequest("GET", "www.example.com", nil)
		require.NoError(t, err)

		// when
		err = certGenStrategy.AddAuthorization(request, func(cert *tls.Certificate) {
			clientCert.SetCertificate(cert)
		})
		require.NoError(t, err)

		// then
		assert.Equal(t, &tls.Certificate{
			Certificate: [][]byte{cert()},
			PrivateKey:  key(),
		}, clientCert.GetCertificate())
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
