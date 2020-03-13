package kyma

import (
	"testing"

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

		api := getTestDirectorAPiDefinition("API1", "name", getTestAPISpec(), nil)
		eventAPI := getTestDirectorEventAPIDefinition("EventAPI1", "name", getTestEventAPISpec())

		apiPackage1 := createAPIPackage("package1", []model.APIDefinition{api}, nil)
		apiPackage2 := createAPIPackage("package2", nil, []model.EventAPIDefinition{eventAPI})
		apiPackage3 := createAPIPackage("package3", []model.APIDefinition{api}, []model.EventAPIDefinition{eventAPI})
		directorApplication := createTestApplication("id1", "name1", []model.APIPackage{apiPackage1, apiPackage2, apiPackage3})

		entry1 := getTestAPIEntry("api1")
		entry2 := getTestEventAPIEntry("eventapi1")

		runtimeService1 := createService("package1", []v1alpha1.Entry{entry1})
		runtimeService2 := createService("package2", []v1alpha1.Entry{entry2})
		runtimeService3 := createService("package3", []v1alpha1.Entry{entry1, entry2})

		runtimeApplication := getTestApplication("name1", "id1", []v1alpha1.Service{runtimeService1, runtimeService2, runtimeService3})

		directorApplications := []model.Application{
			directorApplication,
		}

		existingRuntimeApplications := v1alpha1.ApplicationList{
			Items: []v1alpha1.Application{},
		}

		converterMock.On("Do", directorApplication).Return(runtimeApplication)
		applicationsManagerMock.On("Create", &runtimeApplication).Return(&runtimeApplication, nil)
		applicationsManagerMock.On("List", metav1.ListOptions{}).Return(&existingRuntimeApplications, nil)

		asset1 := getTestAPIAsset("name")

		apiAssets1 := []clusterassetgroup.Asset{asset1}

		asset2 := getTestEventAPIAsset("name")

		apiAssets2 := []clusterassetgroup.Asset{asset2}

		apiAssets3 := []clusterassetgroup.Asset{asset1, asset2}

		rafterServiceMock.On("Put", "package1", apiAssets1).Return(nil)
		rafterServiceMock.On("Put", "package2", apiAssets2).Return(nil)
		rafterServiceMock.On("Put", "package3", apiAssets3).Return(nil)

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

		api1 := getTestDirectorAPiDefinition("API1", "Name", getTestAPISpec(), nil)
		eventAPI1 := getTestDirectorEventAPIDefinition("EventAPI1", "Name", getTestEventAPISpec())
		apiPackage1 := createAPIPackage("package1", []model.APIDefinition{api1}, []model.EventAPIDefinition{eventAPI1})

		api2 := getTestDirectorAPiDefinition("API2", "Name", getTestAPISpec(), nil)
		eventAPI2 := getTestDirectorEventAPIDefinition("EventAPI2", "Name", getTestEventAPISpec())
		apiPackage2 := createAPIPackage("package2", []model.APIDefinition{api2}, []model.EventAPIDefinition{eventAPI2})

		api3 := getTestDirectorAPiDefinition("API3", "Name", nil, nil)
		eventAPI3 := getTestDirectorEventAPIDefinition("EventAPI2", "Name", nil)
		apiPackage3 := createAPIPackage("package3", []model.APIDefinition{api3}, []model.EventAPIDefinition{eventAPI3})

		directorApplication := createTestApplication("id1", "name1", []model.APIPackage{apiPackage1, apiPackage2, apiPackage3})

		runtimeService1 := v1alpha1.Service{
			ID: "package1",
			Entries: []v1alpha1.Entry{
				getTestServiceAPIEntry("API1"),
				getTestEventAPIEntry("EventAPI1"),
			},
		}

		runtimeService2 := v1alpha1.Service{
			ID: "package2",
			Entries: []v1alpha1.Entry{
				getTestServiceAPIEntry("API2"),
				getTestServiceEventAPIEntry("EventAPI2"),
			},
		}

		runtimeService3 := v1alpha1.Service{
			ID: "package3",
			Entries: []v1alpha1.Entry{
				getTestServiceAPIEntry("API3"),
				getTestServiceEventAPIEntry("EventAPI3"),
			},
		}

		runtimeService4 := v1alpha1.Service{
			ID: "package4",
			Entries: []v1alpha1.Entry{
				getTestServiceAPIEntry("API4"),
				getTestServiceEventAPIEntry("EventAPI4"),
			},
		}

		runtimeApplication := getTestApplication("name1", "id1", []v1alpha1.Service{runtimeService1, runtimeService2, runtimeService3})

		directorApplications := []model.Application{
			directorApplication,
		}

		existingRuntimeApplication := getTestApplication("name1", "id1", []v1alpha1.Service{runtimeService2, runtimeService3, runtimeService4})
		existingRuntimeApplications := v1alpha1.ApplicationList{
			Items: []v1alpha1.Application{existingRuntimeApplication},
		}

		apiAssets1 := []clusterassetgroup.Asset{
			getTestAPIAsset("Name"),
			getTestEventAPIAsset("Name"),
		}

		apiAssets2 := []clusterassetgroup.Asset{
			getTestAPIAsset("Name"),
			getTestEventAPIAsset("Name"),
		}

		converterMock.On("Do", directorApplication).Return(runtimeApplication)
		applicationsManagerMock.On("Update", &runtimeApplication).Return(&runtimeApplication, nil)
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

		runtimeService := v1alpha1.Service{
			ID: "package1",
			Entries: []v1alpha1.Entry{
				getTestServiceAPIEntry("API1"),
				getTestServiceEventAPIEntry("EventAPI1"),
			},
		}
		runtimeApplication := getTestApplication("name1", "id1", []v1alpha1.Service{runtimeService})

		existingRuntimeApplications := v1alpha1.ApplicationList{
			Items: []v1alpha1.Application{
				runtimeApplication,
			},
		}

		applicationsManagerMock.On("Delete", runtimeApplication.Name, &metav1.DeleteOptions{}).Return(nil)
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

		runtimeService1 := v1alpha1.Service{
			ID: "package1",
			Entries: []v1alpha1.Entry{
				getTestServiceAPIEntry("API1"),
				getTestServiceEventAPIEntry("EventAPI1"),
			},
		}

		runtimeService2 := v1alpha1.Service{
			ID: "package2",
			Entries: []v1alpha1.Entry{
				getTestServiceAPIEntry("API2"),
				getTestServiceEventAPIEntry("EventAPI2"),
			},
		}

		runtimeApplication := getTestApplication("name1", "id1", []v1alpha1.Service{runtimeService1})
		notManagedRuntimeApplication := getTestApplicationNotManagedByCompass("id2", []v1alpha1.Service{runtimeService2})

		existingRuntimeApplications := v1alpha1.ApplicationList{
			Items: []v1alpha1.Application{
				runtimeApplication,
				notManagedRuntimeApplication,
			},
		}

		applicationsManagerMock.On("Delete", runtimeApplication.Name, &metav1.DeleteOptions{}).Return(nil)
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

	//t.Run("should not break execution when error occurred when applying Application CR", func(t *testing.T) {
	//	// given
	//	applicationsManagerMock := &appMocks.Repository{}
	//	converterMock := &appMocks.Converter{}
	//	rafterServiceMock := &rafterMocks.Service{}
	//
	//	newRuntimeService1 := v1alpha1.Service{
	//		ID: "package1",
	//		Entries: []v1alpha1.Entry{
	//			{
	//				ID:        "API1",
	//				Name:      "Name",
	//				Type:      converters.SpecAPIType,
	//				TargetUrl: "www.example.com/1",
	//			},
	//			{
	//				ID:   "EventAPI1",
	//				Name: "Name",
	//				Type: converters.SpecEventsType,
	//			},
	//		},
	//	}
	//
	//	newRuntimeService2 := v1alpha1.Service{
	//		ID: "package2",
	//		Entries: []v1alpha1.Entry{
	//			{
	//				ID:        "API2",
	//				Name:      "Name",
	//				Type:      converters.SpecAPIType,
	//				TargetUrl: "www.example.com/1",
	//			},
	//			{
	//				ID:   "EventAPI2",
	//				Name: "Name",
	//				Type: converters.SpecEventsType,
	//			},
	//		},
	//	}
	//
	//	existingRuntimeService1 := v1alpha1.Service{
	//		ID: "package3",
	//		Entries: []v1alpha1.Entry{
	//			{
	//				ID:        "API2",
	//				Name:      "Name",
	//				Type:      converters.SpecAPIType,
	//				TargetUrl: "www.example.com/1",
	//			},
	//			{
	//				ID:   "EventAPI2",
	//				Name: "Name",
	//				Type: converters.SpecEventsType,
	//			},
	//		},
	//	}
	//	existingRuntimeService2 := v1alpha1.Service{
	//		ID: "package4",
	//		Entries: []v1alpha1.Entry{
	//			{
	//				ID:        "API1",
	//				Name:      "Name",
	//				Type:      converters.SpecAPIType,
	//				TargetUrl: "www.example.com/1",
	//			},
	//			{
	//				ID:   "EventAPI1",
	//				Name: "Name",
	//				Type: converters.SpecEventsType,
	//			},
	//		},
	//	}
	//
	//	runtimeServiceToBeDeleted1 := v1alpha1.Service{
	//		ID: "package5",
	//		Entries: []v1alpha1.Entry{
	//			{
	//				ID:        "API",
	//				Name:      "Name",
	//				Type:      converters.SpecAPIType,
	//				TargetUrl: "www.example.com/1",
	//			},
	//			{
	//				ID:   "EventAPI",
	//				Name: "Name",
	//				Type: converters.SpecEventsType,
	//			},
	//		},
	//	}
	//
	//	api := getTestDirectorAPiDefinition(
	//		"API1",
	//		"name",
	//		&model.APISpec{
	//			Data:   []byte("spec"),
	//			Type:   model.APISpecTypeOpenAPI,
	//			Format: model.SpecFormatJSON,
	//		},
	//		&model.Credentials{
	//			Basic: &model.Basic{
	//				Username: "admin",
	//				Password: "nimda",
	//			},
	//		})
	//
	//	eventAPI := getTestDirectorEventAPIDefinition(
	//		"EventAPI1",
	//		"name",
	//		&model.EventAPISpec{
	//			Data:   []byte("spec"),
	//			Type:   model.EventAPISpecTypeAsyncAPI,
	//			Format: model.SpecFormatJSON,
	//		})
	//
	//	apiPackage1 := createAPIPackage("package1", []model.APIDefinition{api}, nil)
	//	apiPackage2 := createAPIPackage("package2", nil, []model.EventAPIDefinition{eventAPI})
	//	newDirectorApplication := createTestApplication("id1", "name1", []model.APIPackage{apiPackage1, apiPackage2})
	//
	//	convertedNewRuntimeApplication := getTestApplication("name1", "id1", []v1alpha1.Service{newRuntimeService1, newRuntimeService2})
	//
	//	apiPackage3 := createAPIPackage("package3", []model.APIDefinition{api}, []model.EventAPIDefinition{eventAPI})
	//
	//	existingDirectorApplication := createTestApplication("id2", "name2", []model.APIPackage{apiPackage3})
	//	convertedExistingRuntimeApplication := getTestApplication("name2", "id2", []v1alpha1.Service{newRuntimeService1, newRuntimeService2, existingRuntimeService1, existingRuntimeService2})
	//
	//	runtimeApplicationToBeDeleted := getTestApplication("name3", "id3", []v1alpha1.Service{runtimeServiceToBeDeleted1})
	//
	//	directorApplications := []model.Application{
	//		newDirectorApplication,
	//		existingDirectorApplication,
	//	}
	//
	//	existingRuntimeApplication := getTestApplication("name2", "id2", []v1alpha1.Service{existingRuntimeService1, existingRuntimeService2, runtimeServiceToBeDeleted1})
	//
	//	existingRuntimeApplications := v1alpha1.ApplicationList{
	//		Items: []v1alpha1.Application{existingRuntimeApplication,
	//			runtimeApplicationToBeDeleted},
	//	}
	//
	//	asset1 := clusterassetgroup.Asset{
	//		Name:    "name",
	//		Type:    clusterassetgroup.OpenApiType,
	//		Format:  clusterassetgroup.SpecFormatJSON,
	//		Content: []byte("spec"),
	//	}
	//
	//	asset2 := clusterassetgroup.Asset{
	//		Name:    "name",
	//		Type:    clusterassetgroup.AsyncApi,
	//		Format:  clusterassetgroup.SpecFormatJSON,
	//		Content: []byte("spec"),
	//	}
	//
	//	apiAssets1 := []clusterassetgroup.Asset{asset1, asset2}
	//
	//	converterMock.On("Do", newDirectorApplication).Return(convertedNewRuntimeApplication)
	//	converterMock.On("Do", existingDirectorApplication).Return(convertedExistingRuntimeApplication)
	//	applicationsManagerMock.On("Create", &convertedNewRuntimeApplication).Return(nil, apperrors.Internal("some error"))
	//	applicationsManagerMock.On("Update", &convertedExistingRuntimeApplication).Return(nil, apperrors.Internal("some error"))
	//	applicationsManagerMock.On("Delete", runtimeApplicationToBeDeleted.Name, &metav1.DeleteOptions{}).Return(apperrors.Internal("some error"))
	//	applicationsManagerMock.On("List", metav1.ListOptions{}).Return(&existingRuntimeApplications, nil)
	//	//
	//	rafterServiceMock.On("Put", "package1", apiAssets1).Return(apperrors.Internal("some error"))
	//	rafterServiceMock.On("Put", "package2", apiAssets1).Return(apperrors.Internal("some error"))
	//	rafterServiceMock.On("Delete", "package5").Return(apperrors.Internal("some error"))
	//	rafterServiceMock.On("Delete", "package4").Return(apperrors.Internal("some error"))
	//
	//	// when
	//	kymaService := NewGatewayForNsService(applicationsManagerMock, converterMock, rafterServiceMock)
	//	result, err := kymaService.Apply(directorApplications)
	//
	//	// then
	//	require.NoError(t, err)
	//	require.Equal(t, 3, len(result))
	//	assert.NotNil(t, result[0].Error)
	//	assert.NotNil(t, result[1].Error)
	//	assert.NotNil(t, result[2].Error)
	//	converterMock.AssertExpectations(t)
	//	applicationsManagerMock.AssertExpectations(t)
	//	//rafterServiceMock.AssertNotCalled(t, "CreateApiResources")
	//	rafterServiceMock.AssertExpectations(t)
	//})
}

func createTestApplication(id, name string, apiPackages []model.APIPackage) model.Application {
	return model.Application{
		ID:          id,
		Name:        name,
		APIPackages: apiPackages,
	}
}

func createAPIPackage(id string, apiDefinitions []model.APIDefinition, eventAPIDefinitions []model.EventAPIDefinition) model.APIPackage {
	return model.APIPackage{
		ID:               id,
		APIDefinitions:   apiDefinitions,
		EventDefinitions: eventAPIDefinitions,
	}
}

func getTestAPIEntry(name string) v1alpha1.Entry {
	return v1alpha1.Entry{
		Name:      name,
		Type:      converters.SpecAPIType,
		TargetUrl: "www.example.com/1",
	}
}

func getTestEventAPIEntry(name string) v1alpha1.Entry {
	return v1alpha1.Entry{
		Name: name,
		Type: converters.SpecEventsType,
	}
}

func getTestAPISpec() *model.APISpec {
	return &model.APISpec{
		Data:   []byte("spec"),
		Type:   model.APISpecTypeOpenAPI,
		Format: model.SpecFormatJSON,
	}
}

func getTestEventAPISpec() *model.EventAPISpec {
	return &model.EventAPISpec{
		Data:   []byte("spec"),
		Type:   model.EventAPISpecTypeAsyncAPI,
		Format: model.SpecFormatJSON,
	}
}

func getTestServiceAPIEntry(id string) v1alpha1.Entry {
	return v1alpha1.Entry{
		ID:        "id",
		Name:      "Name",
		Type:      converters.SpecAPIType,
		TargetUrl: "www.example.com/1",
	}
}

func getTestServiceEventAPIEntry(id string) v1alpha1.Entry {
	return v1alpha1.Entry{
		ID:   id,
		Name: "Name",
		Type: converters.SpecEventsType,
	}
}

func getTestAPIAsset(id string) clusterassetgroup.Asset {
	return clusterassetgroup.Asset{
		Name:    id,
		Type:    clusterassetgroup.OpenApiType,
		Format:  clusterassetgroup.SpecFormatJSON,
		Content: []byte("spec"),
	}
}

func getTestEventAPIAsset(id string) clusterassetgroup.Asset {
	return clusterassetgroup.Asset{
		Name:    id,
		Type:    clusterassetgroup.AsyncApi,
		Format:  clusterassetgroup.SpecFormatJSON,
		Content: []byte("spec"),
	}
}

func createService(serviceID string, entries []v1alpha1.Entry) v1alpha1.Service {
	return v1alpha1.Service{
		ID:      serviceID,
		Entries: entries,
	}
}
