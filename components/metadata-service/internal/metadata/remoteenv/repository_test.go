package remoteenv_test

import (
	"errors"
	"testing"

	"github.com/kyma-project/kyma/components/metadata-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/metadata-service/internal/metadata/remoteenv"
	"github.com/kyma-project/kyma/components/metadata-service/internal/metadata/remoteenv/mocks"
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestGetServices(t *testing.T) {
	t.Run("should get all services", func(t *testing.T) {
		// given
		remoteEnvironment := createRemoteEnvironment("production")
		reManagerMock := &mocks.RemoteEnvironmentManager{}
		reManagerMock.On("Get", "production", metav1.GetOptions{}).
			Return(remoteEnvironment, nil)

		repository := remoteenv.NewServiceRepository(reManagerMock)
		require.NotNil(t, repository)

		// when
		services, err := repository.GetAll("production")

		// then
		require.NoError(t, err)
		require.NotNil(t, services)

		assert.Equal(t, len(services), 2)
		s1 := services[0]

		assert.Equal(t, s1.ProviderDisplayName, "SAP Hybris")
		assert.Equal(t, s1.DisplayName, "Orders API")
		assert.Equal(t, s1.LongDescription, "This is Orders API")
		assert.Equal(t, s1.API, &remoteenv.ServiceAPI{
			GatewayURL:  "https://orders-gateway.production.svc.cluster.local/",
			AccessLabel: "access-label-1",
			TargetUrl:   "https://192.168.1.2",
			Credentials: remoteenv.Credentials{
				AuthenticationUrl: "https://192.168.1.3/token",
				SecretName:        "re-ac031e8c-9aa4-4cb7-8999-0d358726ffaa",
			},
		})

		s2 := services[1]

		assert.Equal(t, s2.ProviderDisplayName, "SAP Hybris")
		assert.Equal(t, s2.DisplayName, "Products API")
		assert.Equal(t, s2.LongDescription, "This is Products API")
		assert.Equal(t, s2.API, &remoteenv.ServiceAPI{
			GatewayURL:  "https://products-gateway.production.svc.cluster.local/",
			AccessLabel: "access-label-2",
			TargetUrl:   "https://192.168.1.3",
			Credentials: remoteenv.Credentials{
				AuthenticationUrl: "https://192.168.1.4/token",
				SecretName:        "re-bc031e8c-9aa4-4cb7-8999-0d358726ffab",
			},
		})
	})

	t.Run("should fail if unable to read RE", func(t *testing.T) {
		// given
		reManagerMock := &mocks.RemoteEnvironmentManager{}
		reManagerMock.On("Get", "re", metav1.GetOptions{}).
			Return(nil, errors.New("failed to get RE"))

		repository := remoteenv.NewServiceRepository(reManagerMock)
		require.NotNil(t, repository)

		// when
		services, err := repository.GetAll("re")

		// then
		require.Nil(t, services)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
	})

	t.Run("should fail if RE doesn't exist", func(t *testing.T) {
		// given
		reManagerMock := &mocks.RemoteEnvironmentManager{}
		reManagerMock.On("Get", "not_existent", metav1.GetOptions{}).
			Return(nil, k8serrors.NewNotFound(schema.GroupResource{}, ""))

		repository := remoteenv.NewServiceRepository(reManagerMock)
		require.NotNil(t, repository)

		// when
		services, err := repository.GetAll("not_existent")

		// then
		require.Nil(t, services)
		assert.Equal(t, apperrors.CodeNotFound, err.Code())

		// when
		service, err := repository.Get("not_existent", "id1")

		// then
		require.Error(t, err)
		assert.Equal(t, remoteenv.Service{}, service)
		assert.Equal(t, apperrors.CodeNotFound, err.Code())
	})

	t.Run("should get service by id", func(t *testing.T) {
		// given
		remoteEnvironment := createRemoteEnvironment("production")
		reManagerMock := &mocks.RemoteEnvironmentManager{}
		reManagerMock.On("Get", "production", metav1.GetOptions{}).
			Return(remoteEnvironment, nil)

		repository := remoteenv.NewServiceRepository(reManagerMock)
		require.NotNil(t, repository)

		// when
		service, err := repository.Get("production", "id1")

		// then
		require.NotNil(t, service)
		require.NoError(t, err)

		assert.Equal(t, service.ProviderDisplayName, "SAP Hybris")
		assert.Equal(t, service.DisplayName, "Orders API")
		assert.Equal(t, service.LongDescription, "This is Orders API")
		assert.Equal(t, service.API, &remoteenv.ServiceAPI{
			GatewayURL:  "https://orders-gateway.production.svc.cluster.local/",
			AccessLabel: "access-label-1",
			TargetUrl:   "https://192.168.1.2",
			Credentials: remoteenv.Credentials{
				AuthenticationUrl: "https://192.168.1.3/token",
				SecretName:        "re-ac031e8c-9aa4-4cb7-8999-0d358726ffaa",
			},
		})
	})

	t.Run("should return not found error if service doesn't exist", func(t *testing.T) {
		// given
		remoteEnvironment := createRemoteEnvironment("production")
		reManagerMock := &mocks.RemoteEnvironmentManager{}
		reManagerMock.On("Get", "production", metav1.GetOptions{}).
			Return(remoteEnvironment, nil)

		repository := remoteenv.NewServiceRepository(reManagerMock)
		require.NotNil(t, repository)

		// when
		service, err := repository.Get("production", "not-existent")

		// then
		assert.Equal(t, remoteenv.Service{}, service)
		assert.Equal(t, apperrors.CodeNotFound, err.Code())
	})
}

func TestCreateServices(t *testing.T) {
	t.Run("should create service", func(t *testing.T) {
		// given
		remoteEnvironment := createRemoteEnvironment("production")
		reManagerMock := &mocks.RemoteEnvironmentManager{}
		reManagerMock.On("Get", "production", metav1.GetOptions{}).
			Return(remoteEnvironment, nil)

		service1 := createK8sService()
		newServices := append(remoteEnvironment.Spec.Services, service1)
		newRE := remoteEnvironment.DeepCopy()

		newRE.Spec.Services = newServices
		reManagerMock.On("Update", newRE).Return(newRE, nil)

		repository := remoteenv.NewServiceRepository(reManagerMock)
		require.NotNil(t, repository)

		newService1 := createService()

		// when
		err := repository.Create("production", newService1)

		// then
		require.NoError(t, err)
		reManagerMock.AssertExpectations(t)
	})

	t.Run("should fail if failed to update RE", func(t *testing.T) {
		// given
		remoteEnvironment := createRemoteEnvironment("production")
		reManagerMock := &mocks.RemoteEnvironmentManager{}
		reManagerMock.On("Get", "production", metav1.GetOptions{}).
			Return(remoteEnvironment, nil)

		reManagerMock.On("Update", mock.AnythingOfType("*v1alpha1.RemoteEnvironment")).Return(nil, errors.New("failed to update RE"))

		repository := remoteenv.NewServiceRepository(reManagerMock)
		require.NotNil(t, repository)

		newService1 := createService()

		// when
		err := repository.Create("production", newService1)

		// then
		require.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		reManagerMock.AssertExpectations(t)
	})

	t.Run("should not allow to create service if service with the same id already exists", func(t *testing.T) {
		// given
		remoteEnvironment := createRemoteEnvironment("production")
		reManagerMock := &mocks.RemoteEnvironmentManager{}
		reManagerMock.On("Get", "production", metav1.GetOptions{}).
			Return(remoteEnvironment, nil)

		repository := remoteenv.NewServiceRepository(reManagerMock)
		require.NotNil(t, repository)

		newService := remoteenv.Service{
			ID:              "id1",
			Name:            "promotions-api-c48fe",
			DisplayName:     "Promotions API",
			LongDescription: "This is Promotions API",
			Tags:            []string{"promotions"},
			Events:          true,
		}
		// when
		err := repository.Create("production", newService)

		// then
		assert.Equal(t, apperrors.CodeAlreadyExists, err.Code())
	})

	t.Run("should fail if RE doesn't exist", func(t *testing.T) {
		// given
		reManagerMock := &mocks.RemoteEnvironmentManager{}
		reManagerMock.On("Get", "production", metav1.GetOptions{}).
			Return(nil, k8serrors.NewNotFound(schema.GroupResource{}, ""))

		repository := remoteenv.NewServiceRepository(reManagerMock)
		require.NotNil(t, repository)

		// when
		newService := remoteenv.Service{
			ID:              "id1",
			DisplayName:     "Promotions API",
			LongDescription: "This is Promotions API",
			Tags:            []string{"promotions"},
			Events:          true,
		}
		err := repository.Update("production", newService)

		// then
		assert.Equal(t, apperrors.CodeNotFound, err.Code())
	})
}

func TestDeleteServices(t *testing.T) {

	t.Run("should delete service", func(t *testing.T) {
		// given
		remoteEnvironment := createRemoteEnvironment("production")
		reManagerMock := &mocks.RemoteEnvironmentManager{}
		reManagerMock.On("Get", "production", metav1.GetOptions{}).
			Return(remoteEnvironment, nil)

		newRE := remoteEnvironment.DeepCopy()

		newRE.Spec.Services = remoteEnvironment.Spec.Services[1:]
		reManagerMock.On("Update", newRE).Return(newRE, nil)

		repository := remoteenv.NewServiceRepository(reManagerMock)
		require.NotNil(t, repository)

		// when
		err := repository.Delete("production", "id1")

		// then
		require.NoError(t, err)
		reManagerMock.AssertExpectations(t)
	})

	t.Run("should ignore not found error if service doesn't exist", func(t *testing.T) {
		// given
		remoteEnvironment := createRemoteEnvironment("production")
		reManagerMock := &mocks.RemoteEnvironmentManager{}
		reManagerMock.On("Get", "production", metav1.GetOptions{}).
			Return(remoteEnvironment, nil)

		repository := remoteenv.NewServiceRepository(reManagerMock)
		require.NotNil(t, repository)

		// when
		err := repository.Delete("production", "not-existent")

		// then
		assert.NoError(t, err)
		reManagerMock.AssertExpectations(t)
	})

	t.Run("should fail if failed to update RE", func(t *testing.T) {
		// given
		remoteEnvironment := createRemoteEnvironment("production")
		reManagerMock := &mocks.RemoteEnvironmentManager{}
		reManagerMock.On("Get", "production", metav1.GetOptions{}).
			Return(remoteEnvironment, nil)

		reManagerMock.On("Update", mock.AnythingOfType("*v1alpha1.RemoteEnvironment")).Return(nil, errors.New("failed to update RE"))

		repository := remoteenv.NewServiceRepository(reManagerMock)
		require.NotNil(t, repository)

		// when
		err := repository.Delete("production", "id1")

		// then
		require.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		reManagerMock.AssertExpectations(t)
	})

	t.Run("should return not found error if RE doesn't exist", func(t *testing.T) {
		// given
		reManagerMock := &mocks.RemoteEnvironmentManager{}
		reManagerMock.On("Get", "production", metav1.GetOptions{}).
			Return(nil, k8serrors.NewNotFound(schema.GroupResource{}, ""))

		repository := remoteenv.NewServiceRepository(reManagerMock)
		require.NotNil(t, repository)

		// when
		err := repository.Delete("production", "id1")

		// then
		assert.Equal(t, apperrors.CodeNotFound, err.Code())
	})
}

func TestUpdateServices(t *testing.T) {
	t.Run("should update service", func(t *testing.T) {
		// given
		remoteEnvironment := createRemoteEnvironment("production")
		reManagerMock := &mocks.RemoteEnvironmentManager{}
		reManagerMock.On("Get", "production", metav1.GetOptions{}).
			Return(remoteEnvironment, nil)

		reEntry1 := v1alpha1.Entry{
			Type:        "API",
			GatewayUrl:  "https://promotions-gateway.production.svc.cluster.local/",
			AccessLabel: "access-label-3",
			TargetUrl:   "https://192.168.10.10",
			Credentials: v1alpha1.Credentials{
				AuthenticationUrl: "https://192.168.10.10/token",
				SecretName:        "new_secret",
			},
		}

		reEntry2 := v1alpha1.Entry{
			Type: "Events",
		}

		newRE := remoteEnvironment.DeepCopy()
		newRE.Spec.Services[0].Name = "promotions-api-4e89d"
		newRE.Spec.Services[0].DisplayName = "Promotions API"
		newRE.Spec.Services[0].ProviderDisplayName = "SAP Labs Poland"
		newRE.Spec.Services[0].LongDescription = "This is Promotions API"
		newRE.Spec.Services[0].Tags = []string{"promotions"}
		newRE.Spec.Services[0].Entries = []v1alpha1.Entry{reEntry1, reEntry2}

		reManagerMock.On("Update", newRE).Return(newRE, nil)

		repository := remoteenv.NewServiceRepository(reManagerMock)
		require.NotNil(t, repository)

		service := remoteenv.Service{
			ID:                  "id1",
			Name:                "promotions-api-4e89d",
			DisplayName:         "Promotions API",
			LongDescription:     "This is Promotions API",
			ProviderDisplayName: "SAP Labs Poland",
			Tags:                []string{"promotions"},
			API: &remoteenv.ServiceAPI{
				GatewayURL:  "https://promotions-gateway.production.svc.cluster.local/",
				AccessLabel: "access-label-3",
				TargetUrl:   "https://192.168.10.10",
				Credentials: remoteenv.Credentials{
					AuthenticationUrl: "https://192.168.10.10/token",
					SecretName:        "new_secret",
				},
			},
			Events: true,
		}

		// when
		err := repository.Update("production", service)

		// then
		require.NoError(t, err)
		reManagerMock.AssertExpectations(t)
	})

	t.Run("should return not found error if RE doesn't exist", func(t *testing.T) {
		// given
		remoteEnvironment := createRemoteEnvironment("production")
		reManagerMock := &mocks.RemoteEnvironmentManager{}
		reManagerMock.On("Get", "production", metav1.GetOptions{}).
			Return(remoteEnvironment, nil)

		repository := remoteenv.NewServiceRepository(reManagerMock)
		require.NotNil(t, repository)

		// when
		service := remoteenv.Service{
			ID: "not-existent",
		}
		err := repository.Update("production", service)

		// then
		assert.Equal(t, apperrors.CodeNotFound, err.Code())
	})
}

func createRemoteEnvironment(name string) *v1alpha1.RemoteEnvironment {

	reService1Entry := v1alpha1.Entry{
		Type:        "API",
		GatewayUrl:  "https://orders-gateway.production.svc.cluster.local/",
		AccessLabel: "access-label-1",
		TargetUrl:   "https://192.168.1.2",
		Credentials: v1alpha1.Credentials{
			AuthenticationUrl: "https://192.168.1.3/token",
			SecretName:        "re-ac031e8c-9aa4-4cb7-8999-0d358726ffaa",
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
		GatewayUrl:  "https://products-gateway.production.svc.cluster.local/",
		AccessLabel: "access-label-2",
		TargetUrl:   "https://192.168.1.3",
		Credentials: v1alpha1.Credentials{
			AuthenticationUrl: "https://192.168.1.4/token",
			SecretName:        "re-bc031e8c-9aa4-4cb7-8999-0d358726ffab",
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

func createService() remoteenv.Service {
	return remoteenv.Service{
		ID:                  "id3",
		DisplayName:         "Promotions API",
		LongDescription:     "This is Promotions API",
		ProviderDisplayName: "SAP Hybris",
		Tags:                []string{"promotions"},
		API: &remoteenv.ServiceAPI{
			GatewayURL:  "https://promotions-gateway.production.svc.cluster.local/",
			AccessLabel: "access-label-1",
			TargetUrl:   "https://192.168.1.2",
			Credentials: remoteenv.Credentials{
				AuthenticationUrl: "https://192.168.1.3/token",
				SecretName:        "re-ac031e8c-9aa4-4cb7-8999-0d358726ffaa",
			},
		},
		Events: true,
	}
}

func createK8sService() v1alpha1.Service {
	serviceEntry1 := v1alpha1.Entry{
		Type:        "API",
		GatewayUrl:  "https://promotions-gateway.production.svc.cluster.local/",
		AccessLabel: "access-label-1",
		TargetUrl:   "https://192.168.1.2",
		Credentials: v1alpha1.Credentials{
			AuthenticationUrl: "https://192.168.1.3/token",
			SecretName:        "re-ac031e8c-9aa4-4cb7-8999-0d358726ffaa",
		},
	}
	serviceEntry2 := v1alpha1.Entry{
		Type: "Events",
	}

	return v1alpha1.Service{
		ID:                  "id3",
		Name:                "promotions-api-c48fe",
		DisplayName:         "Promotions API",
		LongDescription:     "This is Promotions API",
		ProviderDisplayName: "SAP Hybris",
		Tags:                []string{"promotions"},
		Entries:             []v1alpha1.Entry{serviceEntry1, serviceEntry2},
	}
}
