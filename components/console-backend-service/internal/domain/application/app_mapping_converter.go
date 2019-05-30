package application

import (
	mappingTypes "github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
)

type applicationMappingConverter struct{}

func (amc *applicationMappingConverter) transformApplicationMappingServiceToGQL(services []mappingTypes.ApplicationMappingService) []*gqlschema.ApplicationMappingService {
	if services == nil {
		return nil
	}

	ms := []*gqlschema.ApplicationMappingService{}
	for _, service := range services {
		var ams gqlschema.ApplicationMappingService
		ams.ID = service.ID
		ms = append(ms, &ams)
	}

	return ms
}

func (amc *applicationMappingConverter) transformApplicationMappingServiceFromGQL(services []*gqlschema.ApplicationMappingService) []mappingTypes.ApplicationMappingService {
	if services == nil {
		return nil
	}

	ms := []mappingTypes.ApplicationMappingService{}
	for _, service := range services {
		var ams mappingTypes.ApplicationMappingService
		ams.ID = service.ID
		ms = append(ms, ams)
	}

	return ms
}
