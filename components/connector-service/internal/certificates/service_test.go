package certificates_test

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"testing"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/certificates"

	"github.com/kyma-project/kyma/components/connector-service/internal/certificates/mocks"

	secretsMock "github.com/kyma-project/kyma/components/connector-service/internal/secrets/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	authSecretName = "nginx-auth-ca"

	appName            = "appName"
	country            = "country"
	organization       = "organization"
	organizationalUnit = "organizationalUnit"
	locality           = "locality"
	province           = "province"
)

var (
	rawCSR = []byte("csr")

	caCrtEncoded = []byte("caCrtEncoded")
	caKeyEncoded = []byte("caKeyEncoded")

	caCrt     = &x509.Certificate{}
	caKey     = &rsa.PrivateKey{}
	csr       = &x509.CertificateRequest{}
	clientCRT = []byte("clientCertificate")
	certChain = []byte("chain")

	subjectValues = certificates.CSRSubject{
		CommonName:         appName,
		Country:            country,
		Organization:       organization,
		OrganizationalUnit: organizationalUnit,
		Locality:           locality,
		Province:           province,
	}
)

func TestSignatureHandler_SignCSR(t *testing.T) {

	t.Run("should create certificate", func(t *testing.T) {
		// given
		secretsRepository := &secretsMock.Repository{}
		secretsRepository.On("Get", authSecretName).Return(caCrtEncoded, caKeyEncoded, nil)

		certUtils := &mocks.CertificateUtility{}
		certUtils.On("LoadCert", caCrtEncoded).Return(caCrt, nil)
		certUtils.On("LoadKey", caKeyEncoded).Return(caKey, nil)
		certUtils.On("LoadCSR", rawCSR).Return(csr, nil)
		certUtils.On("CheckCSRValues", csr, subjectValues).Return(nil)
		certUtils.On("SignCSR", caCrt, csr, caKey).Return(clientCRT, nil)
		certUtils.On("CreateCrtChain", caCrt.Raw, clientCRT).Return(certChain)

		certificatesService := certificates.NewCertificateService(secretsRepository, certUtils, authSecretName, subjectValues)

		// when
		encodedCertChain, apperr := certificatesService.SignCSR(rawCSR, appName)

		// then
		require.NoError(t, apperr)
		assert.NotEmpty(t, encodedCertChain)

		decodedClientCRT, err := decodeBase64(encodedCertChain.ClientCertificate)
		require.NoError(t, err)
		assert.Equal(t, clientCRT, decodedClientCRT)

		decodedChain, err := decodeBase64(encodedCertChain.CertificateChain)
		require.NoError(t, err)
		assert.Equal(t, certChain, decodedChain)

		secretsRepository.AssertExpectations(t)
		certUtils.AssertExpectations(t)
	})

	t.Run("should return Not Found error when secret not found", func(t *testing.T) {
		// given
		secretsRepository := &secretsMock.Repository{}
		secretsRepository.On("Get", authSecretName).Return([]byte(""), []byte(""), apperrors.NotFound("error"))

		certUtils := &mocks.CertificateUtility{}
		certUtils.On("LoadCSR", rawCSR).Return(csr, nil)
		certUtils.On("CheckCSRValues", csr, subjectValues).Return(nil)

		certificatesService := certificates.NewCertificateService(secretsRepository, certUtils, authSecretName, subjectValues)

		// when
		encodedChain, err := certificatesService.SignCSR(rawCSR, appName)

		// then
		require.Error(t, err)
		assert.Equal(t, apperrors.CodeNotFound, err.Code())
		assert.Empty(t, encodedChain)
		secretsRepository.AssertExpectations(t)
		certUtils.AssertExpectations(t)
	})

	t.Run("should return error when couldn't load csr", func(t *testing.T) {
		// given
		secretsRepository := &secretsMock.Repository{}

		certUtils := &mocks.CertificateUtility{}
		certUtils.On("LoadCSR", rawCSR).Return(nil, apperrors.Internal("error"))

		certificatesService := certificates.NewCertificateService(secretsRepository, certUtils, authSecretName, subjectValues)

		// when
		encodedChain, err := certificatesService.SignCSR(rawCSR, appName)

		// then
		require.Error(t, err)
		assert.Empty(t, encodedChain)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		secretsRepository.AssertExpectations(t)
		certUtils.AssertExpectations(t)
	})

	t.Run("should return error when subject check failed", func(t *testing.T) {
		// given
		secretsRepository := &secretsMock.Repository{}

		certUtils := &mocks.CertificateUtility{}
		certUtils.On("LoadCSR", rawCSR).Return(csr, nil)
		certUtils.On("CheckCSRValues", csr, subjectValues).Return(apperrors.Forbidden("error"))

		certificatesService := certificates.NewCertificateService(secretsRepository, certUtils, authSecretName, subjectValues)

		// when
		encodedChain, err := certificatesService.SignCSR(rawCSR, appName)

		// then
		require.Error(t, err)
		assert.Empty(t, encodedChain)
		assert.Equal(t, apperrors.CodeForbidden, err.Code())
		secretsRepository.AssertExpectations(t)
		certUtils.AssertExpectations(t)
	})

	t.Run("should return error when couldn't load cert", func(t *testing.T) {
		// given
		secretsRepository := &secretsMock.Repository{}
		secretsRepository.On("Get", authSecretName).Return(caCrtEncoded, caKeyEncoded, nil)

		certUtils := &mocks.CertificateUtility{}
		certUtils.On("LoadCSR", rawCSR).Return(csr, nil)
		certUtils.On("CheckCSRValues", csr, subjectValues).Return(nil)
		certUtils.On("LoadCert", caCrtEncoded).Return(nil, apperrors.Internal("error"))

		certificatesService := certificates.NewCertificateService(secretsRepository, certUtils, authSecretName, subjectValues)

		// when
		encodedChain, err := certificatesService.SignCSR(rawCSR, appName)

		// then
		require.Error(t, err)
		assert.Empty(t, encodedChain)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		secretsRepository.AssertExpectations(t)
		certUtils.AssertExpectations(t)
	})

	t.Run("should return error when couldn't load key", func(t *testing.T) {
		// given
		secretsRepository := &secretsMock.Repository{}
		secretsRepository.On("Get", authSecretName).Return(caCrtEncoded, caKeyEncoded, nil)

		certUtils := &mocks.CertificateUtility{}
		certUtils.On("LoadCSR", rawCSR).Return(csr, nil)
		certUtils.On("CheckCSRValues", csr, subjectValues).Return(nil)
		certUtils.On("LoadCert", caCrtEncoded).Return(caCrt, nil)
		certUtils.On("LoadKey", caKeyEncoded).Return(nil, apperrors.Internal("error"))

		certificatesService := certificates.NewCertificateService(secretsRepository, certUtils, authSecretName, subjectValues)

		// when
		encodedChain, err := certificatesService.SignCSR(rawCSR, appName)

		// then
		require.Error(t, err)
		assert.Empty(t, encodedChain)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		secretsRepository.AssertExpectations(t)
		certUtils.AssertExpectations(t)
	})

	t.Run("should return error when failed to sign CSR", func(t *testing.T) {
		// given
		secretsRepository := &secretsMock.Repository{}
		secretsRepository.On("Get", authSecretName).Return(caCrtEncoded, caKeyEncoded, nil)

		certUtils := &mocks.CertificateUtility{}
		certUtils.On("LoadCert", caCrtEncoded).Return(caCrt, nil)
		certUtils.On("LoadKey", caKeyEncoded).Return(caKey, nil)
		certUtils.On("LoadCSR", rawCSR).Return(csr, nil)
		certUtils.On("CheckCSRValues", csr, subjectValues).Return(nil)
		certUtils.On("SignCSR", caCrt, csr, caKey).Return(nil, apperrors.Internal("error"))

		certificatesService := certificates.NewCertificateService(secretsRepository, certUtils, authSecretName, subjectValues)

		// when
		encodedChain, err := certificatesService.SignCSR(rawCSR, appName)

		// then
		require.Error(t, err)
		assert.Empty(t, encodedChain)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		secretsRepository.AssertExpectations(t)
		certUtils.AssertExpectations(t)
	})
}

func decodeBase64(base64CrtChain string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(base64CrtChain)
}
