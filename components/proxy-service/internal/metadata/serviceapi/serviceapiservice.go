package serviceapi

import (
	"github.com/kyma-project/kyma/components/proxy-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/proxy-service/internal/metadata/remoteenv"
	"github.com/kyma-project/kyma/components/proxy-service/internal/metadata/secrets"
)

// API is an internal representation of a service's API.
type API struct {
	// TargetUrl points to API.
	TargetUrl string
	// Credentials is a credentials of API.
	Credentials *Credentials
	// Spec contains specification of an API.
	Spec []byte
}

// Credentials contains OAuth configuration.
type Credentials struct {
	// Oauth is OAuth configuration.
	Oauth *Oauth
	Basic *Basic
}

type Basic struct {
	Username string
	Password string
}

// Oauth contains details of OAuth configuration
type Oauth struct {
	// URL to OAuth token provider.
	URL string
	// ClientID to use for authentication.
	ClientID string
	// ClientSecret to use for authentication.
	ClientSecret string
}

const (
	ClientIDKey     = "clientId"
	ClientSecretKey = "clientSecret"
	UsernameKey = "username"
	PasswordKey = "password"
	TypeOAuth = "OAuth"
	TypeBasic = "Basic"
)

// Service manages API definition of a service
type Service interface {
	// Read reads API from Remote Environment API definition. It also reads all additional information.
	Read(*remoteenv.ServiceAPI) (*API, apperrors.AppError)
}

type defaultService struct {
	secretsRepository secrets.Repository
}

func NewService(secretsRepository secrets.Repository) Service {

	return defaultService{
		secretsRepository: secretsRepository,
	}
}

func (sas defaultService) Read(remoteenvAPI *remoteenv.ServiceAPI) (*API, apperrors.AppError) {
	api := &API{
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
			api.Credentials = &Credentials{
				Oauth: getOAuthCredentials(secret),
			}
		} else if credentialsType == TypeBasic {
			api.Credentials = &Credentials{
				Basic: getBasicAuthCredentials(secret),
			}
		} else {
		 	api.Credentials = nil
		}
	}

	return api, nil
}


func getOAuthCredentials(secret map[string][]byte) *Oauth{
	return &Oauth{
		ClientID: string(secret[ClientIDKey]),
		ClientSecret: string(secret[ClientSecretKey]),
	}
}

func getBasicAuthCredentials(secret map[string][]byte) *Basic{
	return &Basic{
		Username: string(secret[UsernameKey]),
		Password: string(secret[PasswordKey]),
	}
}