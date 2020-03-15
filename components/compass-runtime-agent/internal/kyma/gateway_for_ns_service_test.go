package kyma

import (
	"testing"

	"github.com/stretchr/testify/require"

	"kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/rafter/clusterassetgroup"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kyma-project.io/compass-runtime-agent/internal/apperrors"
	rafterMocks "kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/rafter/mocks"
	"kyma-project.io/compass-runtime-agent/internal/kyma/applications/converters"
	appMocks "kyma-project.io/compass-runtime-agent/internal/kyma/applications/mocks"
	"kyma-project.io/compass-runtime-agent/internal/kyma/model"
)

func TestGatewayForNamespaceService(t *testing.T) {

	t.Run("should return error in case failed to determine differences between current and desired runtime state", func(t *testing.T) {
		// given
		applicationsManagerMock := &appMocks.Repository{}
		converterMock := &appMocks.Converter{}
		rafterServiceMock := &rafterMocks.Service{}

		applicationsManagerMock.On("List", metav1.ListOptions{}).Return(nil, apperrors.Internal("some error"))

		directorApplication := getTestDirectorApplication("id1", "name1", []model.APIDefinition{}, []model.EventAPIDefinition{})

		directorApplications := []model.Application{
			directorApplication,
		}

		// when
		kymaService := NewGatewayForNsService(applicationsManagerMock, converterMock, rafterServiceMock)
		_, err := kymaService.Apply(directorApplications)

		// then
		assert.Error(t, err)
		converterMock.AssertExpectations(t)
		applicationsManagerMock.AssertExpectations(t)
		rafterServiceMock.AssertExpectations(t)

	})

	t.Run("should apply Create operation", func(t *testing.T) {
		// given
		applicationsManagerMock := &appMocks.Repository{}
		converterMock := &appMocks.Converter{}
		rafterServiceMock := &rafterMocks.Service{}

		api := fixDirectorAPiDefinition("API1", "name", fixAPISpec(), nil)
		eventAPI := fixDirectorEventAPIDefinition("EventAPI1", "name", fixEventAPISpec())

		apiPackage1 := fixAPIPackage("package1", []model.APIDefinition{api}, nil)
		apiPackage2 := fixAPIPackage("package2", nil, []model.EventAPIDefinition{eventAPI})
		apiPackage3 := fixAPIPackage("package3", []model.APIDefinition{api}, []model.EventAPIDefinition{eventAPI})
		directorApplication := fixDirectorApplication("id1", "name1", apiPackage1, apiPackage2, apiPackage3)

		entry1 := fixAPIEntry("API1", "api1")
		entry2 := fixEventAPIEntry("EventAPI1", "eventapi1")

		newRuntimeService1 := fixService("package1", entry1)
		newRuntimeService2 := fixService("package2", entry2)
		newRuntimeService3 := fixService("package3", entry1, entry2)

		newRuntimeApplication := getTestApplication("name1", "id1", []v1alpha1.Service{newRuntimeService1, newRuntimeService2, newRuntimeService3})

		directorApplications := []model.Application{
			directorApplication,
		}

		existingRuntimeApplications := v1alpha1.ApplicationList{
			Items: []v1alpha1.Application{},
		}

		converterMock.On("Do", directorApplication).Return(newRuntimeApplication)
		applicationsManagerMock.On("Create", &newRuntimeApplication).Return(&newRuntimeApplication, nil)
		applicationsManagerMock.On("List", metav1.ListOptions{}).Return(&existingRuntimeApplications, nil)

		asset1 := fixAPIAsset("API1", "name")
		asset2 := fixEventAPIAsset("EventAPI1", "name")

		expectedApiAssets1 := []clusterassetgroup.Asset{asset1}
		expectedApiAssets2 := []clusterassetgroup.Asset{asset2}
		expectedApiAssets3 := []clusterassetgroup.Asset{asset1, asset2}

		rafterServiceMock.On("Put", "package1", expectedApiAssets1).Return(nil)
		rafterServiceMock.On("Put", "package2", expectedApiAssets2).Return(nil)
		rafterServiceMock.On("Put", "package3", expectedApiAssets3).Return(nil)

		expectedResult := []Result{
			{
				ApplicationName: "name1",
				ApplicationID:   "id1",
				Operation:       Create,
				Error:           nil,
			},
		}

		// when
		kymaService := NewGatewayForNsService(applicationsManagerMock, converterMock, rafterServiceMock)
		result, err := kymaService.Apply(directorApplications)

		// then
		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)
		converterMock.AssertExpectations(t)
		applicationsManagerMock.AssertExpectations(t)
		rafterServiceMock.AssertExpectations(t)
	})

	t.Run("should apply Update operation", func(t *testing.T) {
		// given
		applicationsManagerMock := &appMocks.Repository{}
		converterMock := &appMocks.Converter{}
		rafterServiceMock := &rafterMocks.Service{}

		api1 := fixDirectorAPiDefinition("API1", "Name", fixAPISpec(), nil)
		eventAPI1 := fixDirectorEventAPIDefinition("EventAPI1", "Name", fixEventAPISpec())
		apiPackage1 := fixAPIPackage("package1", []model.APIDefinition{api1}, []model.EventAPIDefinition{eventAPI1})

		api2 := fixDirectorAPiDefinition("API2", "Name", fixAPISpec(), nil)
		eventAPI2 := fixDirectorEventAPIDefinition("EventAPI2", "Name", fixEventAPISpec())
		apiPackage2 := fixAPIPackage("package2", []model.APIDefinition{api2}, []model.EventAPIDefinition{eventAPI2})

		api3 := fixDirectorAPiDefinition("API3", "Name", nil, nil)
		eventAPI3 := fixDirectorEventAPIDefinition("EventAPI2", "Name", nil)
		apiPackage3 := fixAPIPackage("package3", []model.APIDefinition{api3}, []model.EventAPIDefinition{eventAPI3})

		directorApplication := fixDirectorApplication("id1", "name1", apiPackage1, apiPackage2, apiPackage3)

		runtimeServiceToCreate := fixService("package1", fixServiceAPIEntry("API1"), fixEventAPIEntry("EventAPI1", "EventAPI1Name"))
		runtimeServiceToUpdate1 := fixService("package2", fixServiceAPIEntry("API2"), fixEventAPIEntry("EventAPI2", "EventAPI2Name"))
		runtimeServiceToUpdate2 := fixService("package3", fixServiceAPIEntry("API3"), fixEventAPIEntry("EventAPI3", "EventAPI3Name"))
		runtimeServiceToDelete := fixService("package4", fixServiceAPIEntry("API4"), fixEventAPIEntry("EventAPI4", "EventAPI4Name"))

		newRuntimeApplication := getTestApplication("name1", "id1", []v1alpha1.Service{runtimeServiceToCreate, runtimeServiceToUpdate1, runtimeServiceToUpdate2})

		directorApplications := []model.Application{
			directorApplication,
		}

		existingRuntimeApplication := getTestApplication("name1", "id1", []v1alpha1.Service{runtimeServiceToUpdate1, runtimeServiceToUpdate2, runtimeServiceToDelete})
		existingRuntimeApplications := v1alpha1.ApplicationList{
			Items: []v1alpha1.Application{existingRuntimeApplication},
		}

		apiAssets1 := []clusterassetgroup.Asset{
			fixAPIAsset("API1", "Name"),
			fixEventAPIAsset("EventAPI1", "Name"),
		}

		apiAssets2 := []clusterassetgroup.Asset{
			fixAPIAsset("API2", "Name"),
			fixEventAPIAsset("EventAPI2", "Name"),
		}

		converterMock.On("Do", directorApplication).Return(newRuntimeApplication)
		applicationsManagerMock.On("Update", &newRuntimeApplication).Return(&newRuntimeApplication, nil)
		applicationsManagerMock.On("List", metav1.ListOptions{}).Return(&existingRuntimeApplications, nil)

		rafterServiceMock.On("Put", "package1", apiAssets1).Return(nil)
		rafterServiceMock.On("Put", "package2", apiAssets2).Return(nil)
		rafterServiceMock.On("Delete", "package3").Return(nil)
		rafterServiceMock.On("Delete", "package4").Return(nil)

		expectedResult := []Result{
			{
				ApplicationName: "name1",
				ApplicationID:   "id1",
				Operation:       Update,
				Error:           nil,
			},
		}

		// when
		kymaService := NewGatewayForNsService(applicationsManagerMock, converterMock, rafterServiceMock)
		result, err := kymaService.Apply(directorApplications)

		// then
		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)
		converterMock.AssertExpectations(t)
		applicationsManagerMock.AssertExpectations(t)
		rafterServiceMock.AssertExpectations(t)
	})

	t.Run("should apply Delete operation", func(t *testing.T) {
		// given
		applicationsManagerMock := &appMocks.Repository{}
		converterMock := &appMocks.Converter{}
		rafterServiceMock := &rafterMocks.Service{}

		runtimeServiceToDelete := fixService("package1", fixServiceAPIEntry("API1"), fixEventAPIEntry("EventAPI1", "EventAPI1Name"))
		runtimeApplicationToDelete := getTestApplication("name1", "id1", []v1alpha1.Service{runtimeServiceToDelete})

		existingRuntimeApplications := v1alpha1.ApplicationList{
			Items: []v1alpha1.Application{
				runtimeApplicationToDelete,
			},
		}

		applicationsManagerMock.On("Delete", runtimeApplicationToDelete.Name, &metav1.DeleteOptions{}).Return(nil)
		applicationsManagerMock.On("List", metav1.ListOptions{}).Return(&existingRuntimeApplications, nil)
		rafterServiceMock.On("Delete", "package1").Return(nil)

		expectedResult := []Result{
			{
				ApplicationName: "name1",
				ApplicationID:   "",
				Operation:       Delete,
				Error:           nil,
			},
		}

		// when
		kymaService := NewGatewayForNsService(applicationsManagerMock, converterMock, rafterServiceMock)
		result, err := kymaService.Apply([]model.Application{})

		// then
		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)
		converterMock.AssertExpectations(t)
		applicationsManagerMock.AssertExpectations(t)
		rafterServiceMock.AssertExpectations(t)
	})

	t.Run("should manage only Applications with CompassMetadata in the Spec", func(t *testing.T) {
		// given
		applicationsManagerMock := &appMocks.Repository{}
		converterMock := &appMocks.Converter{}
		rafterServiceMock := &rafterMocks.Service{}

		runtimeServiceToDelete := fixService("package1", fixServiceAPIEntry("API1"), fixEventAPIEntry("EventAPI1", "EventAPI1Name"))
		notManagedRuntimeService := fixService("package2", fixServiceAPIEntry("API2"), fixEventAPIEntry("EventAPI2", "EventAPI2Name"))

		runtimeApplicationToDelete := getTestApplication("name1", "id1", []v1alpha1.Service{runtimeServiceToDelete})
		notManagedRuntimeApplication := getTestApplicationNotManagedByCompass("id2", []v1alpha1.Service{notManagedRuntimeService})

		existingRuntimeApplications := v1alpha1.ApplicationList{
			Items: []v1alpha1.Application{
				runtimeApplicationToDelete,
				notManagedRuntimeApplication,
			},
		}

		applicationsManagerMock.On("Delete", runtimeApplicationToDelete.Name, &metav1.DeleteOptions{}).Return(nil)
		applicationsManagerMock.On("List", metav1.ListOptions{}).Return(&existingRuntimeApplications, nil)
		rafterServiceMock.On("Delete", "package1").Return(nil)

		expectedResult := []Result{
			{
				ApplicationName: "name1",
				ApplicationID:   "",
				Operation:       Delete,
				Error:           nil,
			},
		}

		// when
		kymaService := NewGatewayForNsService(applicationsManagerMock, converterMock, rafterServiceMock)
		result, err := kymaService.Apply([]model.Application{})

		// then
		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)
		converterMock.AssertExpectations(t)
		applicationsManagerMock.AssertExpectations(t)
		rafterServiceMock.AssertExpectations(t)
	})

	t.Run("should not break execution when error occurred when applying Application CR", func(t *testing.T) {
		// given
		applicationsManagerMock := &appMocks.Repository{}
		converterMock := &appMocks.Converter{}
		rafterServiceMock := &rafterMocks.Service{}

		newRuntimeService1 := fixService("package1", fixServiceAPIEntry("API1"), fixEventAPIEntry("EventAPI1", "EventAPI1Name"))
		newRuntimeService2 := fixService("package2", fixServiceAPIEntry("API2"), fixEventAPIEntry("EventAPI2", "EventAPI2Name"))

		existingRuntimeService1 := fixService("package3", fixServiceAPIEntry("API3"), fixEventAPIEntry("EventAPI3", "EventAPI1Name"))
		existingRuntimeService2 := fixService("package4", fixServiceAPIEntry("API4"), fixEventAPIEntry("EventAPI4", "EventAPI2Name"))

		runtimeServiceToBeDeleted1 := v1alpha1.Service{
			ID: "package5",
			Entries: []v1alpha1.Entry{
				fixServiceAPIEntry("API1"),
				fixServiceEventAPIEntry("EventAPI1"),
			},
		}

		api := fixDirectorAPiDefinition("API1", "name", fixAPISpec(), nil)
		eventAPI := fixDirectorEventAPIDefinition("EventAPI1", "name", fixEventAPISpec())

		apiPackage1 := fixAPIPackage("package1", []model.APIDefinition{api}, nil)
		apiPackage2 := fixAPIPackage("package2", nil, []model.EventAPIDefinition{eventAPI})
		newDirectorApplication := fixDirectorApplication("id1", "name1", apiPackage1, apiPackage2)

		newRuntimeApplication1 := getTestApplication("name1", "id1", []v1alpha1.Service{newRuntimeService1, newRuntimeService2})

		apiPackage3 := fixAPIPackage("package3", []model.APIDefinition{api}, []model.EventAPIDefinition{eventAPI})

		existingDirectorApplication := fixDirectorApplication("id2", "name2", apiPackage3)
		newRuntimeApplication2 := getTestApplication("name2", "id2", []v1alpha1.Service{newRuntimeService1, newRuntimeService2, existingRuntimeService1, existingRuntimeService2})

		runtimeApplicationToBeDeleted := getTestApplication("name3", "id3", []v1alpha1.Service{runtimeServiceToBeDeleted1})

		directorApplications := []model.Application{
			newDirectorApplication,
			existingDirectorApplication,
		}

		existingRuntimeApplication := getTestApplication("name2", "id2", []v1alpha1.Service{existingRuntimeService1, existingRuntimeService2, runtimeServiceToBeDeleted1})

		existingRuntimeApplications := v1alpha1.ApplicationList{
			Items: []v1alpha1.Application{
				existingRuntimeApplication,
				runtimeApplicationToBeDeleted,
			},
		}

		converterMock.On("Do", newDirectorApplication).Return(newRuntimeApplication1)
		converterMock.On("Do", existingDirectorApplication).Return(newRuntimeApplication2)
		applicationsManagerMock.On("Create", &newRuntimeApplication1).Return(nil, apperrors.Internal("some error"))
		applicationsManagerMock.On("Update", &newRuntimeApplication2).Return(nil, apperrors.Internal("some error"))
		applicationsManagerMock.On("Delete", runtimeApplicationToBeDeleted.Name, &metav1.DeleteOptions{}).Return(apperrors.Internal("some error"))
		applicationsManagerMock.On("List", metav1.ListOptions{}).Return(&existingRuntimeApplications, nil)

		rafterServiceMock.On("Delete", "package5").Return(apperrors.Internal("some error"))

		// when
		kymaService := NewGatewayForNsService(applicationsManagerMock, converterMock, rafterServiceMock)
		result, err := kymaService.Apply(directorApplications)

		// then
		require.NoError(t, err)
		require.Equal(t, 3, len(result))
		assert.NotNil(t, result[0].Error)
		assert.NotNil(t, result[1].Error)
		assert.NotNil(t, result[2].Error)
		converterMock.AssertExpectations(t)
		applicationsManagerMock.AssertExpectations(t)
		rafterServiceMock.AssertExpectations(t)
	})
}

func fixDirectorApplication(id, name string, apiPackages ...model.APIPackage) model.Application {
	return model.Application{
		ID:          id,
		Name:        name,
		APIPackages: apiPackages,
	}
}

func fixAPIPackage(id string, apiDefinitions []model.APIDefinition, eventAPIDefinitions []model.EventAPIDefinition) model.APIPackage {
	return model.APIPackage{
		ID:               id,
		APIDefinitions:   apiDefinitions,
		EventDefinitions: eventAPIDefinitions,
	}
}

func fixAPIEntry(id, name string) v1alpha1.Entry {
	return v1alpha1.Entry{
		ID:        id,
		Name:      name,
		Type:      converters.SpecAPIType,
		TargetUrl: "www.example.com/1",
	}
}

func fixEventAPIEntry(id, name string) v1alpha1.Entry {
	return v1alpha1.Entry{
		ID:   id,
		Name: name,
		Type: converters.SpecEventsType,
	}
}

func fixAPISpec() *model.APISpec {
	return &model.APISpec{
		Data:   []byte("spec"),
		Type:   model.APISpecTypeOpenAPI,
		Format: model.SpecFormatJSON,
	}
}

func fixEventAPISpec() *model.EventAPISpec {
	return &model.EventAPISpec{
		Data:   []byte("spec"),
		Type:   model.EventAPISpecTypeAsyncAPI,
		Format: model.SpecFormatJSON,
	}
}

func fixServiceAPIEntry(id string) v1alpha1.Entry {
	return v1alpha1.Entry{
		ID:        id,
		Name:      "Name",
		Type:      converters.SpecAPIType,
		TargetUrl: "www.example.com/1",
	}
}

func fixServiceEventAPIEntry(id string) v1alpha1.Entry {
	return v1alpha1.Entry{
		ID:   id,
		Name: "Name",
		Type: converters.SpecEventsType,
	}
}

func fixAPIAsset(id, name string) clusterassetgroup.Asset {
	return clusterassetgroup.Asset{
		ID:      id,
		Name:    name,
		Type:    clusterassetgroup.OpenApiType,
		Format:  clusterassetgroup.SpecFormatJSON,
		Content: []byte("spec"),
	}
}

func fixEventAPIAsset(id, name string) clusterassetgroup.Asset {
	return clusterassetgroup.Asset{
		ID:      id,
		Name:    name,
		Type:    clusterassetgroup.AsyncApi,
		Format:  clusterassetgroup.SpecFormatJSON,
		Content: []byte("spec"),
	}
}

func fixService(serviceID string, entries ...v1alpha1.Entry) v1alpha1.Service {
	return v1alpha1.Service{
		ID:      serviceID,
		Entries: entries,
	}
}
