package strategy

import (
	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/applications"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/model"
)

const (
	BasicAuthUsernameKey = "username"
	BasicAuthPasswordKey = "password"
)

// TODO - add tests
type basicAuth struct{}

func (svc *basicAuth) ToCredentials(secretData map[string][]byte, appCredentials *applications.Credentials) model.Credentials {
	username, password := svc.readBasicAuthMap(secretData)

	return model.Credentials{
		Basic: &model.Basic{
			Username: username,
			Password: password,
		},
	}
}

// TODO - what about passing basicCredentials in factory as they are passed to each function?
func (svc *basicAuth) CredentialsProvided(credentials *model.Credentials) bool {
	return svc.basicCredentialsProvided(credentials)
}

func (svc *basicAuth) CreateSecretData(credentials *model.Credentials) (map[string][]byte, apperrors.AppError) {
	return svc.makeBasicAuthMap(credentials.Basic.Username, credentials.Basic.Password), nil
}

func (svc *basicAuth) ToAppCredentials(credentials *model.Credentials, secretName string) applications.Credentials {
	applicationCredentials := applications.Credentials{
		Type:       applications.CredentialsBasicType,
		SecretName: secretName,
	}

	return applicationCredentials
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
