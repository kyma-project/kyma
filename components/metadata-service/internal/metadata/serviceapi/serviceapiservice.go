package serviceapi

import (
	"github.com/kyma-project/kyma/components/metadata-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/metadata-service/internal/k8sconsts"
	"github.com/kyma-project/kyma/components/metadata-service/internal/metadata/accessservice"
	"github.com/kyma-project/kyma/components/metadata-service/internal/metadata/istio"
	"github.com/kyma-project/kyma/components/metadata-service/internal/metadata/remoteenv"
	"github.com/kyma-project/kyma/components/metadata-service/internal/metadata/secrets"
)

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
	secretsService       secrets.Service
	istioService         istio.Service
}

func NewService(
	nameResolver k8sconsts.NameResolver,
	accessServiceManager accessservice.AccessServiceManager,
	secretsService secrets.Service,
	istioService istio.Service) Service {

	return defaultService{
		nameResolver:         nameResolver,
		accessServiceManager: accessServiceManager,
		secretsService:       secretsService,
		istioService:         istioService,
	}
}

//TODO:
func (sas defaultService) New(remoteEnvironment, id string, api *API) (*remoteenv.ServiceAPI, apperrors.AppError) {
	resourceName := sas.nameResolver.GetResourceName(remoteEnvironment, id)
	gatewayUrl := sas.nameResolver.GetGatewayUrl(remoteEnvironment, id)

	serviceAPI := &remoteenv.ServiceAPI{}
	serviceAPI.TargetUrl = api.TargetUrl
	serviceAPI.GatewayURL = gatewayUrl

	err := sas.accessServiceManager.Create(remoteEnvironment, id, resourceName)
	if err != nil {
		return nil, apperrors.Internal("Creating access service failed, %s", err.Error())
	}

	serviceAPI.AccessLabel = resourceName

	err = sas.handleCredentials(remoteEnvironment, id, serviceAPI, api.Credentials)
	if err != nil {
		return nil, err
	}

	err = sas.istioService.Create(remoteEnvironment, id, resourceName)
	if err != nil {
		return nil, apperrors.Internal("Creating Istio resources failed, %s", err.Error())
	}

	return serviceAPI, nil
}

//TODO:
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

		clientId, clientSecret, err := sas.secretsService.GetOauthSecret(remoteEnvironment, remoteenvAPI.CredentialsSecretName)
		if err != nil {
			return nil, apperrors.Internal("Reading oauth credentials from %s secret failed, %s",
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
		return apperrors.Internal("Deleting access service failed, %s", err.Error())
	}

	err = sas.secretsService.DeleteSecret(resourceName)
	if err != nil {
		return apperrors.Internal("Deleting credentials secret failed, %s", err.Error())
	}

	err = sas.istioService.Delete(resourceName)
	if err != nil {
		return apperrors.Internal("Deleting Istio resources failed, %s", err.Error())
	}

	return nil
}

//TODO:
func (sas defaultService) Update(remoteEnvironment, id string, api *API) (*remoteenv.ServiceAPI, apperrors.AppError) {
	resourceName := sas.nameResolver.GetResourceName(remoteEnvironment, id)
	gatewayUrl := sas.nameResolver.GetGatewayUrl(remoteEnvironment, id)

	serviceAPI := &remoteenv.ServiceAPI{}
	serviceAPI.TargetUrl = api.TargetUrl
	serviceAPI.GatewayURL = gatewayUrl

	err := sas.accessServiceManager.Upsert(remoteEnvironment, id, resourceName)
	if err != nil {
		return nil, apperrors.Internal("Creating access service failed, %s", err.Error())
	}

	serviceAPI.AccessLabel = resourceName

	if sas.oauthCredentialsProvided(api.Credentials) {
		err = sas.secretsService.UpdateOauthSecret(remoteEnvironment, resourceName, api.Credentials.Oauth.ClientID, api.Credentials.Oauth.ClientSecret, id)
		if err != nil {
			return nil, apperrors.Internal("Updating credentials secret failed, %s", err.Error())
		}
		serviceAPI.OauthUrl = api.Credentials.Oauth.URL
		serviceAPI.CredentialsSecretName = resourceName
	} else {
		err := sas.secretsService.DeleteSecret(resourceName)
		if err != nil {
			return nil, apperrors.Internal("Deleting credentials secret failed, %s", err.Error())
		}
	}

	err = sas.istioService.Upsert(remoteEnvironment, id, resourceName)
	if err != nil {
		return nil, apperrors.Internal("Updating Istio resources failed, %s", err.Error())
	}

	return serviceAPI, nil
}

//TODO:
func (sas defaultService) handleCredentials(remoteEnvironment, id string, serviceAPI *remoteenv.ServiceAPI, credentials *Credentials) apperrors.AppError {
	if credentials == nil {
		return nil
	}

	if sas.basicCredentialsProvided(credentials) && sas.oauthCredentialsProvided(credentials) {
		return apperrors.WrongInput("Creating access service failed: Multiple authentication methods provided.")
	}

	if sas.oauthCredentialsProvided(credentials) {
		serviceAPI.OauthUrl = credentials.Oauth.URL
	}

	err := sas.createCredentialsSecret(remoteEnvironment, id, credentials)
	if err != nil {
		return err
	}

	return nil
}

//TODO:
func (sas defaultService) createCredentialsSecret(remoteEnvironment, id string, credentials *Credentials) apperrors.AppError {
	return nil
}

func (sas defaultService) oauthCredentialsProvided(credentials *Credentials) bool {
	return credentials != nil && credentials.Oauth.ClientID != "" && credentials.Oauth.ClientSecret != ""
}

func (sas defaultService) basicCredentialsProvided(credentials *Credentials) bool {
	return credentials != nil && credentials.Basic.Username != "" && credentials.Basic.Password != ""
}
