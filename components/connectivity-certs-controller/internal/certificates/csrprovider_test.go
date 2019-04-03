package certificates

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"testing"

	"github.com/kyma-project/kyma/components/connectivity-certs-controller/internal/secrets/mocks"

	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
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
		CommonName:         "ec-default",
	}

	t.Run("should create CSR when key exists in secret", func(t *testing.T) {
		// given
		privateKey := `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEA3LSbpAAAd0rHWOIXfqWZQxyjtfvWeHIZi33gB2yYgaSdFyh4
WJwj5U7+uDNdmPbsyAEJI5Li6kSI+V2A2euVvQzkm07DrdGQFL0Znp9/p2HdSnaD
gKCuXJYSbSubIX02HnsnvJZRRCtnt1EUQekp8t6N5T6EnmmY65woP4SxTHxLNJ+1
xKySs3fWcl167+ZKrBzXkRj9LGs6yuP9AHIP5gGj3a60/ZoGsUfiEsj/pBN7y9KN
WwqZHOPiuHzg0OLXI0Y6A3Of8y7+WrEAyS0ck7Ue5TTxRKKmJJfF7DE7GkTw8I0H
dMad3pJe/aPuDIPRiWxCVDiL9QGcejstBvTz4wIDAQABAoIBAD9lXaOxIHEjuLlO
UGNfm/OMIXZfvY5hb/cClDxttCzhJQKG7HK/fwwaMc6laohKvV8B9ScTxTx3rUS7
2AxAwIVKU8xMxqaCILnkS5ylwhxJXzBJdKKZBRyxOt/C+8+V0NrWk2Z3YyaKtUMR
9hisqhEKXoXv/FYojPV4qJL+QZNvQ+9kI7kBinred5tz5xKk2BHzYs+Em37A01NP
Mn9oftth4H1OtGQkE2mBKjt8vjP2IKVnoG/yGKbTU6C/wkXNnB0cABSdyNquJlbp
T9CVpx2qP173h0yzEAPHSkQe9DPP6JHQSgJZpGUEdoiQe5vKxmGqvRfPbq1UTo/j
jHs/etkCgYEA8zaDNk+vVAXYTjexJy5yT9cIFN8FrLloD0PkLGpEJlBMGhTLDATS
XlJrt6EFqq2abBPaiLkynF2+qvHr2WQlFA0NJHaYlcTZHE10WtFlwwlM9gEm3YeK
jYLwQmfigaB3Ut4ckr3NnZa26LSODb1rpSuWqXZMSfDN2TiaEMtcJ4UCgYEA6E8o
8/UFB8mOXSKfkHC3n/QUmcAhjVqgCF9L/6ht4NQouRRObe1kKQylh2JejetNvZLk
99dnFK6OErevHAv1j0PcBilTq2dXmKFVhkUp9bFj3Kp3ALzinZAwbV9GTWiA+3Wv
rgaLdxJxTR8Cly8JWR1YuKrWj5oCIMmFoqk7ZkcCgYEAkyTt6ZP4PVtz7I6hLVVa
b5dnGkl8A24A2Qt4Jq78IDoAcN8XoWPhapNu/B/9b6+sd6rjUkjJp/THgGDxEgsW
q7ThuKfP1PzNZeQueyuo54DfAQ7dVrXES61mcqarUUWmK4qZuuX+WlNuwgdK0mFB
mSJv+oLJ0QpRYBRwkayXSokCgYAPq89ibaPyO4mMBNrovoHUm32MRaa9x2BGUE9r
JqyK3yUEHzePONVp432DHYKtZjMvV6p0gaZlgcT5xEReyvu8t2IvVDhdtrH1DOUd
Eqta9KV87E7s0NEkueZaanPuot8Yl37LaYuc87SK9E2Tb0vdJBqpEnU46LW+Cnom
V+423wKBgBYFjwG387TT/DudzxbSvrWLGfEddbZM5sH4YcvmqBqu/qSssqd4ueOs
tOetczdVd2on5cQJ9WTvye2s066GRSIcPEnh+DQZTea350HkzHSb3+naylQ84BL3
AZwS8JvJu+rnTnjRcfcvrPYjIigwMzgB1vun3zJz1h4J9s9uCWjZ
-----END RSA PRIVATE KEY-----`

		secretData := map[string][]byte{
			clusterKeySecretKey: []byte(privateKey),
		}

		secretRepository := &mocks.Repository{}
		secretRepository.On("Get", clusterSecretName).Return(secretData, nil)

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

	t.Run("should create CSR when no secret", func(t *testing.T) {
		// given
		secretRepository := &mocks.Repository{}
		secretRepository.On("Get", clusterSecretName).Return(nil, k8serrors.NewNotFound(schema.GroupResource{}, "error"))
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

	t.Run("should create CSR when no key in secret", func(t *testing.T) {
		// given
		secretRepository := &mocks.Repository{}
		secretRepository.On("Get", clusterSecretName).Return(map[string][]byte{}, nil)
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

	t.Run("should create CSR when invalid key in secret", func(t *testing.T) {
		// given
		secretWithInvalidKey := map[string][]byte{
			clusterKeySecretKey: []byte("invalid key"),
		}

		secretRepository := &mocks.Repository{}
		secretRepository.On("Get", clusterSecretName).Return(secretWithInvalidKey, nil)
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
		secretRepository.On("Get", clusterSecretName).Return(map[string][]byte{}, nil)
		secretRepository.On("UpsertWithReplace", clusterSecretName, mock.AnythingOfType("map[string][]uint8")).Return(errors.New("error"))

		csrProvider := NewCSRProvider(clusterSecretName, "", secretRepository)

		// when
		_, err := csrProvider.CreateCSR(subject)

		// then
		require.Error(t, err)
		secretRepository.AssertExpectations(t)
	})

	t.Run("should return error when failed to get secret", func(t *testing.T) {
		// given
		secretRepository := &mocks.Repository{}
		secretRepository.On("Get", clusterSecretName).Return(nil, errors.New("error"))

		csrProvider := NewCSRProvider(clusterSecretName, "", secretRepository)

		// when
		_, err := csrProvider.CreateCSR(subject)

		// then
		require.Error(t, err)
		secretRepository.AssertExpectations(t)
	})

}

func assertSubject(t *testing.T, csr *x509.CertificateRequest) {
	assert.Equal(t, "ec-default", csr.Subject.CommonName)
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
