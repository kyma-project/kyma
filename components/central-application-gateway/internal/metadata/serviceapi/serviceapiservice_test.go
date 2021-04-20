package serviceapi

import (
	"testing"

	"github.com/kyma-project/kyma/components/application-gateway/pkg/authorization"

	"github.com/kyma-project/kyma/components/application-gateway/internal/metadata/model"

	secretsmocks "github.com/kyma-project/kyma/components/application-gateway/internal/metadata/secrets/mocks"

	"github.com/kyma-project/kyma/components/application-gateway/internal/metadata/applications"
	"github.com/kyma-project/kyma/components/application-gateway/pkg/apperrors"
	"github.com/stretchr/testify/assert"
)

const (
	targetUrl    = "http://target.com"
	clientId     = "clientId"
	clientSecret = "clientSecret"
	oauthUrl     = "http://oauth.com"
	secretName   = "credentialsSecret-name"
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
		description             string
		applicationAPI          *applications.ServiceAPI
		credentialsSecret       map[string][]byte
		requestParamsSecretName string
		requestParamsSecret     map[string][]byte
		resultingAPI            *model.API
	}{
		{
			description: "api with oauth credentials",
			applicationAPI: &applications.ServiceAPI{
				TargetURL: targetUrl,
				Credentials: &applications.Credentials{
					Type:       TypeOAuth,
					SecretName: secretName,
					URL:        oauthUrl,
				},
			},
			credentialsSecret: map[string][]byte{
				ClientIDKey:     []byte(clientId),
				ClientSecretKey: []byte(clientSecret),
			},
			resultingAPI: &model.API{
				TargetUrl: targetUrl,
				Credentials: &authorization.Credentials{
					OAuth: &authorization.OAuth{
						ClientID:     clientId,
						ClientSecret: clientSecret,
						URL:          oauthUrl,
					},
				},
			},
		},
		{
			description: "api with basic auth credentials",
			applicationAPI: &applications.ServiceAPI{
				TargetURL: targetUrl,
				Credentials: &applications.Credentials{
					Type:       TypeBasic,
					SecretName: secretName,
					URL:        "",
				},
			},
			credentialsSecret: map[string][]byte{
				UsernameKey: []byte(username),
				PasswordKey: []byte(password),
			},
			resultingAPI: &model.API{
				TargetUrl: targetUrl,
				Credentials: &authorization.Credentials{
					BasicAuth: &authorization.BasicAuth{
						Username: username,
						Password: password,
					},
				},
			},
		},
		{
			description: "api with certificate gen credentials",
			applicationAPI: &applications.ServiceAPI{
				TargetURL: targetUrl,
				Credentials: &applications.Credentials{
					Type:       TypeCertificateGen,
					SecretName: secretName,
					URL:        "",
				},
			},
			credentialsSecret: map[string][]byte{
				CommonNameKey:  []byte(commonName),
				CertificateKey: certificate,
				PrivateKeyKey:  privateKey,
			},
			resultingAPI: &model.API{
				TargetUrl: targetUrl,
				Credentials: &authorization.Credentials{
					CertificateGen: &authorization.CertificateGen{
						CommonName:  commonName,
						Certificate: certificate,
						PrivateKey:  privateKey,
					},
				},
			},
		},
		{
			description: "api without credentials",
			applicationAPI: &applications.ServiceAPI{
				TargetURL: targetUrl,
			},
			credentialsSecret: map[string][]byte{},
			resultingAPI: &model.API{
				TargetUrl: targetUrl,
			},
		},
		{
			description: "api with headers and query parameters",
			applicationAPI: &applications.ServiceAPI{
				TargetURL:                   targetUrl,
				RequestParametersSecretName: "params-secret",
			},
			credentialsSecret:       map[string][]byte{},
			requestParamsSecretName: "params-secret",
			requestParamsSecret: map[string][]byte{
				HeadersKey:         []byte(`{"header":["headerValue"]}`),
				QueryParametersKey: []byte(`{"query":["queryValue"]}`),
			},
			resultingAPI: &model.API{
				TargetUrl: targetUrl,
				RequestParameters: &authorization.RequestParameters{
					Headers: &map[string][]string{
						"header": {"headerValue"},
					},
					QueryParameters: &map[string][]string{
						"query": {"queryValue"},
					},
				},
			},
		},
		{
			description: "api with query parameters only",
			applicationAPI: &applications.ServiceAPI{
				TargetURL:                   targetUrl,
				RequestParametersSecretName: "params-secret",
			},
			credentialsSecret:       map[string][]byte{},
			requestParamsSecretName: "params-secret",
			requestParamsSecret: map[string][]byte{
				QueryParametersKey: []byte(`{"query":["queryValue"]}`),
			},
			resultingAPI: &model.API{
				TargetUrl: targetUrl,
				RequestParameters: &authorization.RequestParameters{
					QueryParameters: &map[string][]string{
						"query": {"queryValue"},
					},
				},
			},
		},
		{
			description: "api with headers only",
			applicationAPI: &applications.ServiceAPI{
				TargetURL:                   targetUrl,
				RequestParametersSecretName: "params-secret",
			},
			credentialsSecret:       map[string][]byte{},
			requestParamsSecretName: "params-secret",
			requestParamsSecret: map[string][]byte{
				HeadersKey: []byte(`{"header":["headerValue"]}`),
			},
			resultingAPI: &model.API{
				TargetUrl: targetUrl,
				RequestParameters: &authorization.RequestParameters{
					Headers: &map[string][]string{
						"header": {"headerValue"},
					},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run("should read "+test.description, func(t *testing.T) {
			// given
			secretsRepository := new(secretsmocks.Repository)
			secretsRepository.On("Get", secretName).Return(test.credentialsSecret, nil)
			if test.requestParamsSecretName != "" {
				secretsRepository.On("Get", test.requestParamsSecretName).Return(test.requestParamsSecret, nil)
			}

			service := NewService(secretsRepository)

			// when
			api, err := service.Read(test.applicationAPI)

			// then
			assert.NoError(t, err)
			assert.Equal(t, test.resultingAPI, api)
		})
	}

	t.Run("should return error when reading credentialsSecret fails", func(t *testing.T) {
		// given
		applicationServiceAPI := &applications.ServiceAPI{
			TargetURL: "http://target.com",
			Credentials: &applications.Credentials{
				Type:       "OAuth",
				SecretName: "credentialsSecret-name",
				URL:        "http://oauth.com",
			},
		}

		secretsRepository := new(secretsmocks.Repository)
		secretsRepository.On("Get", "credentialsSecret-name").
			Return(nil, apperrors.Internal("credentialsSecret error"))

		service := NewService(secretsRepository)

		// when
		api, err := service.Read(applicationServiceAPI)

		// then
		assert.Error(t, err)
		assert.Nil(t, api)
		assert.Contains(t, err.Error(), "credentialsSecret error")

		secretsRepository.AssertExpectations(t)
	})

	t.Run("should return error when reading request parameters fails", func(t *testing.T) {
		// given
		applicationServiceAPI := &applications.ServiceAPI{
			TargetURL:                   "http://target.com",
			RequestParametersSecretName: secretName,
		}

		secretsRepository := new(secretsmocks.Repository)
		secretsRepository.On("Get", secretName).
			Return(nil, apperrors.Internal("request params error"))

		service := NewService(secretsRepository)

		// when
		api, err := service.Read(applicationServiceAPI)

		// then
		assert.Error(t, err)
		assert.Nil(t, api)
		assert.Contains(t, err.Error(), "request params error")

		secretsRepository.AssertExpectations(t)
	})
}
