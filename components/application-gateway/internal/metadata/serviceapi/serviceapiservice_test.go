package serviceapi

import (
	"testing"

	"github.com/kyma-project/kyma/components/application-gateway/internal/metadata/model"

	secretsmocks "github.com/kyma-project/kyma/components/application-gateway/internal/metadata/secrets/mocks"

	"github.com/kyma-project/kyma/components/application-gateway/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-gateway/internal/metadata/applications"
	"github.com/stretchr/testify/assert"
)

const (
	targetUrl    = "http://target.com"
	clientId     = "clientId"
	clientSecret = "clientSecret"
	oauthUrl     = "http://oauth.com"
	secretName   = "secret-name"
	username     = "username"
	password     = "password"
	commonName   = "commonName"
)

var (
	certificate = []byte("certificate")
	privateKey  = []byte("privateKey")
)

func TestDefaultService_Read(t *testing.T) {
	testCases := []struct {
		applicationAPI *applications.ServiceAPI
		secret         map[string][]byte
		resultingAPI   *model.API
	}{
		{
			applicationAPI: &applications.ServiceAPI{
				TargetUrl: targetUrl,
				Credentials: &applications.Credentials{
					Type:       TypeOAuth,
					SecretName: secretName,
					Url:        oauthUrl,
				},
			},
			secret: map[string][]byte{
				ClientIDKey:     []byte(clientId),
				ClientSecretKey: []byte(clientSecret),
			},
			resultingAPI: &model.API{
				TargetUrl: targetUrl,
				Credentials: &model.Credentials{
					OAuth: &model.OAuth{
						ClientID:     clientId,
						ClientSecret: clientSecret,
						URL:          oauthUrl,
					},
				},
			},
		},
		{
			applicationAPI: &applications.ServiceAPI{
				TargetUrl: targetUrl,
				Credentials: &applications.Credentials{
					Type:       TypeBasic,
					SecretName: secretName,
					Url:        "",
				},
			},
			secret: map[string][]byte{
				UsernameKey: []byte(username),
				PasswordKey: []byte(password),
			},
			resultingAPI: &model.API{
				TargetUrl: targetUrl,
				Credentials: &model.Credentials{
					BasicAuth: &model.BasicAuth{
						Username: username,
						Password: password,
					},
				},
			},
		},
		{
			applicationAPI: &applications.ServiceAPI{
				TargetUrl: targetUrl,
				Credentials: &applications.Credentials{
					Type:       TypeCertificateGen,
					SecretName: secretName,
					Url:        "",
				},
			},
			secret: map[string][]byte{
				CommonNameKey:  []byte(commonName),
				CertificateKey: certificate,
				PrivateKeyKey:  privateKey,
			},
			resultingAPI: &model.API{
				TargetUrl: targetUrl,
				Credentials: &model.Credentials{
					CertificateGen: &model.CertificateGen{
						CommonName:  commonName,
						Certificate: certificate,
						PrivateKey:  privateKey,
					},
				},
			},
		},
		{
			applicationAPI: &applications.ServiceAPI{
				TargetUrl: targetUrl,
			},
			secret: map[string][]byte{},
			resultingAPI: &model.API{
				TargetUrl: targetUrl,
			},
		},
	}

	t.Run("should read API", func(t *testing.T) {
		// given
		for _, test := range testCases {
			secretsRepository := new(secretsmocks.Repository)
			secretsRepository.On("Get", secretName).Return(test.secret, nil)
			service := NewService(secretsRepository)

			// when
			api, err := service.Read(test.applicationAPI)

			// then
			assert.NoError(t, err)
			assert.Equal(t, test.resultingAPI, api)
		}
	})

	t.Run("should return error when reading secret fails", func(t *testing.T) {
		// given
		applicationServiceAPI := &applications.ServiceAPI{
			TargetUrl: "http://target.com",
			Credentials: &applications.Credentials{
				Type:       "OAuth",
				SecretName: "secret-name",
				Url:        "http://oauth.com",
			},
		}

		secretsRepository := new(secretsmocks.Repository)
		secretsRepository.On("Get", "secret-name").
			Return(nil, apperrors.Internal("secret error"))

		service := NewService(secretsRepository)

		// when
		api, err := service.Read(applicationServiceAPI)

		// then
		assert.Error(t, err)
		assert.Nil(t, api)
		assert.Contains(t, err.Error(), "secret error")

		secretsRepository.AssertExpectations(t)
	})
}
