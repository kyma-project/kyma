package gateway_for_ns

import (
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"kyma-project.io/compass-runtime-agent/internal/apperrors"
	"kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/rafter"
	"kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/rafter/clusterassetgroup"
	"kyma-project.io/compass-runtime-agent/internal/kyma/model"
)

type Service interface {
	UpsertAPIResources(directorApplication model.Application, runtimeApplication v1alpha1.Application) apperrors.AppError
	DeleteAPIResources(runtimeApplication v1alpha1.Application) apperrors.AppError
	DeleteResourcesOfNonExistentAPI(existentRuntimeApplication v1alpha1.Application, directorApplication model.Application, name string) apperrors.AppError
}

type service struct {
	rafter rafter.Service
}

func NewService(rafter rafter.Service) Service {
	return &service{
		rafter: rafter,
	}
}

func (s service) UpsertAPIResources(directorApplication model.Application, runtimeApplication v1alpha1.Application) apperrors.AppError {

	for _, apiPackage := range directorApplication.APIPackages {
		err := s.upsertAPIResourcesForPackage(apiPackage)
		if err != nil {
			return err
		}
	}

	return nil
}

func createAssetFromAPIDefinition(apiDefinition model.APIDefinition) clusterassetgroup.Asset {
	return clusterassetgroup.Asset{
		Name:    apiDefinition.ID,
		Type:    getApiType(apiDefinition.APISpec),
		Content: getSpec(apiDefinition.APISpec),
		Format:  getSpecFormat(apiDefinition.APISpec),
	}
}

func (s service) upsertAPIResourcesForPackage(apiPackage model.APIPackage) apperrors.AppError {
	assets := make([]clusterassetgroup.Asset, 0, len(apiPackage.APIDefinitions)+len(apiPackage.EventDefinitions))
	for _, apiDefinition := range apiPackage.APIDefinitions {
		if apiDefinition.APISpec != nil {
			assets = append(assets, createAssetFromAPIDefinition(apiDefinition))
		}
	}

	for _, eventAPIDefinition := range apiPackage.EventDefinitions {
		if eventAPIDefinition.EventAPISpec != nil {
			assets = append(assets, createAssetFromEventAPIDefinition(eventAPIDefinition))
		}
	}

	return s.rafter.Put(apiPackage.ID, assets)
}

func (s service) DeleteAPIResources(runtimeApplication v1alpha1.Application) apperrors.AppError {

	for _, service := range runtimeApplication.Spec.Services {
		s.rafter.Delete(service.ID)
	}

	return nil
}

func (s service) DeleteResourcesOfNonExistentAPI(existentRuntimeApplication v1alpha1.Application, directorApplication model.Application, name string) apperrors.AppError {
	return nil
}

func createAssetFromEventAPIDefinition(eventAPIDefinition model.EventAPIDefinition) clusterassetgroup.Asset {

	return clusterassetgroup.Asset{
		Name:    eventAPIDefinition.ID,
		Type:    getEventApiType(eventAPIDefinition.EventAPISpec),
		Content: getEventSpec(eventAPIDefinition.EventAPISpec),
		Format:  clusterassetgroup.SpecFormat(getEventSpecFormat(eventAPIDefinition.EventAPISpec)),
	}

}

func getSpec(apiSpec *model.APISpec) []byte {
	if apiSpec == nil {
		return nil
	}

	return apiSpec.Data
}

func getEventSpec(eventApiSpec *model.EventAPISpec) []byte {
	if eventApiSpec == nil {
		return nil
	}

	return eventApiSpec.Data
}

func getSpecFormat(apiSpec *model.APISpec) clusterassetgroup.SpecFormat {
	if apiSpec == nil {
		return ""
	}
	return convertSpecFormat(apiSpec.Format)
}

func getEventSpecFormat(eventApiSpec *model.EventAPISpec) clusterassetgroup.SpecFormat {
	if eventApiSpec == nil {
		return ""
	}
	return convertSpecFormat(eventApiSpec.Format)
}

func getApiType(apiSpec *model.APISpec) clusterassetgroup.ApiType {
	if apiSpec == nil {
		return clusterassetgroup.Empty
	}
	if apiSpec.Type == model.APISpecTypeOdata {
		return clusterassetgroup.ODataApiType
	}
	if apiSpec.Type == model.APISpecTypeOpenAPI {
		return clusterassetgroup.OpenApiType
	}
	return clusterassetgroup.Empty
}

func getEventApiType(eventApiSpec *model.EventAPISpec) clusterassetgroup.ApiType {
	if eventApiSpec == nil {
		return clusterassetgroup.Empty
	}
	if eventApiSpec.Type == model.EventAPISpecTypeAsyncAPI {
		return clusterassetgroup.AsyncApi
	}
	return clusterassetgroup.Empty
}

func convertSpecFormat(specFormat model.SpecFormat) clusterassetgroup.SpecFormat {
	if specFormat == model.SpecFormatJSON {
		return clusterassetgroup.SpecFormatJSON
	}
	if specFormat == model.SpecFormatYAML {
		return clusterassetgroup.SpecFormatYAML
	}
	if specFormat == model.SpecFormatXML {
		return clusterassetgroup.SpecFormatXML
	}
	return ""
}
