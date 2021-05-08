package strategy

import (
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/apperrors"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/applications"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/model"
)

const (
	OauthClientIDKey     = "clientId"
	OauthClientSecretKey = "clientSecret"
)

type oauth struct{}

func (svc *oauth) ToCredentials(secretData SecretData, appCredentials *applications.Credentials) (model.Credentials, apperrors.AppError) {
	clientId, clientSecret := svc.readOauthMap(secretData)
	return model.Credentials{
		Oauth: &model.Oauth{
			ClientID:     clientId,
			ClientSecret: clientSecret,
			URL:          appCredentials.AuthenticationUrl,
		},
		CSRFInfo:          convertToModelCSRInfo(appCredentials),
	}, nil
}

func (svc *oauth) CredentialsProvided(credentials *model.Credentials) bool {
	return svc.oauthCredentialsProvided(credentials)
}

func (svc *oauth) CreateSecretData(credentials *model.Credentials) (SecretData, apperrors.AppError) {
	return svc.makeOauthMap(credentials.Oauth.ClientID, credentials.Oauth.ClientSecret)
}

func (svc *oauth) ToCredentialsInfo(credentials *model.Credentials, secretName string) applications.Credentials {

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

func (svc *oauth) oauthCredentialsProvided(credentials *model.Credentials) bool {
	return credentials != nil && credentials.Oauth != nil && credentials.Oauth.ClientID != "" && credentials.Oauth.ClientSecret != ""
}

func (svc *oauth) makeOauthMap(clientID, clientSecret string) (map[string][]byte, apperrors.AppError) {
	m := map[string][]byte{
		OauthClientIDKey:     []byte(clientID),
		OauthClientSecretKey: []byte(clientSecret),
	}
	return m, nil
}

func (svc *oauth) readOauthMap(data map[string][]byte) (clientID, clientSecret string) {
	return string(data[OauthClientIDKey]), string(data[OauthClientSecretKey])
}
