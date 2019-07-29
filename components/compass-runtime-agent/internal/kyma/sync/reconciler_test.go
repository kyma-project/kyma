package sync

import (
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/applications"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/applications/mocks"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestReconciler(t *testing.T) {

	t.Run("should handle application creation", func(t *testing.T) {
		// given
		mockApplicationsInterface := &mocks.Manager{}
		reconciler := NewReconciler(mockApplicationsInterface)

		services := []v1alpha1.Service{
			{
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
			},
		}

		application := v1alpha1.Application{
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
				Services:         services,
			},
		}

		mockApplicationsInterface.On("List", metav1.ListOptions{}).Return(&v1alpha1.ApplicationList{
			Items: []v1alpha1.Application{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Application",
						APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "App2",
					},
				},
			},
		}, nil)

		expectedResult := []ApplicationAction{
			{
				Operation:   Create,
				Application: application,
				ServiceActions: []ServiceAction{
					{
						Operation: Create,
						Service:   services[0],
					},
				},
			},
		}

		// when
		actions, err := reconciler.Do([]v1alpha1.Application{
			application,
		})

		// then
		assert.NoError(t, err)
		assert.Equal(t, expectedResult, actions)
	})

	t.Run("should handle application deletion", func(t *testing.T) {

	})

	t.Run("should handle application update", func(t *testing.T) {

	})

}
