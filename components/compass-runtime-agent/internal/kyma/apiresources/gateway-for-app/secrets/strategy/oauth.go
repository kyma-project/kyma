package strategy

import (
	"kyma-project.io/compass-runtime-agent/internal/apperrors"
	"kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/gateway-for-app/secrets/model"
	"kyma-project.io/compass-runtime-agent/internal/kyma/applications"
	"kyma-project.io/compass-runtime-agent/internal/kyma/applications/converters"
)

const (
	OauthClientIDKey     = "clientId"
	OauthClientSecretKey = "clientSecret"
)

type oauth struct{}

func (svc *oauth) ToCredentials(secretData SecretData, appCredentials *applications.Credentials) model.CredentialsWithCSRF {
	clientId, clientSecret := svc.readOauthMap(secretData)

	return model.CredentialsWithCSRF{
		Oauth: &model.Oauth{
			ClientID:     clientId,
			ClientSecret: clientSecret,
			URL:          appCredentials.AuthenticationUrl,
		}, CSRFInfo: convertToModelCSRInfo(appCredentials),
	}
}

func (svc *oauth) CredentialsProvided(credentials *model.CredentialsWithCSRF) bool {
	return svc.oauthCredentialsProvided(credentials)
}

func (svc *oauth) CreateSecretData(credentials *model.CredentialsWithCSRF) (SecretData, apperrors.AppError) {
	return svc.makeOauthMap(credentials.Oauth.ClientID, credentials.Oauth.ClientSecret), nil
}

func (svc *oauth) ToCredentialsInfo(credentials *model.CredentialsWithCSRF, secretName string) applications.Credentials {

	applicationCredentials := applications.Credentials{
		AuthenticationUrl: credentials.Oauth.URL,
		Type:              converters.CredentialsOAuthType,
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
	return credentials != nil && credentials.Oauth != nil
}

func (svc *oauth) makeOauthMap(clientID, clientSecret string) map[string][]byte {
	return map[string][]byte{
		OauthClientIDKey:     []byte(clientID),
		OauthClientSecretKey: []byte(clientSecret),
	}
}

func (svc *oauth) readOauthMap(data map[string][]byte) (clientID, clientSecret string) {
	return string(data[OauthClientIDKey]), string(data[OauthClientSecretKey])
}
