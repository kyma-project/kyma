package certificates

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCsrProvider_CreateCSR(t *testing.T) {

	subject := pkix.Name{
		OrganizationalUnit: []string{"OrgUnit"},
		Organization:       []string{"Organization"},
		Locality:           []string{"Waldorf"},
		Province:           []string{"Waldorf"},
		Country:            []string{"DE"},
		CommonName:         "test-app",
	}

	t.Run("should create CSR with new key", func(t *testing.T) {
		// given
		csrProvider := NewCSRProvider()

		// when
		csr, key, err := csrProvider.CreateCSR(subject)

		// then
		require.NoError(t, err)
		require.NotEmpty(t, csr)
		require.NotEmpty(t, key)

		receivedCSR := decodeCSR(t, csr)

		require.NotNil(t, receivedCSR)
		assertSubject(t, receivedCSR)
	})

}

func assertSubject(t *testing.T, csr *x509.CertificateRequest) {
	assert.Equal(t, "test-app", csr.Subject.CommonName)
	assert.Equal(t, "OrgUnit", csr.Subject.OrganizationalUnit[0])
	assert.Equal(t, "Organization", csr.Subject.Organization[0])
	assert.Equal(t, "Waldorf", csr.Subject.Locality[0])
	assert.Equal(t, "Waldorf", csr.Subject.Province[0])
	assert.Equal(t, "DE", csr.Subject.Country[0])
}

func decodeCSR(t *testing.T, encodedCSR string) *x509.CertificateRequest {
	csrBytes, err := base64.StdEncoding.DecodeString(encodedCSR)
	require.NoError(t, err)

	pemCSR, _ := pem.Decode(csrBytes)
	require.NotNil(t, pemCSR)

	receivedCSR, err := x509.ParseCertificateRequest(pemCSR.Bytes)
	require.NoError(t, err)

	return receivedCSR
}
