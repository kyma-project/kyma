package strategy

import (
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/apperrors"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/applications"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/model"
)

const (
	BasicAuthUsernameKey = "username"
	BasicAuthPasswordKey = "password"
)

type basicAuth struct{}

func (svc *basicAuth) ToCredentials(secretData SecretData, appCredentials *applications.Credentials) (model.Credentials, apperrors.AppError) {
	username, password := svc.readBasicAuthMap(secretData)
	return model.Credentials{
		Basic: &model.Basic{
			Username: username,
			Password: password,
		},
		CSRFInfo:          convertToModelCSRInfo(appCredentials),
	}, nil
}

func (svc *basicAuth) CredentialsProvided(credentials *model.Credentials) bool {
	return svc.basicCredentialsProvided(credentials)
}

func (svc *basicAuth) CreateSecretData(credentials *model.Credentials) (SecretData, apperrors.AppError) {
	return svc.makeBasicAuthMap(credentials.Basic.Username, credentials.Basic.Password), nil
}

func (svc *basicAuth) ToCredentialsInfo(credentials *model.Credentials, secretName string) applications.Credentials {
	applicationCredentials := applications.Credentials{
		Type:       applications.CredentialsBasicType,
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

func (svc *basicAuth) basicCredentialsProvided(credentials *model.Credentials) bool {
	return credentials != nil && credentials.Basic != nil && credentials.Basic.Username != "" && credentials.Basic.Password != ""
}
