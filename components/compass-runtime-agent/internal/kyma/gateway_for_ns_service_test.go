package kyma

import (
	"testing"

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

	//t.Run("should apply Create operation", func(t *testing.T) {
	//	// given
	//	applicationsManagerMock := &appMocks.Repository{}
	//	converterMock := &appMocks.Converter{}
	//	rafterServiceMock := &rafterMocks.Service{}
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
	//	eventAPI := getTestDirectorEventAPIDefinition("EventAPI1", "name", nil)
	//
	//	apiPackage1 := createAPIPackage("package1", []model.APIDefinition{api}, nil)
	//	apiPackage2 := createAPIPackage("package2", nil, []model.EventAPIDefinition{eventAPI})
	//	directorApplication := createTestApplication("id1", "name1", []model.APIPackage{apiPackage1, apiPackage2})
	//
	//	entry1 := getTestAPIEntry("api1")
	//	entry2 := getTestEventAPIEntry("eventapi1")
	//
	//	runtimeService1 := createService("API1", []v1alpha1.Entry{entry1})
	//	runtimeService2 := createService("EventAPI1", []v1alpha1.Entry{entry2})
	//
	//	runtimeApplication := getTestApplication("name1", "id1", []v1alpha1.Service{runtimeService1, runtimeService2})
	//
	//	directorApplications := []model.Application{
	//		directorApplication,
	//	}
	//
	//	existingRuntimeApplications := v1alpha1.ApplicationList{
	//		Items: []v1alpha1.Application{},
	//	}
	//
	//	converterMock.On("Do", directorApplication).Return(runtimeApplication)
	//	applicationsManagerMock.On("Create", &runtimeApplication).Return(&runtimeApplication, nil)
	//	applicationsManagerMock.On("List", metav1.ListOptions{}).Return(&existingRuntimeApplications, nil)
	//
	//	apiAssets := []clusterassetgroup.Asset{
	//		{
	//			Name:    "name",
	//			Type:    clusterassetgroup.OpenApiType,
	//			Format:  clusterassetgroup.SpecFormatJSON,
	//			Content: []byte("spec"),
	//		},
	//	}
	//
	//	rafterServiceMock.On("Put", "package1", apiAssets).Return(nil)
	//
	//	expectedResult := []Result{
	//		{
	//			ApplicationName: "name1",
	//			ApplicationID:   "id1",
	//			Operation:       Create,
	//			Error:           nil,
	//		},
	//	}
	//
	//	// when
	//	kymaService := NewGatewayForNsService(applicationsManagerMock, converterMock, rafterServiceMock)
	//	result, err := kymaService.Apply(directorApplications)
	//
	//	// then
	//	assert.NoError(t, err)
	//	assert.Equal(t, expectedResult, result)
	//	converterMock.AssertExpectations(t)
	//	applicationsManagerMock.AssertExpectations(t)
	//	rafterServiceMock.AssertExpectations(t)
	//})

	//
	//t.Run("should apply Update operation", func(t *testing.T) {
	//	// given
	//	applicationsManagerMock := &appMocks.Repository{}
	//	converterMock := &appMocks.Converter{}
	//	resourcesServiceMocks := &resourcesServiceMocks.Service{}
	//
	//	api := getTestDirectorAPiDefinition("API1", "Name", nil, nil)
	//	eventAPI := getTestDirectorEventAPIDefinition("EventAPI1", "name", &model.EventAPISpec{
	//		Data:   []byte("spec"),
	//		Type:   model.EventAPISpecTypeAsyncAPI,
	//		Format: model.SpecFormatJSON,
	//	})
	//
	//	directorApplication := getTestDirectorApplication("id1", "name1", []model.APIDefinition{api}, []model.EventAPIDefinition{eventAPI})
	//
	//	runtimeService1 := getTestServiceWithCredentials("API1")
	//	runtimeService2 := getTestServiceWithoutCredentials("EventAPI1")
	//	runtimeService3 := getTestServiceWithoutCredentials("API2")
	//
	//	runtimeApplication := getTestApplication("name1", "id1", []v1alpha1.Service{runtimeService1, runtimeService2})
	//
	//	directorApplications := []model.Application{
	//		directorApplication,
	//	}
	//
	//	existingRuntimeApplication := getTestApplication("name1", "id1", []v1alpha1.Service{runtimeService1, runtimeService3})
	//	existingRuntimeApplications := v1alpha1.ApplicationList{
	//		Items: []v1alpha1.Application{existingRuntimeApplication},
	//	}
	//
	//	apiAssets := []clusterassetgroup.Asset{
	//		{
	//			Name:    "name",
	//			Type:    clusterassetgroup.AsyncApi,
	//			Format:  clusterassetgroup.SpecFormatJSON,
	//			Content: []byte("spec"),
	//		},
	//	}
	//
	//	converterMock.On("Do", directorApplication).Return(runtimeApplication)
	//	applicationsManagerMock.On("Update", &runtimeApplication).Return(&runtimeApplication, nil)
	//	applicationsManagerMock.On("List", metav1.ListOptions{}).Return(&existingRuntimeApplications, nil)
	//	resourcesServiceMocks.On("UpdateApiResources", "name1", types.UID(""), "API1", mock.MatchedBy(getCredentialsMatcher(api.Credentials)), []clusterassetgroup.Asset(nil)).Return(nil)
	//	resourcesServiceMocks.On("CreateEventApiResources", "name1", "EventAPI1", apiAssets).Return(nil)
	//	resourcesServiceMocks.On("DeleteApiResources", "name1", "API2", "").Return(nil)
	//
	//	expectedResult := []Result{
	//		{
	//			ApplicationName: "name1",
	//			ApplicationID:   "id1",
	//			Operation:       Update,
	//			Error:           nil,
	//		},
	//	}
	//
	//	// when
	//	kymaService := NewGatewayForAppService(applicationsManagerMock, converterMock, resourcesServiceMocks)
	//	result, err := kymaService.Apply(directorApplications)
	//
	//	// then
	//	assert.NoError(t, err)
	//	assert.Equal(t, expectedResult, result)
	//	converterMock.AssertExpectations(t)
	//	applicationsManagerMock.AssertExpectations(t)
	//	resourcesServiceMocks.AssertExpectations(t)
	//})
	//
	//t.Run("should apply Delete operation", func(t *testing.T) {
	//	// given
	//	applicationsManagerMock := &appMocks.Repository{}
	//	converterMock := &appMocks.Converter{}
	//	resourcesServiceMocks := &resourcesServiceMocks.Service{}
	//
	//	runtimeService := getTestServiceWithCredentials("API1")
	//	runtimeApplication := getTestApplication("name1", "id1", []v1alpha1.Service{runtimeService})
	//
	//	existingRuntimeApplications := v1alpha1.ApplicationList{
	//		Items: []v1alpha1.Application{
	//			runtimeApplication,
	//		},
	//	}
	//
	//	applicationsManagerMock.On("Delete", runtimeApplication.Name, &metav1.DeleteOptions{}).Return(nil)
	//	applicationsManagerMock.On("List", metav1.ListOptions{}).Return(&existingRuntimeApplications, nil)
	//	resourcesServiceMocks.On("DeleteApiResources", "name1", "API1", "credentialsSecretName1").Return(nil)
	//
	//	expectedResult := []Result{
	//		{
	//			ApplicationName: "name1",
	//			ApplicationID:   "",
	//			Operation:       Delete,
	//			Error:           nil,
	//		},
	//	}
	//
	//	// when
	//	kymaService := NewGatewayForAppService(applicationsManagerMock, converterMock, resourcesServiceMocks)
	//	result, err := kymaService.Apply([]model.Application{})
	//
	//	// then
	//	assert.NoError(t, err)
	//	assert.Equal(t, expectedResult, result)
	//	converterMock.AssertExpectations(t)
	//	applicationsManagerMock.AssertExpectations(t)
	//	resourcesServiceMocks.AssertExpectations(t)
	//})
	//
	//t.Run("should manage only Applications with CompassMetadata in the Spec", func(t *testing.T) {
	//	// given
	//	applicationsManagerMock := &appMocks.Repository{}
	//	converterMock := &appMocks.Converter{}
	//	resourcesServiceMocks := &resourcesServiceMocks.Service{}
	//
	//	runtimeService1 := getTestServiceWithCredentials("API1")
	//	runtimeService2 := getTestServiceWithCredentials("API2")
	//	runtimeApplication := getTestApplication("name1", "id1", []v1alpha1.Service{runtimeService1})
	//	notManagedRuntimeApplication := getTestApplicationNotManagedByCompass("id2", []v1alpha1.Service{runtimeService2})
	//
	//	existingRuntimeApplications := v1alpha1.ApplicationList{
	//		Items: []v1alpha1.Application{
	//			runtimeApplication,
	//			notManagedRuntimeApplication,
	//		},
	//	}
	//
	//	applicationsManagerMock.On("Delete", mock.AnythingOfType("string"), &metav1.DeleteOptions{}).Return(nil)
	//	applicationsManagerMock.On("List", metav1.ListOptions{}).Return(&existingRuntimeApplications, nil)
	//	resourcesServiceMocks.On("DeleteApiResources", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)
	//
	//	expectedResult := []Result{
	//		{
	//			ApplicationName: "name1",
	//			ApplicationID:   "",
	//			Operation:       Delete,
	//			Error:           nil,
	//		},
	//	}
	//
	//	// when
	//	kymaService := NewGatewayForAppService(applicationsManagerMock, converterMock, resourcesServiceMocks)
	//	result, err := kymaService.Apply([]model.Application{})
	//
	//	// then
	//	assert.NoError(t, err)
	//	assert.Equal(t, expectedResult, result)
	//	converterMock.AssertExpectations(t)
	//	applicationsManagerMock.AssertExpectations(t)
	//	resourcesServiceMocks.AssertExpectations(t)
	//})
	//
	//t.Run("should not break execution when error occurred when applying Application CR", func(t *testing.T) {
	//	// given
	//	applicationsManagerMock := &appMocks.Repository{}
	//	converterMock := &appMocks.Converter{}
	//	resourcesServiceMocks := &resourcesServiceMocks.Service{}
	//
	//	newRuntimeService1 := getTestServiceWithCredentials("API1")
	//	newRuntimeService2 := getTestServiceWithCredentials("EventAPI1")
	//
	//	existingRuntimeService1 := getTestServiceWithCredentials("API2")
	//	existingRuntimeService2 := getTestServiceWithCredentials("EventAPI2")
	//
	//	runtimeServiceToBeDeleted1 := getTestServiceWithCredentials("API3")
	//	runtimeServiceToBeDeleted2 := getTestServiceWithCredentials("EventAPI3")
	//
	//	newDirectorApi := getTestDirectorAPiDefinition("API1", "Name", nil, nil)
	//	newDirectorEventApi := getTestDirectorEventAPIDefinition("EventAPI1", "name", nil)
	//
	//	newDirectorApplication := getTestDirectorApplication("id1", "name1",
	//		[]model.APIDefinition{newDirectorApi}, []model.EventAPIDefinition{newDirectorEventApi})
	//	convertedNewRuntimeApplication := getTestApplication("name1", "id1", []v1alpha1.Service{newRuntimeService1, newRuntimeService2})
	//
	//	existingDirectorApi := getTestDirectorAPiDefinition("API2", "Name", nil, nil)
	//	existingDirectorEventApi := getTestDirectorEventAPIDefinition("EventAPI2", "name", nil)
	//
	//	existingDirectorApplication := getTestDirectorApplication("id2", "name2", []model.APIDefinition{newDirectorApi, existingDirectorApi}, []model.EventAPIDefinition{newDirectorEventApi, existingDirectorEventApi})
	//	convertedExistingRuntimeApplication := getTestApplication("name2", "id2", []v1alpha1.Service{newRuntimeService1, newRuntimeService2, existingRuntimeService1, existingRuntimeService2})
	//
	//	runtimeApplicationToBeDeleted := getTestApplication("name3", "id3", []v1alpha1.Service{runtimeServiceToBeDeleted1, runtimeServiceToBeDeleted2})
	//
	//	directorApplications := []model.Application{
	//		newDirectorApplication,
	//		existingDirectorApplication,
	//	}
	//
	//	existingRuntimeApplication := getTestApplication("name2", "id2", []v1alpha1.Service{existingRuntimeService1, existingRuntimeService2, runtimeServiceToBeDeleted1, runtimeServiceToBeDeleted2})
	//
	//	existingRuntimeApplications := v1alpha1.ApplicationList{
	//		Items: []v1alpha1.Application{existingRuntimeApplication,
	//			runtimeApplicationToBeDeleted},
	//	}
	//
	//	converterMock.On("Do", newDirectorApplication).Return(convertedNewRuntimeApplication)
	//	converterMock.On("Do", existingDirectorApplication).Return(convertedExistingRuntimeApplication)
	//	applicationsManagerMock.On("Create", &convertedNewRuntimeApplication).Return(nil, apperrors.Internal("some error"))
	//	applicationsManagerMock.On("Update", &convertedExistingRuntimeApplication).Return(nil, apperrors.Internal("some error"))
	//	applicationsManagerMock.On("Delete", runtimeApplicationToBeDeleted.Name, &metav1.DeleteOptions{}).Return(apperrors.Internal("some error"))
	//	applicationsManagerMock.On("List", metav1.ListOptions{}).Return(&existingRuntimeApplications, nil)
	//
	//	resourcesServiceMocks.On("DeleteApiResources", "name3", "API3", "credentialsSecretName1").Return(apperrors.Internal("some error"))
	//	resourcesServiceMocks.On("DeleteApiResources", "name3", "EventAPI3", "credentialsSecretName1").Return(apperrors.Internal("some error"))
	//
	//	// when
	//	kymaService := NewGatewayForAppService(applicationsManagerMock, converterMock, resourcesServiceMocks)
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
	//	resourcesServiceMocks.AssertNotCalled(t, "CreateApiResources")
	//	resourcesServiceMocks.AssertExpectations(t)
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

func createService(serviceID string, entries []v1alpha1.Entry) v1alpha1.Service {
	return v1alpha1.Service{
		ID:      serviceID,
		Entries: entries,
	}
}
