package serviceapi

import (
	"encoding/json"

	"github.com/kyma-project/kyma/components/application-gateway/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-gateway/internal/metadata/applications"
	"github.com/kyma-project/kyma/components/application-gateway/internal/metadata/model"
	"github.com/kyma-project/kyma/components/application-gateway/internal/metadata/secrets"
)

const (
	ClientIDKey        = "clientId"
	ClientSecretKey    = "clientSecret"
	UsernameKey        = "username"
	PasswordKey        = "password"
	TypeOAuth          = "OAuth"
	TypeBasic          = "Basic"
	TypeCertificateGen = "CertificateGen"
	PrivateKeyKey      = "key"
	CertificateKey     = "crt"
	CommonNameKey      = "commonName"

	HeadersKey         = "headers"
	QueryParametersKey = "queryParameters"
)

// Service manages API definition of a service
type Service interface {
	// Read reads API from Application API definition. It also reads all additional information.
	Read(*applications.ServiceAPI) (*model.API, apperrors.AppError)
}

type defaultService struct {
	secretsRepository secrets.Repository
}

func NewService(secretsRepository secrets.Repository) Service {

	return defaultService{
		secretsRepository: secretsRepository,
	}
}

func (sas defaultService) Read(applicationAPI *applications.ServiceAPI) (*model.API, apperrors.AppError) {
	api := &model.API{
		TargetUrl: applicationAPI.TargetUrl,
	}

	if applicationAPI.Credentials != nil {
		credentialsSecretName := applicationAPI.Credentials.SecretName

		secret, err := sas.secretsRepository.Get(credentialsSecretName)
		if err != nil {
			return nil, err
		}

		api.Credentials = sas.readCredentials(secret, applicationAPI)
	}

	if applicationAPI.RequestParametersSecretName != "" {
		secret, err := sas.secretsRepository.Get(applicationAPI.RequestParametersSecretName)
		if err != nil {
			return nil, err
		}

		requestParameters, err := getRequestParameters(secret)
		if err != nil {
			return nil, err
		}

		api.RequestParameters = requestParameters
	}

	return api, nil
}

func (sas defaultService) readCredentials(secret map[string][]byte, applicationAPI *applications.ServiceAPI) *model.Credentials {
	var credentials *model.Credentials

	credentialsType := applicationAPI.Credentials.Type

	if credentialsType == TypeOAuth {
		credentials = &model.Credentials{
			OAuth: getOAuthCredentials(secret, applicationAPI.Credentials.Url),
		}
	} else if credentialsType == TypeBasic {
		credentials = &model.Credentials{
			BasicAuth: getBasicAuthCredentials(secret),
		}
	} else if credentialsType == TypeCertificateGen {
		credentials = &model.Credentials{
			CertificateGen: getCertificateGenCredentials(secret),
		}
	} else {
		credentials = nil
	}

	if credentials != nil {
		credentials.CSRFTokenEndpointURL = applicationAPI.Credentials.CSRFTokenEndpointURL
	}

	return credentials
}

func getRequestParameters(secret map[string][]byte) (*model.RequestParameters, apperrors.AppError) {
	headers := &map[string][]string{}
	err := json.Unmarshal(secret[HeadersKey], headers)
	if err != nil {
		return nil, apperrors.Internal("Failed to unmarshal headers, %s", err.Error())
	}

	queryParameters := &map[string][]string{}
	err = json.Unmarshal(secret[QueryParametersKey], queryParameters)
	if err != nil {
		return nil, apperrors.Internal("Failed to unmarshal query parameters, %s", err.Error())
	}

	return &model.RequestParameters{
		Headers:         headers,
		QueryParameters: queryParameters,
	}, nil
}

func getOAuthCredentials(secret map[string][]byte, url string) *model.OAuth {
	return &model.OAuth{
		ClientID:     string(secret[ClientIDKey]),
		ClientSecret: string(secret[ClientSecretKey]),
		URL:          url,
	}
}

func getBasicAuthCredentials(secret map[string][]byte) *model.BasicAuth {
	return &model.BasicAuth{
		Username: string(secret[UsernameKey]),
		Password: string(secret[PasswordKey]),
	}
}

func getCertificateGenCredentials(secret map[string][]byte) *model.CertificateGen {
	return &model.CertificateGen{
		CommonName:  string(secret[CommonNameKey]),
		Certificate: secret[CertificateKey],
		PrivateKey:  secret[PrivateKeyKey],
	}
}
