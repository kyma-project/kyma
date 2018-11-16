package serviceapi

import (
	"github.com/kyma-project/kyma/components/metadata-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/metadata-service/internal/k8sconsts"
	"github.com/kyma-project/kyma/components/metadata-service/internal/metadata/accessservice"
	"github.com/kyma-project/kyma/components/metadata-service/internal/metadata/istio"
	"github.com/kyma-project/kyma/components/metadata-service/internal/metadata/model"
	"github.com/kyma-project/kyma/components/metadata-service/internal/metadata/remoteenv"
	"github.com/kyma-project/kyma/components/metadata-service/internal/metadata/secrets"
)

// Service manages API definition of a service
type Service interface {
	// New handles a new API. It creates all requires resources.
	New(remoteEnvironment, id string, api *model.API) (*remoteenv.ServiceAPI, apperrors.AppError)
	// Read reads API from Remote Environment API definition. It also reads all additional information.
	Read(remoteEnvironment string, serviceApi *remoteenv.ServiceAPI) (*model.API, apperrors.AppError)
	// Delete removes API with given id.
	Delete(remoteEnvironment, id string) apperrors.AppError
	// Update replaces existing API with a new one.
	Update(remoteEnvironment, id string, api *model.API) (*remoteenv.ServiceAPI, apperrors.AppError)
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

func (sas defaultService) New(remoteEnvironment, id string, api *model.API) (*remoteenv.ServiceAPI, apperrors.AppError) {
	resourceName := sas.nameResolver.GetResourceName(remoteEnvironment, id)
	gatewayUrl := sas.nameResolver.GetGatewayUrl(remoteEnvironment, id)

	serviceAPI := &remoteenv.ServiceAPI{}
	serviceAPI.TargetUrl = api.TargetUrl
	serviceAPI.SpecificationUrl = api.SpecificationUrl
	serviceAPI.ApiType = api.ApiType
	serviceAPI.GatewayURL = gatewayUrl
	serviceAPI.AccessLabel = resourceName

	err := sas.accessServiceManager.Create(remoteEnvironment, id, resourceName)
	if err != nil {
		return nil, apperrors.Internal("Creating access service failed, %s", err.Error())
	}

	if api.Credentials != nil {
		credentials, err := sas.secretsService.Create(remoteEnvironment, id, api.Credentials)
		if err != nil {
			return nil, err
		}

		serviceAPI.Credentials = credentials
	}

	err = sas.istioService.Create(remoteEnvironment, id, resourceName)
	if err != nil {
		return nil, apperrors.Internal("Creating Istio resources failed, %s", err.Error())
	}

	return serviceAPI, nil
}

func (sas defaultService) Read(remoteEnvironment string, remoteenvAPI *remoteenv.ServiceAPI) (*model.API, apperrors.AppError) {
	api := &model.API{
		TargetUrl:        remoteenvAPI.TargetUrl,
		SpecificationUrl: remoteenvAPI.SpecificationUrl,
		ApiType:          remoteenvAPI.ApiType,
	}

	if remoteenvAPI.Credentials.Type != "" {
		credentials, err := sas.secretsService.Get(remoteEnvironment, remoteenvAPI.Credentials)
		if err != nil {
			return nil, err
		}

		api.Credentials = &credentials
	}

	return api, nil
}

func (sas defaultService) Delete(remoteEnvironment, id string) apperrors.AppError {
	resourceName := sas.nameResolver.GetResourceName(remoteEnvironment, id)

	err := sas.accessServiceManager.Delete(resourceName)
	if err != nil {
		return apperrors.Internal("Deleting access service failed, %s", err.Error())
	}

	err = sas.secretsService.Delete(resourceName)
	if err != nil {
		return apperrors.Internal("Deleting credentials secret failed, %s", err.Error())
	}

	err = sas.istioService.Delete(resourceName)
	if err != nil {
		return apperrors.Internal("Deleting Istio resources failed, %s", err.Error())
	}

	return nil
}

func (sas defaultService) Update(remoteEnvironment, id string, api *model.API) (*remoteenv.ServiceAPI, apperrors.AppError) {
	resourceName := sas.nameResolver.GetResourceName(remoteEnvironment, id)
	gatewayUrl := sas.nameResolver.GetGatewayUrl(remoteEnvironment, id)

	serviceAPI := &remoteenv.ServiceAPI{}
	serviceAPI.TargetUrl = api.TargetUrl
	serviceAPI.SpecificationUrl = api.SpecificationUrl
	serviceAPI.ApiType = api.ApiType
	serviceAPI.GatewayURL = gatewayUrl
	serviceAPI.AccessLabel = resourceName

	err := sas.accessServiceManager.Upsert(remoteEnvironment, id, resourceName)
	if err != nil {
		return nil, apperrors.Internal("Creating access service failed, %s", err.Error())
	}

	if api.Credentials != nil {
		credentials, err := sas.secretsService.Update(remoteEnvironment, id, api.Credentials)
		if err != nil {
			return nil, err
		}

		serviceAPI.Credentials = credentials
	} else {
		err := sas.secretsService.Delete(resourceName)
		if err != nil {
			return nil, err
		}
	}

	err = sas.istioService.Upsert(remoteEnvironment, id, resourceName)
	if err != nil {
		return nil, apperrors.Internal("Updating Istio resources failed, %s", err.Error())
	}

	return serviceAPI, nil
}
