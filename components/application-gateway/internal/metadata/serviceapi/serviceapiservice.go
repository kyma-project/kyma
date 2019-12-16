package serviceapi

import (
	"encoding/json"

	"github.com/kyma-project/kyma/components/application-gateway/pkg/authorization"

	"github.com/kyma-project/kyma/components/application-gateway/internal/metadata/applications"
	"github.com/kyma-project/kyma/components/application-gateway/internal/metadata/model"
	"github.com/kyma-project/kyma/components/application-gateway/internal/metadata/secrets"
	"github.com/kyma-project/kyma/components/application-gateway/pkg/apperrors"
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

		api.Credentials, err = sas.readCredentials(secret, applicationAPI)
		if err != nil {
			return nil, err
		}
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

func (sas defaultService) readCredentials(secret map[string][]byte, applicationAPI *applications.ServiceAPI) (*authorization.Credentials, apperrors.AppError) {
	var credentials *authorization.Credentials

	credentialsType := applicationAPI.Credentials.Type

	if credentialsType == TypeOAuth {
		oAuthCredentials, err := getOAuthCredentials(secret, applicationAPI.Credentials.Url)
		if err != nil {
			return nil, err
		}
		credentials = &authorization.Credentials{
			OAuth: oAuthCredentials,
		}
	} else if credentialsType == TypeBasic {
		credentials = &authorization.Credentials{
			BasicAuth: getBasicAuthCredentials(secret),
		}
	} else if credentialsType == TypeCertificateGen {
		credentials = &authorization.Credentials{
			CertificateGen: getCertificateGenCredentials(secret),
		}
	} else {
		credentials = nil
	}

	if credentials != nil {
		credentials.CSRFTokenEndpointURL = applicationAPI.Credentials.CSRFTokenEndpointURL
	}

	return credentials, nil
}

func getRequestParameters(secret map[string][]byte) (*authorization.RequestParameters, apperrors.AppError) {
	requestParameters := &authorization.RequestParameters{}

	headersData := secret[HeadersKey]
	if headersData != nil {
		var headers = &map[string][]string{}
		err := json.Unmarshal(headersData, headers)
		if err != nil {
			return nil, apperrors.Internal("Failed to unmarshal headers, %s", err.Error())
		}

		requestParameters.Headers = headers
	}

	queryParamsData := secret[QueryParametersKey]
	if queryParamsData != nil {
		var queryParameters = &map[string][]string{}
		err := json.Unmarshal(queryParamsData, queryParameters)
		if err != nil {
			return nil, apperrors.Internal("Failed to unmarshal query parameters, %s", err.Error())
		}

		requestParameters.QueryParameters = queryParameters
	}

	if requestParameters.Headers == nil && requestParameters.QueryParameters == nil {
		return nil, nil
	}

	return requestParameters, nil
}

func getOAuthCredentials(secret map[string][]byte, url string) (*authorization.OAuth, apperrors.AppError) {
	requestParameters, err := getRequestParameters(secret)
	if err != nil {
		return nil, err
	}

	return &authorization.OAuth{
		ClientID:          string(secret[ClientIDKey]),
		ClientSecret:      string(secret[ClientSecretKey]),
		URL:               url,
		RequestParameters: requestParameters,
	}, nil
}

func getBasicAuthCredentials(secret map[string][]byte) *authorization.BasicAuth {
	return &authorization.BasicAuth{
		Username: string(secret[UsernameKey]),
		Password: string(secret[PasswordKey]),
	}
}

func getCertificateGenCredentials(secret map[string][]byte) *authorization.CertificateGen {
	return &authorization.CertificateGen{
		CommonName:  string(secret[CommonNameKey]),
		Certificate: secret[CertificateKey],
		PrivateKey:  secret[PrivateKeyKey],
	}
}
