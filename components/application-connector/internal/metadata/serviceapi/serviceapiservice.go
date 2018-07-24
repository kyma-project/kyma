package serviceapi

import (
	"github.com/kyma-project/kyma/components/application-connector/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-connector/internal/k8sconsts"
	"github.com/kyma-project/kyma/components/application-connector/internal/metadata/accessservice"
	"github.com/kyma-project/kyma/components/application-connector/internal/metadata/istio"
	"github.com/kyma-project/kyma/components/application-connector/internal/metadata/remoteenv"
	"github.com/kyma-project/kyma/components/application-connector/internal/metadata/secrets"
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
	// New handles a new API. It creates all requires resources.
	New(remoteEnvironment, id string, api *API) (*remoteenv.ServiceAPI, apperrors.AppError)
	// Read reads API from Remote Environment API definition. It also reads all additional information.
	Read(remoteEnvironment string, serviceApi *remoteenv.ServiceAPI) (*API, apperrors.AppError)
	// Delete removes API with given id.
	Delete(remoteEnvironment, id string) apperrors.AppError
	// Update replaces existing API with a new one.
	Update(remoteEnvironment, id string, api *API) (*remoteenv.ServiceAPI, apperrors.AppError)
}

type defaultService struct {
	nameResolver         k8sconsts.NameResolver
	accessServiceManager accessservice.AccessServiceManager
	secretsRepository    secrets.Repository
	istioService         istio.Service
}

func NewService(
	nameResolver k8sconsts.NameResolver,
	accessServiceManager accessservice.AccessServiceManager,
	secretsRepository secrets.Repository,
	istioService istio.Service) Service {

	return defaultService{
		nameResolver:         nameResolver,
		accessServiceManager: accessServiceManager,
		secretsRepository:    secretsRepository,
		istioService:         istioService,
	}
}

func (sas defaultService) New(remoteEnvironment, id string, api *API) (*remoteenv.ServiceAPI, apperrors.AppError) {
	resourceName := sas.nameResolver.GetResourceName(remoteEnvironment, id)
	gatewayUrl := sas.nameResolver.GetMetadataUrl(remoteEnvironment, id)

	serviceAPI := &remoteenv.ServiceAPI{}
	serviceAPI.TargetUrl = api.TargetUrl
	serviceAPI.GatewayURL = gatewayUrl

	err := sas.accessServiceManager.Create(remoteEnvironment, id, resourceName)
	if err != nil {
		return nil, apperrors.Internal("failed to create access service, %s", err)
	}

	serviceAPI.AccessLabel = resourceName

	if sas.oauthCredentialsProvided(api.Credentials) {
		err := sas.secretsRepository.Create(remoteEnvironment, resourceName, api.Credentials.Oauth.ClientID, api.Credentials.Oauth.ClientSecret, id)
		if err != nil {
			return nil, apperrors.Internal("failed to create credentials secret, %s", err)
		}
		serviceAPI.OauthUrl = api.Credentials.Oauth.URL
		serviceAPI.CredentialsSecretName = resourceName
	}

	err = sas.istioService.Create(remoteEnvironment, id, resourceName)
	if err != nil {
		return nil, apperrors.Internal("failed to create istio resources, %s", err)
	}

	return serviceAPI, nil
}

func (sas defaultService) Read(remoteEnvironment string, remoteenvAPI *remoteenv.ServiceAPI) (*API, apperrors.AppError) {
	api := &API{
		TargetUrl: remoteenvAPI.TargetUrl,
	}

	if remoteenvAPI.OauthUrl != "" && remoteenvAPI.CredentialsSecretName != "" {
		api.Credentials = &Credentials{
			Oauth: Oauth{
				URL: remoteenvAPI.OauthUrl,
			},
		}

		clientId, clientSecret, err := sas.secretsRepository.Get(remoteEnvironment, remoteenvAPI.CredentialsSecretName)
		if err != nil {
			return nil, apperrors.Internal("failed to read oauth credentials from %s secret, %s",
				remoteenvAPI.CredentialsSecretName, err.Error())
		}
		api.Credentials.Oauth.ClientID = clientId
		api.Credentials.Oauth.ClientSecret = clientSecret
	}

	return api, nil
}

func (sas defaultService) Delete(remoteEnvironment, id string) apperrors.AppError {
	resourceName := sas.nameResolver.GetResourceName(remoteEnvironment, id)

	err := sas.accessServiceManager.Delete(resourceName)
	if err != nil {
		return apperrors.Internal("failed to delete access service, %s", err)
	}

	err = sas.secretsRepository.Delete(resourceName)
	if err != nil {
		return apperrors.Internal("failed to delete credentials secret, %s", err)
	}

	err = sas.istioService.Delete(resourceName)
	if err != nil {
		return apperrors.Internal("failed to delete istio resources, %s", err)
	}

	return nil
}

func (sas defaultService) Update(remoteEnvironment, id string, api *API) (*remoteenv.ServiceAPI, apperrors.AppError) {
	resourceName := sas.nameResolver.GetResourceName(remoteEnvironment, id)
	gatewayUrl := sas.nameResolver.GetMetadataUrl(remoteEnvironment, id)

	serviceAPI := &remoteenv.ServiceAPI{}
	serviceAPI.TargetUrl = api.TargetUrl
	serviceAPI.GatewayURL = gatewayUrl

	err := sas.accessServiceManager.Upsert(remoteEnvironment, id, resourceName)
	if err != nil {
		return nil, apperrors.Internal("failed to create access service, %s", err)
	}

	serviceAPI.AccessLabel = resourceName

	if sas.oauthCredentialsProvided(api.Credentials) {
		err = sas.secretsRepository.Upsert(remoteEnvironment, resourceName, api.Credentials.Oauth.ClientID, api.Credentials.Oauth.ClientSecret, id)
		if err != nil {
			return nil, apperrors.Internal("failed to update credentials secret, %s", err)
		}
		serviceAPI.OauthUrl = api.Credentials.Oauth.URL
		serviceAPI.CredentialsSecretName = resourceName
	} else {
		err := sas.secretsRepository.Delete(resourceName)
		if err != nil {
			return nil, apperrors.Internal("failed to delete credentials secret, %s", err)
		}
	}

	err = sas.istioService.Upsert(remoteEnvironment, id, resourceName)
	if err != nil {
		return nil, apperrors.Internal("failed to update istio resources, %s", err)
	}

	return serviceAPI, nil
}

func (sas defaultService) oauthCredentialsProvided(credentials *Credentials) bool {
	return credentials != nil && credentials.Oauth.ClientID != "" && credentials.Oauth.ClientSecret != ""
}
