package kyma

import (
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	resourcesServiceMocks "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/apiresources/mocks"
	appMocks "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/applications/mocks"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/model"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/sync"
	syncMocks "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/sync/mocks"
	"github.com/stretchr/testify/assert"
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
			ID:          "api1",
			Description: "API",
			TargetUrl:   "www.examle.com",
		}

		eventAPI := model.EventAPIDefinition{
			ID:          "eventApi1",
			Description: "Event API 1",
		}

		application1 := model.Application{
			ID:   "id1",
			Name: "First App",
			APIs: []model.APIDefinition{
				api,
			},
			EventAPIs: []model.EventAPIDefinition{
				eventAPI,
			},
		}

		directorApplications := []model.Application{
			application1,
		}

		applicationActions := []sync.ApplicationAction{
			{
				Operation:   sync.Create,
				Application: application1,
				APIActions: []sync.APIAction{
					{
						Operation: sync.Create,
						API:       api,
					},
				},
				EventAPIActions: []sync.EventAPIAction{
					{
						Operation: sync.Create,
						EventAPI:  eventAPI,
					},
				},
			},
		}

		reconcilerMock.On("Do", directorApplications).Return(applicationActions, nil)
		converterMock.On("Do", application1).Return(v1alpha1.Application{})
		applicationsManagerMock.On("Create", &v1alpha1.Application{}).Return(&v1alpha1.Application{}, nil)
		resourcesServiceMocks.On("CreateApiResources", application1, api).Return(nil)
		resourcesServiceMocks.On("CreateEventApiResources", application1, eventAPI).Return(nil)
		resourcesServiceMocks.On("CreateSecrets", application1, api).Return(nil)

		expectedResult := []Result{
			{
				ApplicationID: "id1",
				Operation:     sync.Create,
				Error:         nil,
			},
		}

		// when
		service := NewService(reconcilerMock, applicationsManagerMock, converterMock, resourcesServiceMocks)
		result, err := service.Apply(directorApplications)

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
