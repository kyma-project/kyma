package sync

import (
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/applications/mocks"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/model"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestReconciler(t *testing.T) {

	t.Run("should handle application creation", func(t *testing.T) {
		// given
		mockApplicationsInterface := &mocks.Manager{}
		reconciler := NewReconciler(mockApplicationsInterface)

		mockApplicationsInterface.On("List", metav1.ListOptions{}).Return(&v1alpha1.ApplicationList{
			Items: []v1alpha1.Application{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Application",
						APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "id2",
					},
				},
			},
		}, nil)

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

		expectedResult := []ApplicationAction{
			{
				Operation:   Create,
				Application: application1,
				APIActions: []APIAction{
					{
						Operation: Create,
						API:       api,
					},
				},
				EventAPIActions: []EventAPIAction{
					{
						Operation: Create,
						EventAPI:  eventAPI,
					},
				},
			},
		}

		// when
		actions, err := reconciler.Do(directorApplications)

		// then
		assert.NoError(t, err)
		assert.Equal(t, expectedResult, actions)
	})

	t.Run("should handle application deletion", func(t *testing.T) {

	})

	t.Run("should handle application update", func(t *testing.T) {

	})

}
