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

	serviceAPI, err = sas.handleCredentialsCreate(remoteEnvironment, id, serviceAPI, api.Credentials)
	if err != nil {
		return nil, err
	}

	err = sas.istioService.Create(remoteEnvironment, id, resourceName)
	if err != nil {
		return nil, apperrors.Internal("Creating Istio resources failed, %s", err.Error())
	}

	return serviceAPI, nil
}

func (sas defaultService) Read(remoteEnvironment string, remoteenvAPI *remoteenv.ServiceAPI) (*API, apperrors.AppError) {
	api, err := sas.handleCredentialsFetch(remoteEnvironment, remoteenvAPI)
	if err != nil {
		return nil, err
	}

	api.TargetUrl = remoteenvAPI.TargetUrl

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

	serviceAPI, err = sas.handleCredentialsUpdate(remoteEnvironment, id, serviceAPI, api.Credentials)
	if err != nil {
		return nil, err
	}

	err = sas.istioService.Upsert(remoteEnvironment, id, resourceName)
	if err != nil {
		return nil, apperrors.Internal("Updating Istio resources failed, %s", err.Error())
	}

	return serviceAPI, nil
}

func (sas defaultService) handleCredentialsCreate(remoteEnvironment, id string, serviceAPI *remoteenv.ServiceAPI, credentials *Credentials) (*remoteenv.ServiceAPI, apperrors.AppError) {
	if credentials == nil {
		return serviceAPI, nil
	}

	if sas.basicCredentialsProvided(credentials) && sas.oauthCredentialsProvided(credentials) {
		return nil, apperrors.WrongInput("Creating access service failed: Multiple authentication methods provided.")
	}

	if sas.oauthCredentialsProvided(credentials) {
		serviceAPI.Credentials.AuthenticationUrl = credentials.Oauth.URL
	}

	name := sas.nameResolver.GetResourceName(remoteEnvironment, id)

	err := sas.createCredentialsSecret(remoteEnvironment, id, name, credentials)
	if err != nil {
		return nil, err
	}

	serviceAPI.Credentials.SecretName = sas.nameResolver.GetResourceName(remoteEnvironment, id)

	return serviceAPI, nil
}

func (sas defaultService) handleCredentialsUpdate(remoteEnvironment, id string, serviceAPI *remoteenv.ServiceAPI, credentials *Credentials) (*remoteenv.ServiceAPI, apperrors.AppError) {
	name := sas.nameResolver.GetResourceName(remoteEnvironment, id)

	if credentials == nil {
		return serviceAPI, sas.secretsService.DeleteSecret(name)
	}

	if sas.basicCredentialsProvided(credentials) && sas.oauthCredentialsProvided(credentials) {
		return nil, apperrors.WrongInput("Creating access service failed: Multiple authentication methods provided.")
	}

	if sas.oauthCredentialsProvided(credentials) {
		serviceAPI.Credentials.AuthenticationUrl = credentials.Oauth.URL
	}

	err := sas.updateCredentialsSecret(remoteEnvironment, id, name, credentials)
	if err != nil {
		return nil, err
	}

	serviceAPI.Credentials.SecretName = sas.nameResolver.GetResourceName(remoteEnvironment, id)

	return serviceAPI, nil
}

func (sas defaultService) handleCredentialsFetch(remoteEnvironment string, remoteenvAPI *remoteenv.ServiceAPI) (*API, apperrors.AppError) {
	api := &API{}

	if remoteenvAPI.Credentials.Type == remoteenv.CredentialsOAuthType {
		api.Credentials = &Credentials{
			Oauth: Oauth{
				URL: remoteenvAPI.Credentials.AuthenticationUrl,
			},
		}

		clientId, clientSecret, err := sas.secretsService.GetOauthSecret(remoteEnvironment, remoteenvAPI.Credentials.SecretName)
		if err != nil {
			return nil, apperrors.Internal("Reading oauth credentials from %s secret failed, %s",
				remoteenvAPI.Credentials.SecretName, err.Error())
		}
		api.Credentials.Oauth.ClientID = clientId
		api.Credentials.Oauth.ClientSecret = clientSecret
	}

	if remoteenvAPI.Credentials.Type == remoteenv.CredentialsBasicType {
		api.Credentials = &Credentials{
			Basic: Basic{},
		}

		username, password, err := sas.secretsService.GetBasicAuthSecret(remoteEnvironment, remoteenvAPI.Credentials.SecretName)
		if err != nil {
			return nil, apperrors.Internal("Reading oauth credentials from %s secret failed, %s",
				remoteenvAPI.Credentials.SecretName, err.Error())
		}
		api.Credentials.Basic.Username = username
		api.Credentials.Basic.Password = password
	}

	return api, nil
}

func (sas defaultService) createCredentialsSecret(remoteEnvironment, id, name string, credentials *Credentials) apperrors.AppError {
	if sas.oauthCredentialsProvided(credentials) {
		return sas.secretsService.CreateOauthSecret(
			remoteEnvironment,
			name,
			credentials.Oauth.ClientID,
			credentials.Oauth.ClientSecret,
			id,
		)
	}

	if sas.basicCredentialsProvided(credentials) {
		return sas.secretsService.CreateBasicAuthSecret(remoteEnvironment,
			name,
			credentials.Basic.Username,
			credentials.Basic.Password,
			id,
		)
	}

	return nil
}

func (sas defaultService) updateCredentialsSecret(remoteEnvironment, id, name string, credentials *Credentials) apperrors.AppError {
	if sas.oauthCredentialsProvided(credentials) {
		return sas.secretsService.UpdateOauthSecret(
			remoteEnvironment,
			name,
			credentials.Oauth.ClientID,
			credentials.Oauth.ClientSecret,
			id,
		)
	}

	if sas.basicCredentialsProvided(credentials) {
		return sas.secretsService.UpdateBasicAuthSecret(remoteEnvironment,
			name,
			credentials.Basic.Username,
			credentials.Basic.Password,
			id,
		)
	}

	return nil
}

func (sas defaultService) oauthCredentialsProvided(credentials *Credentials) bool {
	return credentials != nil && credentials.Oauth.ClientID != "" && credentials.Oauth.ClientSecret != ""
}

func (sas defaultService) basicCredentialsProvided(credentials *Credentials) bool {
	return credentials != nil && credentials.Basic.Username != "" && credentials.Basic.Password != ""
}
