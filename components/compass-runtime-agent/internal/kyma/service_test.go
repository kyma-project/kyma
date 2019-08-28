package kyma

import (
	"testing"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/apiresources/assetstore/docstopic"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/types"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/apperrors"
	resourcesServiceMocks "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/apiresources/mocks"
	secretsmodel "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/apiresources/secrets/model"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/applications"
	appMocks "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/applications/mocks"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/model"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestService(t *testing.T) {

	nilSpec := []byte(nil)

	t.Run("should return error in case failed to determine differences between current and desired runtime state", func(t *testing.T) {
		// given
		applicationsManagerMock := &appMocks.Repository{}
		converterMock := &appMocks.Converter{}
		resourcesServiceMocks := &resourcesServiceMocks.Service{}

		applicationsManagerMock.On("List", metav1.ListOptions{}).Return(nil, apperrors.Internal("some error"))

		directorApplication := getTestDirectorApplication("id1", []model.APIDefinition{}, []model.EventAPIDefinition{})

		directorApplications := []model.Application{
			directorApplication,
		}

		// when
		kymaService := NewService(applicationsManagerMock, converterMock, resourcesServiceMocks)
		_, err := kymaService.Apply(directorApplications)

		// then
		assert.Error(t, err)
		converterMock.AssertExpectations(t)
		applicationsManagerMock.AssertExpectations(t)
		resourcesServiceMocks.AssertExpectations(t)

	})

	t.Run("should apply Create operation", func(t *testing.T) {
		// given
		applicationsManagerMock := &appMocks.Repository{}
		converterMock := &appMocks.Converter{}
		resourcesServiceMocks := &resourcesServiceMocks.Service{}

		api := getTestDirectorAPiDefinition("API1", &model.APISpec{Data: []byte("spec"), Type: model.APISpecTypeOpenAPI}, &model.Credentials{
			Basic: &model.Basic{
				Username: "admin",
				Password: "nimda",
			},
		})

		eventAPI := getTestDirectorEventAPIDefinition("EventAPI1", nil)

		directorApplication := getTestDirectorApplication("id1", []model.APIDefinition{api}, []model.EventAPIDefinition{eventAPI})

		runtimeService1 := getTestServiceWithCredentials("API1")
		runtimeService2 := getTestServiceWithCredentials("EventAPI1")

		runtimeApplication := getTestApplication("id1", []v1alpha1.Service{runtimeService1, runtimeService2})

		directorApplications := []model.Application{
			directorApplication,
		}

		existingRuntimeApplications := v1alpha1.ApplicationList{
			Items: []v1alpha1.Application{},
		}

		converterMock.On("Do", directorApplication).Return(runtimeApplication)
		applicationsManagerMock.On("Create", &runtimeApplication).Return(&runtimeApplication, nil)
		applicationsManagerMock.On("List", metav1.ListOptions{}).Return(&existingRuntimeApplications, nil)

		resourcesServiceMocks.On("CreateApiResources", "id1", runtimeApplication.UID, "API1", mock.MatchedBy(getCredentialsMatcher(api.Credentials)), []byte("spec"), docstopic.OpenApiType).Return(nil)
		resourcesServiceMocks.On("CreateEventApiResources", "id1", "EventAPI1", nilSpec, docstopic.Empty).Return(nil)

		expectedResult := []Result{
			{
				ApplicationID: "id1",
				Operation:     Create,
				Error:         nil,
			},
		}

		// when
		kymaService := NewService(applicationsManagerMock, converterMock, resourcesServiceMocks)
		result, err := kymaService.Apply(directorApplications)

		// then
		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)
		converterMock.AssertExpectations(t)
		applicationsManagerMock.AssertExpectations(t)
		resourcesServiceMocks.AssertExpectations(t)
	})

	t.Run("should apply Update operation", func(t *testing.T) {
		// given
		applicationsManagerMock := &appMocks.Repository{}
		converterMock := &appMocks.Converter{}
		resourcesServiceMocks := &resourcesServiceMocks.Service{}

		api := getTestDirectorAPiDefinition("API1", nil, nil)
		eventAPI := getTestDirectorEventAPIDefinition("EventAPI1", &model.EventAPISpec{
			Data: []byte("spec"), Type: model.EventAPISpecTypeAsyncAPI,
		})

		directorApplication := getTestDirectorApplication("id1", []model.APIDefinition{api}, []model.EventAPIDefinition{eventAPI})

		runtimeService1 := getTestServiceWithCredentials("API1")
		runtimeService2 := getTestServiceWithoutCredentials("EventAPI1")
		runtimeService3 := getTestServiceWithoutCredentials("API2")

		runtimeApplication := getTestApplication("id1", []v1alpha1.Service{runtimeService1, runtimeService2})

		directorApplications := []model.Application{
			directorApplication,
		}

		existingRuntimeApplication := getTestApplication("id1", []v1alpha1.Service{runtimeService1, runtimeService3})
		existingRuntimeApplications := v1alpha1.ApplicationList{
			Items: []v1alpha1.Application{existingRuntimeApplication},
		}

		converterMock.On("Do", directorApplication).Return(runtimeApplication)
		applicationsManagerMock.On("Update", &runtimeApplication).Return(&runtimeApplication, nil)
		applicationsManagerMock.On("List", metav1.ListOptions{}).Return(&existingRuntimeApplications, nil)
		resourcesServiceMocks.On("UpdateApiResources", "id1", types.UID(""), "API1", mock.MatchedBy(getCredentialsMatcher(api.Credentials)), nilSpec, docstopic.Empty).Return(nil)
		resourcesServiceMocks.On("CreateEventApiResources", "id1", "EventAPI1", []byte("spec"), docstopic.AsyncApi).Return(nil)
		resourcesServiceMocks.On("DeleteApiResources", "id1", "API2", "").Return(nil)

		expectedResult := []Result{
			{
				ApplicationID: "id1",
				Operation:     Update,
				Error:         nil,
			},
		}

		// when
		kymaService := NewService(applicationsManagerMock, converterMock, resourcesServiceMocks)
		result, err := kymaService.Apply(directorApplications)

		// then
		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)
		converterMock.AssertExpectations(t)
		applicationsManagerMock.AssertExpectations(t)
		resourcesServiceMocks.AssertExpectations(t)
	})

	t.Run("should apply Delete operation", func(t *testing.T) {
		// given
		applicationsManagerMock := &appMocks.Repository{}
		converterMock := &appMocks.Converter{}
		resourcesServiceMocks := &resourcesServiceMocks.Service{}

		runtimeService := getTestServiceWithCredentials("API1")
		runtimeApplication := getTestApplication("id1", []v1alpha1.Service{runtimeService})

		existingRuntimeApplications := v1alpha1.ApplicationList{
			Items: []v1alpha1.Application{
				runtimeApplication,
			},
		}

		applicationsManagerMock.On("Delete", runtimeApplication.Name, &metav1.DeleteOptions{}).Return(nil)
		applicationsManagerMock.On("List", metav1.ListOptions{}).Return(&existingRuntimeApplications, nil)
		resourcesServiceMocks.On("DeleteApiResources", "id1", "API1", "credentialsSecretName1").Return(nil)

		expectedResult := []Result{
			{
				ApplicationID: "id1",
				Operation:     Delete,
				Error:         nil,
			},
		}

		// when
		kymaService := NewService(applicationsManagerMock, converterMock, resourcesServiceMocks)
		result, err := kymaService.Apply([]model.Application{})

		// then
		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)
		converterMock.AssertExpectations(t)
		applicationsManagerMock.AssertExpectations(t)
		resourcesServiceMocks.AssertExpectations(t)
	})

	t.Run("should not break execution when error occurred when applying Application CR", func(t *testing.T) {
		// given
		applicationsManagerMock := &appMocks.Repository{}
		converterMock := &appMocks.Converter{}
		resourcesServiceMocks := &resourcesServiceMocks.Service{}

		newRuntimeService1 := getTestServiceWithCredentials("API1")
		newRuntimeService2 := getTestServiceWithCredentials("EventAPI1")

		existingRuntimeService1 := getTestServiceWithCredentials("API2")
		existingRuntimeService2 := getTestServiceWithCredentials("EventAPI2")

		runtimeServiceToBeDeleted1 := getTestServiceWithCredentials("API3")
		runtimeServiceToBeDeleted2 := getTestServiceWithCredentials("EventAPI3")

		newDirectorApi := getTestDirectorAPiDefinition("API1", nil, nil)
		newDirectorEventApi := getTestDirectorEventAPIDefinition("EventAPI1", nil)

		newDirectorApplication := getTestDirectorApplication("id1",
			[]model.APIDefinition{newDirectorApi}, []model.EventAPIDefinition{newDirectorEventApi})
		convertedNewRuntimeApplication := getTestApplication("id1", []v1alpha1.Service{newRuntimeService1, newRuntimeService2})

		existingDirectorApi := getTestDirectorAPiDefinition("API2", nil, nil)
		existingDirectorEventApi := getTestDirectorEventAPIDefinition("EventAPI2", nil)

		existingDirectorApplication := getTestDirectorApplication("id2", []model.APIDefinition{newDirectorApi, existingDirectorApi}, []model.EventAPIDefinition{newDirectorEventApi, existingDirectorEventApi})
		convertedExistingRuntimeApplication := getTestApplication("id2", []v1alpha1.Service{newRuntimeService1, newRuntimeService2, existingRuntimeService1, existingRuntimeService2})

		runtimeApplicationToBeDeleted := getTestApplication("id3", []v1alpha1.Service{runtimeServiceToBeDeleted1, runtimeServiceToBeDeleted2})

		directorApplications := []model.Application{
			newDirectorApplication,
			existingDirectorApplication,
		}

		existingRuntimeApplication := getTestApplication("id2", []v1alpha1.Service{existingRuntimeService1, existingRuntimeService2, runtimeServiceToBeDeleted1, runtimeServiceToBeDeleted2})

		existingRuntimeApplications := v1alpha1.ApplicationList{
			Items: []v1alpha1.Application{existingRuntimeApplication,
				runtimeApplicationToBeDeleted},
		}

		converterMock.On("Do", newDirectorApplication).Return(convertedNewRuntimeApplication)
		converterMock.On("Do", existingDirectorApplication).Return(convertedExistingRuntimeApplication)
		applicationsManagerMock.On("Create", &convertedNewRuntimeApplication).Return(nil, apperrors.Internal("some error"))
		applicationsManagerMock.On("Update", &convertedExistingRuntimeApplication).Return(nil, apperrors.Internal("some error"))
		applicationsManagerMock.On("Delete", runtimeApplicationToBeDeleted.Name, &metav1.DeleteOptions{}).Return(apperrors.Internal("some error"))
		applicationsManagerMock.On("List", metav1.ListOptions{}).Return(&existingRuntimeApplications, nil)

		resourcesServiceMocks.On("DeleteApiResources", "id3", "API3", "credentialsSecretName1").Return(apperrors.Internal("some error"))
		resourcesServiceMocks.On("DeleteApiResources", "id3", "EventAPI3", "credentialsSecretName1").Return(apperrors.Internal("some error"))

		// when
		kymaService := NewService(applicationsManagerMock, converterMock, resourcesServiceMocks)
		result, err := kymaService.Apply(directorApplications)

		// then
		require.NoError(t, err)
		require.Equal(t, 3, len(result))
		assert.NotNil(t, result[0].Error)
		assert.NotNil(t, result[1].Error)
		assert.NotNil(t, result[2].Error)
		converterMock.AssertExpectations(t)
		applicationsManagerMock.AssertExpectations(t)
		resourcesServiceMocks.AssertNotCalled(t, "CreateApiResources")
		resourcesServiceMocks.AssertExpectations(t)
	})
}

func getCredentialsMatcher(expected *model.Credentials) func(*secretsmodel.CredentialsWithCSRF) bool {
	return func(credentials *secretsmodel.CredentialsWithCSRF) bool {
		if credentials == nil {
			return expected == nil
		}

		if expected == nil {
			return credentials == nil
		}

		if credentials.Basic != nil && expected.Basic != nil {
			matched := credentials.Basic.Username == expected.Basic.Username && credentials.Basic.Password == expected.Basic.Password
			if !matched {
				return false
			}
		}

		if credentials.Oauth != nil && expected.Oauth != nil {
			matched := credentials.Oauth.ClientID == expected.Oauth.ClientID && credentials.Oauth.ClientSecret == expected.Oauth.ClientSecret
			if !matched {
				return false
			}
		}

		if credentials.CSRFInfo != nil && expected.CSRFInfo != nil {
			return credentials.CSRFInfo.TokenEndpointURL == expected.CSRFInfo.TokenEndpointURL
		}

		return true
	}
}

func getTestDirectorApplication(id string, apiDefinitions []model.APIDefinition, eventApiDefinitions []model.EventAPIDefinition) model.Application {
	return model.Application{
		ID:        id,
		Name:      "First App",
		APIs:      apiDefinitions,
		EventAPIs: eventApiDefinitions,
	}
}

func getTestDirectorAPiDefinition(id string, spec *model.APISpec, credentials *model.Credentials) model.APIDefinition {
	return model.APIDefinition{
		ID:          id,
		Description: "API",
		TargetUrl:   "www.example.com",
		APISpec:     spec,
		Credentials: credentials,
	}
}

func getTestDirectorEventAPIDefinition(id string, spec *model.EventAPISpec) model.EventAPIDefinition {
	return model.EventAPIDefinition{
		ID:           id,
		Description:  "Event API 1",
		EventAPISpec: spec,
	}
}

func getTestServiceWithCredentials(serviceID string) v1alpha1.Service {
	return v1alpha1.Service{
		ID: serviceID,
		Entries: []v1alpha1.Entry{
			{
				Type:             applications.SpecAPIType,
				TargetUrl:        "www.example.com/1",
				SpecificationUrl: "",
				Credentials: v1alpha1.Credentials{
					Type:              applications.CredentialsBasicType,
					SecretName:        "credentialsSecretName1",
					AuthenticationUrl: "",
				},
				RequestParametersSecretName: "paramatersSecretName1",
			},
		},
	}
}

func getTestServiceWithoutCredentials(serviceID string) v1alpha1.Service {
	return v1alpha1.Service{
		ID: serviceID,
		Entries: []v1alpha1.Entry{
			{
				Type:             applications.SpecAPIType,
				TargetUrl:        "www.example.com/1",
				SpecificationUrl: "",
			},
		},
	}
}

func getTestApplication(applicationName string, services []v1alpha1.Service) v1alpha1.Application {
	return v1alpha1.Application{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Application",
			APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: applicationName,
		},
		Spec: v1alpha1.ApplicationSpec{
			Description: "Description",
			Services:    services,
		},
	}
}
