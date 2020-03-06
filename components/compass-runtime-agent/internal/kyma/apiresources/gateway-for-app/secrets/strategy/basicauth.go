package strategy

import (
	"kyma-project.io/compass-runtime-agent/internal/apperrors"
	"kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/gateway-for-app/secrets/model"
	"kyma-project.io/compass-runtime-agent/internal/kyma/applications"
	"kyma-project.io/compass-runtime-agent/internal/kyma/applications/converters"
)

const (
	BasicAuthUsernameKey = "username"
	BasicAuthPasswordKey = "password"
)

type basicAuth struct{}

func (svc *basicAuth) ToCredentials(secretData SecretData, appCredentials *applications.Credentials) model.CredentialsWithCSRF {
	username, password := svc.readBasicAuthMap(secretData)

	return model.CredentialsWithCSRF{
		Basic: &model.Basic{
			Username: username,
			Password: password,
		},
		CSRFInfo: convertToModelCSRInfo(appCredentials),
	}
}

func (svc *basicAuth) CredentialsProvided(credentials *model.CredentialsWithCSRF) bool {
	return svc.basicCredentialsProvided(credentials)
}

func (svc *basicAuth) CreateSecretData(credentials *model.CredentialsWithCSRF) (SecretData, apperrors.AppError) {
	return svc.makeBasicAuthMap(credentials.Basic.Username, credentials.Basic.Password), nil
}

func (svc *basicAuth) ToCredentialsInfo(credentials *model.CredentialsWithCSRF, secretName string) applications.Credentials {
	applicationCredentials := applications.Credentials{
		Type:       converters.CredentialsBasicType,
		SecretName: secretName,
		CSRFInfo:   toAppCSRFInfo(credentials),
	}

	return applicationCredentials
}

func (svc *basicAuth) ShouldUpdate(currentData SecretData, newData SecretData) bool {
	return string(currentData[BasicAuthUsernameKey]) != string(newData[BasicAuthUsernameKey]) ||
		string(currentData[BasicAuthPasswordKey]) != string(newData[BasicAuthPasswordKey])
}

func (svc *basicAuth) makeBasicAuthMap(username, password string) map[string][]byte {
	return map[string][]byte{
		BasicAuthUsernameKey: []byte(username),
		BasicAuthPasswordKey: []byte(password),
	}
}

func (svc *basicAuth) readBasicAuthMap(data map[string][]byte) (username, password string) {
	return string(data[BasicAuthUsernameKey]), string(data[BasicAuthPasswordKey])
}

func (svc *basicAuth) basicCredentialsProvided(credentials *model.CredentialsWithCSRF) bool {
	return credentials != nil && credentials.Basic != nil
}
