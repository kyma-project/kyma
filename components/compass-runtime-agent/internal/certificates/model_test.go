package certificates

import (
	"crypto/rsa"
	"crypto/x509"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

func TestClientCredentials_AsTLSCertificate(t *testing.T) {

	pemCredentials := PemEncodedCredentials{
		ClientKey:         clientKey,
		CertificateChain:  crtChain,
		ClientCertificate: clientCRT,
		CACertificates:    caCRT,
	}

	credentials, err := pemCredentials.AsCredentials()
	require.NoError(t, err)

	tlsCert := credentials.AsTLSCertificate()
	require.NotEmpty(t, tlsCert)

	assert.NotEmpty(t, tlsCert.PrivateKey)
	assert.NotEmpty(t, tlsCert.Certificate)

	privKey, ok := tlsCert.PrivateKey.(*rsa.PrivateKey)
	assert.True(t, ok)
	assert.Equal(t, credentials.ClientKey, privKey)

	certs := make([]*x509.Certificate, 0, len(tlsCert.Certificate))
	for _, bytes := range tlsCert.Certificate {
		cert, err := x509.ParseCertificate(bytes)
		require.NoError(t, err)

		certs = append(certs, cert)
	}

	assert.Equal(t, 2, len(certs))
	assert.NotEmpty(t, certs[0])
	assert.NotEmpty(t, certs[1])
	assert.Equal(t, credentials.CertificateChain, certs)
}
