package kyma

import (
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"kyma-project.io/compass-runtime-agent/internal/apperrors"
	"kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/rafter/clusterassetgroup"
	"kyma-project.io/compass-runtime-agent/internal/kyma/model"
)

func createAssetFromEventAPIDefinition(eventAPIDefinition model.EventAPIDefinition) clusterassetgroup.Asset {

	return clusterassetgroup.Asset{
		ID:      eventAPIDefinition.ID,
		Name:    eventAPIDefinition.Name,
		Type:    getEventApiType(eventAPIDefinition.EventAPISpec),
		Content: getEventSpec(eventAPIDefinition.EventAPISpec),
		Format:  clusterassetgroup.SpecFormat(getEventSpecFormat(eventAPIDefinition.EventAPISpec)),
	}
}

func createAssetFromAPIDefinition(apiDefinition model.APIDefinition) clusterassetgroup.Asset {

	return clusterassetgroup.Asset{
		ID:      apiDefinition.ID,
		Name:    apiDefinition.Name,
		Type:    getApiType(apiDefinition.APISpec),
		Content: getSpec(apiDefinition.APISpec),
		Format:  getSpecFormat(apiDefinition.APISpec),
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
