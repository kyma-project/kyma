package kyma

import (
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/sirupsen/logrus"
	"kyma-project.io/compass-runtime-agent/internal/apperrors"
	"kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/rafter/clusterassetgroup"
	"kyma-project.io/compass-runtime-agent/internal/kyma/model"
)

func createAssetFromEventAPIDefinition(eventAPIDefinition model.EventAPIDefinition) clusterassetgroup.Asset {

	return clusterassetgroup.Asset{
		Name:    eventAPIDefinition.Name,
		Type:    getEventApiType(eventAPIDefinition.EventAPISpec),
		Content: getEventSpec(eventAPIDefinition.EventAPISpec),
		Format:  clusterassetgroup.SpecFormat(getEventSpecFormat(eventAPIDefinition.EventAPISpec)),
	}
}

func createAssetsFromEventAPIDefinition(eventAPIDefinition model.EventAPIDefinition) []clusterassetgroup.Asset {
	// <AG>
	logrus.Infof("Creating asset from Event API Definition: %v", eventAPIDefinition)
	// <AG>

	if eventAPIDefinition.EventAPISpec != nil {
		return []clusterassetgroup.Asset{
			createAssetFromEventAPIDefinition(eventAPIDefinition),
		}
	}

	return nil

}

func createAssetFromAPIDefinition(apiDefinition model.APIDefinition) clusterassetgroup.Asset {
	// <AG>
	logrus.Infof("Creating asset from API Definition: %v", apiDefinition)
	// <AG>

	return clusterassetgroup.Asset{
		Name:    apiDefinition.Name,
		Type:    getApiType(apiDefinition.APISpec),
		Content: getSpec(apiDefinition.APISpec),
		Format:  getSpecFormat(apiDefinition.APISpec),
	}
}

func createAssetsFromAPIDefinition(apiDefinition model.APIDefinition) []clusterassetgroup.Asset {

	if apiDefinition.APISpec != nil {
		return []clusterassetgroup.Asset{
			createAssetFromAPIDefinition(apiDefinition),
		}
	}

	return nil
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

func newResult(application v1alpha1.Application, applicationID string, operation Operation, appError apperrors.AppError) Result {
	return Result{
		ApplicationName: application.Name,
		ApplicationID:   applicationID,
		Operation:       operation,
		Error:           appError,
	}
}

func ServiceExists(id string, application v1alpha1.Application) bool {
	return getServiceIndex(id, application) != -1
}

func GetService(id string, application v1alpha1.Application) v1alpha1.Service {
	for _, service := range application.Spec.Services {
		if service.ID == id {
			return service
		}
	}

	return v1alpha1.Service{}
}

func getServiceIndex(id string, application v1alpha1.Application) int {
	for i, service := range application.Spec.Services {
		if service.ID == id {
			return i
		}
	}

	return -1
}

func ApplicationExists(applicationName string, applicationList []v1alpha1.Application) bool {
	if applicationList == nil {
		return false
	}

	for _, runtimeApplication := range applicationList {
		if runtimeApplication.Name == applicationName {
			return true
		}
	}

	return false
}

func GetApplication(applicationName string, applicationList []v1alpha1.Application) v1alpha1.Application {
	if applicationList == nil {
		return v1alpha1.Application{}
	}

	for _, runtimeApplication := range applicationList {
		if runtimeApplication.Name == applicationName {
			return runtimeApplication
		}
	}

	return v1alpha1.Application{}
}
