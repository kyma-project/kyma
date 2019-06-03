package application

import (
	"testing"

	mappingTypes "github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
)

func TestApplicationMappingConverter_transformApplicationMappingServiceToGQL(t *testing.T) {
	// Given
	amc := applicationMappingConverter{}
	fixAms := []mappingTypes.ApplicationMappingService{
		{
			ID: "63410f90-6401-4123-878b-3ddf1112dc06",
		},
		{
			ID: "d2ea2944-0d53-4035-8dce-02d929e4665e",
		},
	}

	// When
	result := amc.transformApplicationMappingServiceToGQL(fixAms)

	// Then
	assert.Len(t, result, 2)
	assert.Contains(t, result, &gqlschema.ApplicationMappingService{ID: "63410f90-6401-4123-878b-3ddf1112dc06"})
	assert.Contains(t, result, &gqlschema.ApplicationMappingService{ID: "d2ea2944-0d53-4035-8dce-02d929e4665e"})
}

func TestApplicationMappingConverter_transformApplicationMappingServiceFromGQL(t *testing.T) {
	// Given
	amc := applicationMappingConverter{}
	fixAms := []*gqlschema.ApplicationMappingService{
		{
			ID: "aa943f0d-d69b-4855-9adb-66d8f3373c33",
		},
		{
			ID: "3f1fd883-a630-43bf-847d-8bf6585bb3c5",
		},
	}

	// When
	result := amc.transformApplicationMappingServiceFromGQL(fixAms)

	// Then
	assert.Len(t, result, 2)
	assert.Contains(t, result, mappingTypes.ApplicationMappingService{ID: "3f1fd883-a630-43bf-847d-8bf6585bb3c5"})
	assert.Contains(t, result, mappingTypes.ApplicationMappingService{ID: "aa943f0d-d69b-4855-9adb-66d8f3373c33"})
}
