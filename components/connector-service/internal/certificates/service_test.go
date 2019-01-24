package certificates_test

import (
	"crypto/rsa"
	"crypto/x509"
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
	crtBase64 = "crtBase64"

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
		certUtils.On("CreateCrtChain", caCrt, csr, caKey).Return(crtBase64, nil)

		certificatesService := certificates.NewCertificateService(secretsRepository, certUtils, authSecretName, subjectValues)

		// when
		crt, err := certificatesService.SignCSR(rawCSR, appName)

		// then
		require.NoError(t, err)

		assert.Equal(t, crtBase64, crt)
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
		crt, err := certificatesService.SignCSR(rawCSR, appName)

		// then
		require.Error(t, err)
		assert.Equal(t, apperrors.CodeNotFound, err.Code())
		assert.Equal(t, "", crt)
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
		crt, err := certificatesService.SignCSR(rawCSR, appName)

		// then
		require.Error(t, err)
		assert.Equal(t, "", crt)
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
		crt, err := certificatesService.SignCSR(rawCSR, appName)

		// then
		require.Error(t, err)
		assert.Equal(t, "", crt)
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
		crt, err := certificatesService.SignCSR(rawCSR, appName)

		// then
		require.Error(t, err)
		assert.Equal(t, "", crt)
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
		crt, err := certificatesService.SignCSR(rawCSR, appName)

		// then
		require.Error(t, err)
		assert.Equal(t, "", crt)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		secretsRepository.AssertExpectations(t)
		certUtils.AssertExpectations(t)
	})
}
