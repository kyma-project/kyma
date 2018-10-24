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
	Oauth Oauth
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

	if remoteenvAPI.OauthUrl != "" && remoteenvAPI.CredentialsSecretName != "" {
		api.Credentials = &Credentials{
			Oauth: Oauth{
				URL: remoteenvAPI.OauthUrl,
			},
		}

		clientId, clientSecret, err := sas.secretsRepository.Get(remoteenvAPI.CredentialsSecretName)
		if err != nil {
			return nil, apperrors.Internal("failed to read oauth credentials from %s secret, %s",
				remoteenvAPI.CredentialsSecretName, err.Error())
		}
		api.Credentials.Oauth.ClientID = clientId
		api.Credentials.Oauth.ClientSecret = clientSecret
	}

	return api, nil
}
