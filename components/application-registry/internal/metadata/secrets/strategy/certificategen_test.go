package strategy

import (
	"crypto/x509/pkix"
	"testing"

	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/certificates"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/applications"

	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/model"

	"github.com/stretchr/testify/assert"
)

const (
	commonName  = "commonName"
	certificate = "cert"
	privateKey  = "key"
)

var (
	certGenCredentials = &model.Credentials{
		CertificateGen: &model.CertificateGen{
			CommonName: commonName,
		},
	}
)

func TestCertificateGen_ToCredentials(t *testing.T) {
	secretData := map[string][]byte{
		CertificateGenPrivateKeyKey: []byte(privateKey),
		CertificateGenCertKey:       []byte(certificate),
		CertificateGenCNKey:         []byte(commonName),
	}

	t.Run("should convert to credentials", func(t *testing.T) {
		// given
		certificateGenStrategy := certificateGen{}

		// when
		credentials := certificateGenStrategy.ToCredentials(secretData, nil)

		// then
		assert.Equal(t, commonName, credentials.CertificateGen.CommonName)
	})
}

func TestCertificateGen_CredentialsProvided(t *testing.T) {
	testCases := []struct {
		credentials *model.Credentials
		result      bool
	}{
		{
			credentials: &model.Credentials{
				CertificateGen: &model.CertificateGen{CommonName: commonName},
			},
			result: true,
		},
		{
			credentials: &model.Credentials{
				CertificateGen: &model.CertificateGen{CommonName: ""},
			},
			result: false,
		},
		{
			credentials: nil,
			result:      false,
		},
	}

	t.Run("should check if credentials provided", func(t *testing.T) {
		// given
		certificateGenStrategy := certificateGen{}

		for _, test := range testCases {
			// when
			result := certificateGenStrategy.CredentialsProvided(test.credentials)

			// then
			assert.Equal(t, test.result, result)
		}
	})
}

func TestCertificateGen_CreateSecretData(t *testing.T) {
	t.Run("should create secret data", func(t *testing.T) {
		// given
		certGenerator := func(subject pkix.Name) (*certificates.KeyCertPair, apperrors.AppError) {
			return &certificates.KeyCertPair{
				PrivateKey:  []byte(privateKey),
				Certificate: []byte(certificate),
			}, nil
		}

		certificateGenStrategy := certificateGen{
			certificateGenerator: certGenerator,
		}

		// when
		secretData, err := certificateGenStrategy.CreateSecretData(certGenCredentials)

		//then
		require.NoError(t, err)
		assert.Equal(t, []byte(privateKey), secretData[CertificateGenPrivateKeyKey])
		assert.Equal(t, []byte(certificate), secretData[CertificateGenCertKey])
		assert.Equal(t, []byte(commonName), secretData[CertificateGenCNKey])
	})
}

func TestCertificateGen_ToCredentialsInfo(t *testing.T) {
	t.Run("should convert to app credentials", func(t *testing.T) {
		// given
		certificateGenStrategy := certificateGen{}

		// when
		appCredentials := certificateGenStrategy.ToCredentialsInfo(certGenCredentials, secretName)

		// then
		assert.Equal(t, applications.CredentialsCertificateGenType, appCredentials.Type)
		assert.Equal(t, secretName, appCredentials.SecretName)
		assert.Equal(t, "", appCredentials.AuthenticationUrl)
	})
}

func TestCertificateGen_ShouldUpdate(t *testing.T) {
	testCases := []struct {
		currentData SecretData
		newData     SecretData
		result      bool
	}{
		{
			currentData: SecretData{
				CertificateGenCNKey: []byte(commonName),
			},
			newData: SecretData{
				CertificateGenCNKey: []byte("new common name"),
			},
			result: true,
		},
		{
			currentData: SecretData{
				CertificateGenCNKey: []byte(commonName),
			},
			newData: SecretData{
				CertificateGenCNKey: []byte(commonName),
			},
			result: false,
		},
		{
			currentData: SecretData{},
			newData: SecretData{
				CertificateGenCNKey: []byte(commonName),
			},
			result: true,
		},
	}

	t.Run("should return true when update needed", func(t *testing.T) {
		// given
		certGenStrategy := certificateGen{}

		for _, test := range testCases {
			// when
			result := certGenStrategy.ShouldUpdate(test.currentData, test.newData)

			// then
			assert.Equal(t, test.result, result)
		}
	})
}
