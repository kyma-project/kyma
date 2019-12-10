package strategy

import (
	"encoding/json"

	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/applications"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/model"
)

const (
	OauthClientIDKey     = "clientId"
	OauthClientSecretKey = "clientSecret"
	HeadersKey           = "headers"
	QueryParametersKey   = "queryParameters"
)

type oauth struct{}

func (svc *oauth) ToCredentials(secretData SecretData, appCredentials *applications.Credentials) (model.CredentialsWithCSRF, apperrors.AppError) {
	clientId, clientSecret, requestParameters, err := svc.readOauthMap(secretData)
	if err != nil {
		return model.CredentialsWithCSRF{}, apperrors.Internal("Failed to read OAuth map, %v", err)
	}

	return model.CredentialsWithCSRF{
		Oauth: &model.Oauth{
			ClientID:          clientId,
			ClientSecret:      clientSecret,
			URL:               appCredentials.AuthenticationUrl,
			RequestParameters: requestParameters,
		}, CSRFInfo: convertToModelCSRInfo(appCredentials),
	}, nil
}

func (svc *oauth) CredentialsProvided(credentials *model.CredentialsWithCSRF) bool {
	return svc.oauthCredentialsProvided(credentials)
}

func (svc *oauth) CreateSecretData(credentials *model.CredentialsWithCSRF) (SecretData, apperrors.AppError) {
	return svc.makeOauthMap(credentials.Oauth.ClientID, credentials.Oauth.ClientSecret, credentials.Oauth.RequestParameters)
}

func (svc *oauth) ToCredentialsInfo(credentials *model.CredentialsWithCSRF, secretName string) applications.Credentials {

	applicationCredentials := applications.Credentials{
		AuthenticationUrl: credentials.Oauth.URL,
		Type:              applications.CredentialsOAuthType,
		SecretName:        secretName,
		CSRFInfo:          toAppCSRFInfo(credentials),
	}

	return applicationCredentials
}

func (svc *oauth) ShouldUpdate(currentData SecretData, newData SecretData) bool {
	return string(currentData[OauthClientIDKey]) != string(newData[OauthClientIDKey]) ||
		string(currentData[OauthClientSecretKey]) != string(newData[OauthClientSecretKey])
}

func (svc *oauth) oauthCredentialsProvided(credentials *model.CredentialsWithCSRF) bool {
	return credentials != nil && credentials.Oauth != nil && credentials.Oauth.ClientID != "" && credentials.Oauth.ClientSecret != ""
}

func (svc *oauth) makeOauthMap(clientID, clientSecret string, requestParameters *model.RequestParameters) (map[string][]byte, apperrors.AppError) {
	m := map[string][]byte{
		OauthClientIDKey:     []byte(clientID),
		OauthClientSecretKey: []byte(clientSecret),
	}
	if requestParameters != nil && requestParameters.Headers != nil {
		headers, err := json.Marshal(requestParameters.Headers)
		if err != nil {
			return map[string][]byte{}, apperrors.Internal("Failed to marshall headers from request parameters: %v", err)
		}
		m[HeadersKey] = headers
	}
	if requestParameters != nil && requestParameters.QueryParameters != nil {
		queryParameters, err := json.Marshal(requestParameters.QueryParameters)
		if err != nil {
			return map[string][]byte{}, apperrors.Internal("Failed to marshall query parameters from request parameters: %v", err)
		}
		m[QueryParametersKey] = queryParameters
	}
	return m, nil
}

func (svc *oauth) readOauthMap(data map[string][]byte) (clientID, clientSecret string, requestParameters *model.RequestParameters, err error) {
	requestParameters, err = getRequestParameters(data)
	if err != nil {
		return "", "", nil, err
	}
	return string(data[OauthClientIDKey]), string(data[OauthClientSecretKey]), requestParameters, nil
}

func getRequestParameters(secret map[string][]byte) (*model.RequestParameters, error) {
	requestParameters := &model.RequestParameters{}

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
