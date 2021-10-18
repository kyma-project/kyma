package certificates_test

import (
	"crypto/tls"
	"crypto/x509/pkix"
	"github.com/kyma-project/kyma/components/application-connector/application-registry/internal/metadata/certificates"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateKeyAndCertificate(t *testing.T) {
	t.Run("should return certificate and key pair", func(t *testing.T) {
		// given
		subject := pkix.Name{
			CommonName: "commonName",
		}

		// when
		keyCertPair, apperr := certificates.GenerateKeyAndCertificate(subject)

		// then
		require.NoError(t, apperr)
		require.NotNil(t, keyCertPair)

		cert, err := tls.X509KeyPair(keyCertPair.Certificate, keyCertPair.PrivateKey)
		require.NoError(t, err)
		require.Nil(t, cert.Leaf)
	})

}
