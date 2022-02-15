package applications_test

import (
	"context"
	"testing"

	"github.com/kyma-project/kyma/components/application-gateway/internal/metadata/applications"
	"github.com/kyma-project/kyma/components/application-gateway/internal/metadata/applications/mocks"
	"github.com/kyma-project/kyma/components/application-gateway/pkg/apperrors"
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetServices(t *testing.T) {

	t.Run("should get service by id", func(t *testing.T) {
		// given
		app := createApplication("production")
		managerMock := &mocks.Manager{}
		managerMock.On("Get", context.Background(), "production", metav1.GetOptions{}).
			Return(app, nil)

		repository := applications.NewServiceRepository("production", managerMock)
		require.NotNil(t, repository)

		// when
		service, err := repository.Get("id1")

		// then
		require.NotNil(t, service)
		require.NoError(t, err)

		assert.Equal(t, service.ProviderDisplayName, "SAP Hybris")
		assert.Equal(t, service.DisplayName, "Orders API")
		assert.Equal(t, service.LongDescription, "This is Orders API")
		assert.Equal(t, service.API, &applications.ServiceAPI{
			GatewayURL:  "http://re-ec-default-4862c1fb-a047-4add-94e3-c4ff594b3514.kyma-integration.svc.cluster.local",
			AccessLabel: "access-label-1",
			TargetUrl:   "https://192.168.1.2",
			Credentials: &applications.Credentials{
				Type:       "OAuth",
				SecretName: "SecretName",
				Url:        "www.example.com/token",
			},
		})
	})

	t.Run("should return not found error if service doesn't exist", func(t *testing.T) {
		// given
		app := createApplication("production")
		reManagerMock := &mocks.Manager{}
		reManagerMock.On("Get", context.Background(), "production", metav1.GetOptions{}).
			Return(app, nil)

		repository := applications.NewServiceRepository("production", reManagerMock)
		require.NotNil(t, repository)

		// when
		service, err := repository.Get("not-existent")

		// then
		assert.Equal(t, applications.Service{}, service)
		assert.Equal(t, apperrors.CodeNotFound, err.Code())
	})
}

func createApplication(name string) *v1alpha1.Application {

	reService1Entry := v1alpha1.Entry{
		Type:        "API",
		GatewayUrl:  "http://re-ec-default-4862c1fb-a047-4add-94e3-c4ff594b3514.kyma-integration.svc.cluster.local",
		AccessLabel: "access-label-1",
		TargetUrl:   "https://192.168.1.2",
		Credentials: v1alpha1.Credentials{
			Type:              "OAuth",
			SecretName:        "SecretName",
			AuthenticationUrl: "www.example.com/token",
		},
	}
	reService1 := v1alpha1.Service{
		ID:                  "id1",
		DisplayName:         "Orders API",
		LongDescription:     "This is Orders API",
		ProviderDisplayName: "SAP Hybris",
		Tags:                []string{"orders"},
		Entries:             []v1alpha1.Entry{reService1Entry},
	}

	reService2Entry := v1alpha1.Entry{
		Type:        "API",
		GatewayUrl:  "http://re-ec-default-f8c614e1-867e-4887-9f6c-334ca2603251.kyma-integration.svc.cluster.local",
		AccessLabel: "access-label-2",
		TargetUrl:   "https://192.168.1.3",
		Credentials: v1alpha1.Credentials{
			Type:              "OAuth",
			SecretName:        "SecretName",
			AuthenticationUrl: "www.example.com/token",
		},
	}

	reService2 := v1alpha1.Service{
		ID:                  "id2",
		DisplayName:         "Products API",
		LongDescription:     "This is Products API",
		ProviderDisplayName: "SAP Hybris",
		Tags:                []string{"products"},
		Entries:             []v1alpha1.Entry{reService2Entry},
	}

	reSpec1 := v1alpha1.ApplicationSpec{
		Description: "test_1",
		Services: []v1alpha1.Service{
			reService1,
			reService2,
		},
	}

	return &v1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec:       reSpec1,
	}
}
