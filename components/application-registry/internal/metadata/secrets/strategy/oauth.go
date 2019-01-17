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

func (svc *oauth) ToCredentials(secretData map[string][]byte, appCredentials *applications.Credentials) model.Credentials {
	clientId, clientSecret := svc.readOauthMap(secretData)

	return model.Credentials{
		Oauth: &model.Oauth{
			ClientID:     clientId,
			ClientSecret: clientSecret,
			URL:          appCredentials.AuthenticationUrl,
		},
	}
}

func (svc *oauth) CredentialsProvided(credentials *model.Credentials) bool {
	return svc.oauthCredentialsProvided(credentials)
}

func (svc *oauth) CreateSecretData(credentials *model.Credentials) (map[string][]byte, apperrors.AppError) {
	return svc.makeOauthMap(credentials.Oauth.ClientID, credentials.Oauth.ClientSecret), nil
}

func (svc *oauth) ToAppCredentials(credentials *model.Credentials, secretName string) applications.Credentials {
	applicationCredentials := applications.Credentials{
		AuthenticationUrl: credentials.Oauth.URL,
		Type:              applications.CredentialsOAuthType,
		SecretName:        secretName,
	}

	return applicationCredentials
}

func (svc *oauth) oauthCredentialsProvided(credentials *model.Credentials) bool {
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
