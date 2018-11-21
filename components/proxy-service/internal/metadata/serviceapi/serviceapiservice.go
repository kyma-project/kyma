package serviceapi

import (
	"github.com/kyma-project/kyma/components/proxy-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/proxy-service/internal/metadata/model"
	"github.com/kyma-project/kyma/components/proxy-service/internal/metadata/remoteenv"
	"github.com/kyma-project/kyma/components/proxy-service/internal/metadata/secrets"
)

const (
	ClientIDKey     = "clientId"
	ClientSecretKey = "clientSecret"
	UsernameKey     = "username"
	PasswordKey     = "password"
	TypeOAuth       = "OAuth"
	TypeBasic       = "BasicAuth"
)

// Service manages API definition of a service
type Service interface {
	// Read reads API from Remote Environment API definition. It also reads all additional information.
	Read(*remoteenv.ServiceAPI) (*model.API, apperrors.AppError)
}

type defaultService struct {
	secretsRepository secrets.Repository
}

func NewService(secretsRepository secrets.Repository) Service {

	return defaultService{
		secretsRepository: secretsRepository,
	}
}

func (sas defaultService) Read(remoteenvAPI *remoteenv.ServiceAPI) (*model.API, apperrors.AppError) {
	api := &model.API{
		TargetUrl: remoteenvAPI.TargetUrl,
	}

	if remoteenvAPI.Credentials != nil {
		credentialsSecretName := remoteenvAPI.Credentials.SecretName
		credentialsType := remoteenvAPI.Credentials.Type

		secret, err := sas.secretsRepository.Get(credentialsSecretName)

		if err != nil {
			return nil, err
		}

		if credentialsType == TypeOAuth {
			api.Credentials = &model.Credentials{
				OAuth: getOAuthCredentials(secret, remoteenvAPI.Credentials.Url),
			}
		} else if credentialsType == TypeBasic {
			api.Credentials = &model.Credentials{
				BasicAuth: getBasicAuthCredentials(secret),
			}
		} else {
			api.Credentials = nil
		}
	}

	return api, nil
}

func getOAuthCredentials(secret map[string][]byte, url string) *model.OAuth {
	return &model.OAuth{
		ClientID:     string(secret[ClientIDKey]),
		ClientSecret: string(secret[ClientSecretKey]),
		URL:          url,
	}
}

func getBasicAuthCredentials(secret map[string][]byte) *model.BasicAuth {
	return &model.BasicAuth{
		Username: string(secret[UsernameKey]),
		Password: string(secret[PasswordKey]),
	}
}
