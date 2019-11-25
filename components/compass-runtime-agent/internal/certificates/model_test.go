package certificates

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"testing"

	gqlschema "github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

func TestClientCredentials_AsTLSCertificate(t *testing.T) {
	// given
	pemCredentials := PemEncodedCredentials{
		ClientKey:         clientKey,
		CertificateChain:  crtChain,
		ClientCertificate: clientCRT,
		CACertificates:    caCRT,
	}

	credentials, err := pemCredentials.AsCredentials()
	require.NoError(t, err)

	// when
	tlsCert := credentials.AsTLSCertificate()

	// then
	require.NotNil(t, tlsCert)
	require.NotEmpty(t, tlsCert.PrivateKey)
	require.NotEmpty(t, tlsCert.Certificate)

	privKey, ok := tlsCert.PrivateKey.(*rsa.PrivateKey)
	assert.True(t, ok)
	assert.Equal(t, credentials.ClientKey, privKey)

	certs := make([]*x509.Certificate, 0, len(tlsCert.Certificate))
	for _, bytes := range tlsCert.Certificate {
		cert, err := x509.ParseCertificate(bytes)
		require.NoError(t, err)

		certs = append(certs, cert)
	}

	require.Equal(t, 2, len(certs))
	assert.NotEmpty(t, certs[0])
	assert.NotEmpty(t, certs[1])
	assert.Equal(t, credentials.CertificateChain, certs)
}

func TestNewCredentials(t *testing.T) {
	// given
	expectedCredentials, err := PemEncodedCredentials{
		ClientKey:         clientKey,
		ClientCertificate: clientCRT,
		CertificateChain:  crtChain,
		CACertificates:    caCRT,
	}.AsCredentials()
	require.NoError(t, err)

	certificateResponse := gqlschema.CertificationResult{
		CertificateChain:  base64.StdEncoding.EncodeToString(crtChain),
		CaCertificate:     base64.StdEncoding.EncodeToString(caCRT),
		ClientCertificate: base64.StdEncoding.EncodeToString(clientCRT),
	}

	key, err := ParsePrivateKey(clientKey)
	require.NoError(t, err)

	// when
	credentials, err := NewCredentials(key, certificateResponse)
	require.NoError(t, err)

	// then
	assert.Equal(t, expectedCredentials, credentials)
}
