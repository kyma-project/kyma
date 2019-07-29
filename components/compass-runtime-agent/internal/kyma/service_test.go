package kyma

import (
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	resourcesServiceMocks "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/apiresources/mocks"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/applications"
	appMocks "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/applications/mocks"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/model"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestService(t *testing.T) {

	t.Run("should return error in case failed to determine differences between current and desired runtime state", func(t *testing.T) {

	})

	t.Run("should not break execution when error occurred", func(t *testing.T) {

	})

	t.Run("should apply Create operation", func(t *testing.T) {
		// given
		applicationsManagerMock := &appMocks.Manager{}
		converterMock := &appMocks.Converter{}
		resourcesServiceMocks := &resourcesServiceMocks.Service{}

		api := model.APIDefinition{
			ID:          "API1",
			Description: "API",
			TargetUrl:   "www.examle.com",
			APISpec: &model.APISpec{
				Data: []byte("spec"),
			},
		}

		eventAPI := model.EventAPIDefinition{
			ID:          "eventApi1",
			Description: "Event API 1",
		}

		directorApplication := model.Application{
			ID:   "id1",
			Name: "First App",
			APIs: []model.APIDefinition{
				api,
			},
			EventAPIs: []model.EventAPIDefinition{
				eventAPI,
			},
		}

		runtimeService1 := getTestServiceWithApi("API1")
		runtimeService2 := getTestServiceWithApi("eventApi1")

		runtimeApplication := getTestApplication("id1", []v1alpha1.Service{runtimeService1, runtimeService2})

		directorApplications := []model.Application{
			directorApplication,
		}

		existingRuntimeApplications := v1alpha1.ApplicationList{
			Items: []v1alpha1.Application{
				getTestApplication("id2", []v1alpha1.Service{}),
			},
		}

		converterMock.On("Do", directorApplication).Return(runtimeApplication)
		applicationsManagerMock.On("Create", &runtimeApplication).Return(&runtimeApplication, nil)
		applicationsManagerMock.On("List", metav1.ListOptions{}).Return(&existingRuntimeApplications, nil)
		resourcesServiceMocks.On("CreateApiResources", runtimeApplication, runtimeService1, []byte("spec")).Return(nil)
		resourcesServiceMocks.On("CreateApiResources", runtimeApplication, runtimeService2, []byte(nil)).Return(nil)

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

	})

	t.Run("should apply Delete operation", func(t *testing.T) {

	})
}

func getTestServiceWithApi(serviceID string) v1alpha1.Service {
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
