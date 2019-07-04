package strategy

import (
	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/applications"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/model"
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

func (svc *oauth) makeOauthMap(clientID, clientSecret string) map[string][]byte {
	return map[string][]byte{
		OauthClientIDKey:     []byte(clientID),
		OauthClientSecretKey: []byte(clientSecret),
	}
}

func (svc *oauth) readOauthMap(data map[string][]byte) (clientID, clientSecret string) {
	return string(data[OauthClientIDKey]), string(data[OauthClientSecretKey])
}
