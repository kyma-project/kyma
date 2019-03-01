package strategy

import (
	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/applications"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/model"
)

const (
	SCRFTokenAuthEndpoint = "authEndpoint"
)

type csrfToken struct{}

func (svc *csrfToken) ToCredentials(secretData SecretData, appCredentials *applications.Credentials) model.Credentials {
	authEndpoint := svc.readCSRFTokenMap(secretData)

	return model.Credentials{
		CSRFToken: &model.CSRFToken{
			AuthEndpoint: authEndpoint,
		},
	}
}

func (svc *csrfToken) CredentialsProvided(credentials *model.Credentials) bool {
	return svc.csrfTokenCredentialsProvided(credentials)
}

func (svc *csrfToken) CreateSecretData(credentials *model.Credentials) (SecretData, apperrors.AppError) {
	return svc.makeCSRFTokenMap(credentials.CSRFToken.AuthEndpoint), nil
}

func (svc *csrfToken) ToCredentialsInfo(credentials *model.Credentials, secretName string) applications.Credentials {
	applicationCredentials := applications.Credentials{
		Type:       applications.CredentialsCSRFTokenType,
		SecretName: secretName,
	}

	return applicationCredentials
}

func (svc *csrfToken) ShouldUpdate(currentData SecretData, newData SecretData) bool {
	return string(currentData[SCRFTokenAuthEndpoint]) != string(newData[SCRFTokenAuthEndpoint])
}

func (svc *csrfToken) readCSRFTokenMap(data map[string][]byte) (authEndpoint string) {
	return string(data[SCRFTokenAuthEndpoint])
}

func (svc *csrfToken) csrfTokenCredentialsProvided(credentials *model.Credentials) bool {
	return credentials != nil && credentials.CSRFToken != nil && credentials.CSRFToken.AuthEndpoint != ""
}

func (svc *csrfToken) makeCSRFTokenMap(authEndpoint string) map[string][]byte {
	return map[string][]byte{
		SCRFTokenAuthEndpoint: []byte(authEndpoint),
	}
}