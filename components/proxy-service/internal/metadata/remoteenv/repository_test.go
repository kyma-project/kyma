package remoteenv_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/proxy-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/proxy-service/internal/metadata/remoteenv"
	"github.com/kyma-project/kyma/components/proxy-service/internal/metadata/remoteenv/mocks"
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetServices(t *testing.T) {

	t.Run("should get service by id", func(t *testing.T) {
		// given
		remoteEnvironment := createRemoteEnvironment("production")
		reManagerMock := &mocks.Manager{}
		reManagerMock.On("Get", "production", metav1.GetOptions{}).
			Return(remoteEnvironment, nil)

		repository := remoteenv.NewServiceRepository("production", reManagerMock)
		require.NotNil(t, repository)

		// when
		service, err := repository.Get("id1")

		// then
		require.NotNil(t, service)
		require.NoError(t, err)

		assert.Equal(t, service.ProviderDisplayName, "SAP Hybris")
		assert.Equal(t, service.DisplayName, "Orders API")
		assert.Equal(t, service.LongDescription, "This is Orders API")
		assert.Equal(t, service.API, &remoteenv.ServiceAPI{
			GatewayURL:            "http://re-ec-default-4862c1fb-a047-4add-94e3-c4ff594b3514.kyma-integration.svc.cluster.local",
			AccessLabel:           "access-label-1",
			TargetUrl:             "https://192.168.1.2",
			OauthUrl:              "https://192.168.1.3/token",
			CredentialsSecretName: "re-ac031e8c-9aa4-4cb7-8999-0d358726ffaa",
		})
	})

	t.Run("should return not found error if service doesn't exist", func(t *testing.T) {
		// given
		remoteEnvironment := createRemoteEnvironment("production")
		reManagerMock := &mocks.Manager{}
		reManagerMock.On("Get", "production", metav1.GetOptions{}).
			Return(remoteEnvironment, nil)

		repository := remoteenv.NewServiceRepository("production", reManagerMock)
		require.NotNil(t, repository)

		// when
		service, err := repository.Get("not-existent")

		// then
		assert.Equal(t, remoteenv.Service{}, service)
		assert.Equal(t, apperrors.CodeNotFound, err.Code())
	})
}

func createRemoteEnvironment(name string) *v1alpha1.RemoteEnvironment {

	reService1Entry := v1alpha1.Entry{
		Type:                  "API",
		GatewayUrl:            "http://re-ec-default-4862c1fb-a047-4add-94e3-c4ff594b3514.kyma-integration.svc.cluster.local",
		AccessLabel:           "access-label-1",
		TargetUrl:             "https://192.168.1.2",
		OauthUrl:              "https://192.168.1.3/token",
		CredentialsSecretName: "re-ac031e8c-9aa4-4cb7-8999-0d358726ffaa",
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
		Type:                  "API",
		GatewayUrl:            "http://re-ec-default-f8c614e1-867e-4887-9f6c-334ca2603251.kyma-integration.svc.cluster.local",
		AccessLabel:           "access-label-2",
		TargetUrl:             "https://192.168.1.3",
		OauthUrl:              "https://192.168.1.4/token",
		CredentialsSecretName: "re-bc031e8c-9aa4-4cb7-8999-0d358726ffab",
	}

	reService2 := v1alpha1.Service{
		ID:                  "id2",
		DisplayName:         "Products API",
		LongDescription:     "This is Products API",
		ProviderDisplayName: "SAP Hybris",
		Tags:                []string{"products"},
		Entries:             []v1alpha1.Entry{reService2Entry},
	}

	reSpec1 := v1alpha1.RemoteEnvironmentSpec{
		Description: "test_1",
		Services: []v1alpha1.Service{
			reService1,
			reService2,
		},
	}

	return &v1alpha1.RemoteEnvironment{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec:       reSpec1,
	}
}
