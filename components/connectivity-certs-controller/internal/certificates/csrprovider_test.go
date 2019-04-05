package certificates

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"testing"

	"github.com/kyma-project/kyma/components/connectivity-certs-controller/internal/secrets/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	clusterSecretName = "cluster-cert-secret"
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
		secretRepository := &mocks.Repository{}
		secretRepository.On("UpsertWithReplace", clusterSecretName, mock.AnythingOfType("map[string][]uint8")).Return(nil).
			Run(func(args mock.Arguments) {
				secretData, ok := args[1].(map[string][]byte)
				require.True(t, ok)
				assert.NotEmpty(t, secretData[clusterKeySecretKey])
			})

		csrProvider := NewCSRProvider(clusterSecretName, "", secretRepository)

		// when
		csr, err := csrProvider.CreateCSR(subject)

		// then
		require.NoError(t, err)
		require.NotEmpty(t, csr)

		receivedCSR := decodeCSR(t, csr)

		require.NotNil(t, receivedCSR)
		assertSubject(t, receivedCSR)
		secretRepository.AssertExpectations(t)
	})

	t.Run("should return error when failed to override secret", func(t *testing.T) {
		// given
		secretRepository := &mocks.Repository{}
		secretRepository.On("UpsertWithReplace", clusterSecretName, mock.AnythingOfType("map[string][]uint8")).Return(errors.New("error"))

		csrProvider := NewCSRProvider(clusterSecretName, "", secretRepository)

		// when
		_, err := csrProvider.CreateCSR(subject)

		// then
		require.Error(t, err)
		secretRepository.AssertExpectations(t)
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
