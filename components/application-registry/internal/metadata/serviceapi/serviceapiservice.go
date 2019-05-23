package serviceapi

import (
	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-registry/internal/k8sconsts"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/accessservice"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/applications"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/istio"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/model"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/secrets"
)

// Service manages API definition of a service
type Service interface {
	// New handles a new API. It creates all requires resources.
	New(application, id string, api *model.API) (*applications.ServiceAPI, apperrors.AppError)
	// Read reads API from Application API definition. It also reads all additional information.
	Read(application string, serviceApi *applications.ServiceAPI) (*model.API, apperrors.AppError)
	// Delete removes API with given id.
	Delete(application, id string) apperrors.AppError
	// Update replaces existing API with a new one.
	Update(application, id string, api *model.API) (*applications.ServiceAPI, apperrors.AppError)
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

func (sas defaultService) New(application, id string, api *model.API) (*applications.ServiceAPI, apperrors.AppError) {
	resourceName := sas.nameResolver.GetResourceName(application, id)
	gatewayUrl := sas.nameResolver.GetGatewayUrl(application, id)

	serviceAPI := &applications.ServiceAPI{}
	serviceAPI.TargetUrl = api.TargetUrl
	serviceAPI.SpecificationUrl = api.SpecificationUrl
	serviceAPI.ApiType = api.ApiType
	serviceAPI.GatewayURL = gatewayUrl
	serviceAPI.AccessLabel = resourceName
	serviceAPI.Headers = api.Headers
	serviceAPI.QueryParameters = api.QueryParameters

	err := sas.accessServiceManager.Create(application, id, resourceName)
	if err != nil {
		return nil, apperrors.Internal("Creating access service failed, %s", err.Error())
	}

	if api.Credentials != nil {
		credentials, err := sas.secretsService.Create(application, id, api.Credentials)
		if err != nil {
			return nil, err
		}

		serviceAPI.Credentials = credentials
	}

	err = sas.istioService.Create(application, id, resourceName)
	if err != nil {
		return nil, apperrors.Internal("Creating Istio resources failed, %s", err.Error())
	}

	return serviceAPI, nil
}

func (sas defaultService) Read(application string, applicationAPI *applications.ServiceAPI) (*model.API, apperrors.AppError) {
	api := &model.API{
		TargetUrl:        applicationAPI.TargetUrl,
		SpecificationUrl: applicationAPI.SpecificationUrl,
		ApiType:          applicationAPI.ApiType,
		Headers:          applicationAPI.Headers,
		QueryParameters:  applicationAPI.QueryParameters,
	}

	if applicationAPI.Credentials.Type != "" {
		credentials, err := sas.secretsService.Get(application, applicationAPI.Credentials)
		if err != nil {
			return nil, err
		}

		api.Credentials = &credentials
	}

	return api, nil
}

func (sas defaultService) Delete(application, id string) apperrors.AppError {
	resourceName := sas.nameResolver.GetResourceName(application, id)

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

func (sas defaultService) Update(application, id string, api *model.API) (*applications.ServiceAPI, apperrors.AppError) {
	resourceName := sas.nameResolver.GetResourceName(application, id)
	gatewayUrl := sas.nameResolver.GetGatewayUrl(application, id)

	serviceAPI := &applications.ServiceAPI{}
	serviceAPI.TargetUrl = api.TargetUrl
	serviceAPI.SpecificationUrl = api.SpecificationUrl
	serviceAPI.ApiType = api.ApiType
	serviceAPI.GatewayURL = gatewayUrl
	serviceAPI.AccessLabel = resourceName
	serviceAPI.Headers = api.Headers
	serviceAPI.QueryParameters = api.QueryParameters

	err := sas.accessServiceManager.Upsert(application, id, resourceName)
	if err != nil {
		return nil, apperrors.Internal("Creating access service failed, %s", err.Error())
	}

	if api.Credentials != nil {
		credentials, err := sas.secretsService.Upsert(application, id, api.Credentials)
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

	err = sas.istioService.Upsert(application, id, resourceName)
	if err != nil {
		return nil, apperrors.Internal("Updating Istio resources failed, %s", err.Error())
	}

	return serviceAPI, nil
}
