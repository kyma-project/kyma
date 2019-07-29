package kyma

import (
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	resourcesServiceMocks "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/apiresources/mocks"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/applications"
	appMocks "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/applications/mocks"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/model"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/sync"
	syncMocks "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/sync/mocks"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestConverter(t *testing.T) {

	t.Run("should return error in case failed to determine differences between current and desired runtime state", func(t *testing.T) {

	})

	t.Run("should not break execution when error occurred", func(t *testing.T) {

	})

	t.Run("should apply Create operation", func(t *testing.T) {
		// given
		reconcilerMock := &syncMocks.Reconciler{}
		applicationsManagerMock := &appMocks.Manager{}
		converterMock := &appMocks.Converter{}
		resourcesServiceMocks := &resourcesServiceMocks.Service{}

		api := model.APIDefinition{
			ID:          "App1",
			Description: "API",
			TargetUrl:   "www.examle.com",
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

		service := v1alpha1.Service{
			ID:          "serviceId1",
			Identifier:  "",
			Name:        "servicename1-cb830",
			DisplayName: "serviceName1",
			Description: "API 1 description",
			Labels: map[string]string{
				"connected-app": "App1",
			},
			LongDescription:     "",
			ProviderDisplayName: "",
			Tags:                []string{},
			Entries: []v1alpha1.Entry{
				{
					Type:             applications.SpecAPIType,
					GatewayUrl:       "application-gateway.kyma-integration.svc.cluster.local",
					AccessLabel:      "resourceName1",
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

		runtimeApplication := v1alpha1.Application{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Application",
				APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "App1",
			},
			Spec: v1alpha1.ApplicationSpec{
				Description:      "Description",
				SkipInstallation: false,
				AccessLabel:      "App1",
				Labels:           map[string]string{},
				Services: []v1alpha1.Service{
					service,
				},
			},
		}

		directorApplications := []model.Application{
			directorApplication,
		}

		applicationActions := []sync.ApplicationAction{
			{
				Operation:   sync.Create,
				Application: runtimeApplication,
				ServiceActions: []sync.ServiceAction{
					{
						Operation: sync.Create,
						Service:   service,
					},
				},
			},
		}

		converterMock.On("Do", directorApplication).Return(runtimeApplication)
		reconcilerMock.On("Do", []v1alpha1.Application{runtimeApplication}).Return(applicationActions, nil)
		applicationsManagerMock.On("Create", &runtimeApplication).Return(&runtimeApplication, nil)
		resourcesServiceMocks.On("CreateApiResources", runtimeApplication, service).Return(nil)
		resourcesServiceMocks.On("CreateSecrets", runtimeApplication, service).Return(nil)

		expectedResult := []Result{
			{
				ApplicationID: "App1",
				Operation:     sync.Create,
				Error:         nil,
			},
		}

		// when
		kymaService := NewService(reconcilerMock, applicationsManagerMock, converterMock, resourcesServiceMocks)
		result, err := kymaService.Apply(directorApplications)

		// then
		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)
		reconcilerMock.AssertExpectations(t)
		converterMock.AssertExpectations(t)
		applicationsManagerMock.AssertExpectations(t)
		resourcesServiceMocks.AssertExpectations(t)
	})

	t.Run("should apply Update operation", func(t *testing.T) {

	})

	t.Run("should apply Delete operation", func(t *testing.T) {

	})
}
