package applications_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/applications"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/applications/mocks"
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
		application := createApplication("production")
		appManagerMock := &mocks.AppManager{}
		appManagerMock.On("Get", context.Background(), "production", metav1.GetOptions{}).
			Return(application, nil)

		repository := applications.NewServiceRepository(appManagerMock)
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
		assert.Equal(t, s1.API, &applications.ServiceAPI{
			GatewayURL:  "https://orders-gateway.production.svc.cluster.local/",
			AccessLabel: "access-label-1",
			TargetUrl:   "https://192.168.1.2",
			Credentials: applications.Credentials{
				AuthenticationUrl: "https://192.168.1.3/token",
				SecretName:        "app-ac031e8c-9aa4-4cb7-8999-0d358726ffaa",
			},
		})

		s2 := services[1]

		assert.Equal(t, s2.ProviderDisplayName, "SAP Hybris")
		assert.Equal(t, s2.DisplayName, "Products API")
		assert.Equal(t, s2.LongDescription, "This is Products API")
		assert.Equal(t, s2.API, &applications.ServiceAPI{
			GatewayURL:  "https://products-gateway.production.svc.cluster.local/",
			AccessLabel: "access-label-2",
			TargetUrl:   "https://192.168.1.3",
			Credentials: applications.Credentials{
				AuthenticationUrl: "https://192.168.1.4/token",
				SecretName:        "app-bc031e8c-9aa4-4cb7-8999-0d358726ffab",
			},
		})
	})

	t.Run("should fail if unable to read App", func(t *testing.T) {
		// given
		appManagerMock := &mocks.AppManager{}
		appManagerMock.On("Get", context.Background(), "app", metav1.GetOptions{}).
			Return(nil, errors.New("failed to get App"))

		repository := applications.NewServiceRepository(appManagerMock)
		require.NotNil(t, repository)

		// when
		services, err := repository.GetAll("app")

		// then
		require.Nil(t, services)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
	})

	t.Run("should fail if App doesn't exist", func(t *testing.T) {
		// given
		appManagerMock := &mocks.AppManager{}
		appManagerMock.On("Get", context.Background(), "not_existent", metav1.GetOptions{}).
			Return(nil, k8serrors.NewNotFound(schema.GroupResource{}, ""))

		repository := applications.NewServiceRepository(appManagerMock)
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
		assert.Equal(t, applications.Service{}, service)
		assert.Equal(t, apperrors.CodeNotFound, err.Code())
	})

	t.Run("should get service by id", func(t *testing.T) {
		// given
		application := createApplication("production")
		appManagerMock := &mocks.AppManager{}
		appManagerMock.On("Get", context.Background(), "production", metav1.GetOptions{}).
			Return(application, nil)

		repository := applications.NewServiceRepository(appManagerMock)
		require.NotNil(t, repository)

		// when
		service, err := repository.Get("production", "id1")

		// then
		require.NotNil(t, service)
		require.NoError(t, err)

		assert.Equal(t, service.ProviderDisplayName, "SAP Hybris")
		assert.Equal(t, service.DisplayName, "Orders API")
		assert.Equal(t, service.LongDescription, "This is Orders API")
		assert.Equal(t, service.API, &applications.ServiceAPI{
			GatewayURL:  "https://orders-gateway.production.svc.cluster.local/",
			AccessLabel: "access-label-1",
			TargetUrl:   "https://192.168.1.2",
			Credentials: applications.Credentials{
				AuthenticationUrl: "https://192.168.1.3/token",
				SecretName:        "app-ac031e8c-9aa4-4cb7-8999-0d358726ffaa",
			},
		})
	})

	t.Run("should return not found error if service doesn't exist", func(t *testing.T) {
		// given
		application := createApplication("production")
		appManagerMock := &mocks.AppManager{}
		appManagerMock.On("Get", context.Background(), "production", metav1.GetOptions{}).
			Return(application, nil)

		repository := applications.NewServiceRepository(appManagerMock)
		require.NotNil(t, repository)

		// when
		service, err := repository.Get("production", "not-existent")

		// then
		assert.Equal(t, applications.Service{}, service)
		assert.Equal(t, apperrors.CodeNotFound, err.Code())
	})
}

func TestCreateServices(t *testing.T) {
	t.Run("should create service", func(t *testing.T) {
		// given
		application := createApplication("production")

		appManagerMock := &mocks.AppManager{}
		appManagerMock.On("Get", context.Background(), "production", metav1.GetOptions{}).
			Return(application, nil)

		service1 := createK8sService()
		newServices := append(application.Spec.Services, service1)
		newApp := application.DeepCopy()

		newApp.Spec.Services = newServices
		appManagerMock.On("Update", context.Background(), newApp, metav1.UpdateOptions{}).Return(newApp, nil)

		repository := applications.NewServiceRepository(appManagerMock)
		require.NotNil(t, repository)

		newService1 := createService()

		// when
		err := repository.Create("production", newService1)

		// then
		require.NoError(t, err)
		appManagerMock.AssertExpectations(t)
	})

	t.Run("should fail if failed to update App", func(t *testing.T) {
		// given
		application := createApplication("production")
		appManagerMock := &mocks.AppManager{}
		appManagerMock.On("Get", context.Background(), "production", metav1.GetOptions{}).
			Return(application, nil)

		appManagerMock.On("Update", context.Background(), mock.AnythingOfType("*v1alpha1.Application"), metav1.UpdateOptions{}).Return(nil, errors.New("failed to update Application"))

		repository := applications.NewServiceRepository(appManagerMock)
		require.NotNil(t, repository)

		newService1 := createService()

		// when
		err := repository.Create("production", newService1)

		// then
		require.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		appManagerMock.AssertExpectations(t)
	})

	t.Run("should not allow to create service if service with the same id already exists", func(t *testing.T) {
		// given
		application := createApplication("production")
		appManagerMock := &mocks.AppManager{}
		appManagerMock.On("Get", context.Background(), "production", metav1.GetOptions{}).
			Return(application, nil)

		repository := applications.NewServiceRepository(appManagerMock)
		require.NotNil(t, repository)

		newService := applications.Service{
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

	t.Run("should fail if App doesn't exist", func(t *testing.T) {
		// given
		appManagerMock := &mocks.AppManager{}
		appManagerMock.On("Get", context.Background(), "production", metav1.GetOptions{}).
			Return(nil, k8serrors.NewNotFound(schema.GroupResource{}, ""))

		repository := applications.NewServiceRepository(appManagerMock)
		require.NotNil(t, repository)

		// when
		newService := applications.Service{
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
		application := createApplication("production")
		appManagerMock := &mocks.AppManager{}
		appManagerMock.On("Get", context.Background(), "production", metav1.GetOptions{}).
			Return(application, nil)

		newApp := application.DeepCopy()

		newApp.Spec.Services = application.Spec.Services[1:]
		appManagerMock.On("Update", context.Background(), newApp, metav1.UpdateOptions{}).Return(newApp, nil)

		repository := applications.NewServiceRepository(appManagerMock)
		require.NotNil(t, repository)

		// when
		err := repository.Delete("production", "id1")

		// then
		require.NoError(t, err)
		appManagerMock.AssertExpectations(t)
	})

	t.Run("should ignore not found error if service doesn't exist", func(t *testing.T) {
		// given
		application := createApplication("production")
		appManagerMock := &mocks.AppManager{}
		appManagerMock.On("Get", context.Background(), "production", metav1.GetOptions{}).
			Return(application, nil)
		appManagerMock.On("Update", context.Background(), mock.AnythingOfType("*v1alpha1.Application"), metav1.UpdateOptions{}).
			Return(application, nil)

		repository := applications.NewServiceRepository(appManagerMock)
		require.NotNil(t, repository)

		// when
		err := repository.Delete("production", "not-existent")

		// then
		assert.NoError(t, err)
		appManagerMock.AssertExpectations(t)
	})

	t.Run("should fail if failed to update App", func(t *testing.T) {
		// given
		application := createApplication("production")
		appManagerMock := &mocks.AppManager{}
		appManagerMock.On("Get", context.Background(), "production", metav1.GetOptions{}).
			Return(application, nil)
		appManagerMock.On("Update", context.Background(), mock.AnythingOfType("*v1alpha1.Application"), metav1.UpdateOptions{}).Return(nil, errors.New("failed to update App"))

		repository := applications.NewServiceRepository(appManagerMock)
		require.NotNil(t, repository)

		// when
		err := repository.Delete("production", "id1")

		// then
		require.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		appManagerMock.AssertExpectations(t)
	})

	t.Run("should return not found error if App doesn't exist", func(t *testing.T) {
		// given
		appManagerMock := &mocks.AppManager{}
		appManagerMock.On("Get", context.Background(), "production", metav1.GetOptions{}).
			Return(nil, k8serrors.NewNotFound(schema.GroupResource{}, ""))

		repository := applications.NewServiceRepository(appManagerMock)
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
		application := createApplication("production")
		appManagerMock := &mocks.AppManager{}
		appManagerMock.On("Get", context.Background(), "production", metav1.GetOptions{}).
			Return(application, nil)

		reEntry1 := v1alpha1.Entry{
			Name:        "promotions-api-4e89d",
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

		newApp := application.DeepCopy()
		newApp.Spec.Services[0].Name = "promotions-api-4e89d"
		newApp.Spec.Services[0].DisplayName = "Promotions API"
		newApp.Spec.Services[0].ProviderDisplayName = "SAP Labs Poland"
		newApp.Spec.Services[0].LongDescription = "This is Promotions API"
		newApp.Spec.Services[0].Tags = []string{"promotions"}
		newApp.Spec.Services[0].Entries = []v1alpha1.Entry{reEntry1, reEntry2}

		appManagerMock.On("Update", context.Background(), newApp, metav1.UpdateOptions{}).Return(newApp, nil)

		repository := applications.NewServiceRepository(appManagerMock)
		require.NotNil(t, repository)

		service := applications.Service{
			ID:                  "id1",
			Name:                "promotions-api-4e89d",
			DisplayName:         "Promotions API",
			LongDescription:     "This is Promotions API",
			ProviderDisplayName: "SAP Labs Poland",
			Tags:                []string{"promotions"},
			API: &applications.ServiceAPI{
				GatewayURL:  "https://promotions-gateway.production.svc.cluster.local/",
				AccessLabel: "access-label-3",
				TargetUrl:   "https://192.168.10.10",
				Credentials: applications.Credentials{
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
		appManagerMock.AssertExpectations(t)
	})

	t.Run("should return not found error if App doesn't exist", func(t *testing.T) {
		// given
		application := createApplication("production")
		appManagerMock := &mocks.AppManager{}
		appManagerMock.On("Get", context.Background(), "production", metav1.GetOptions{}).
			Return(application, nil)

		repository := applications.NewServiceRepository(appManagerMock)
		require.NotNil(t, repository)

		// when
		service := applications.Service{
			ID: "not-existent",
		}
		err := repository.Update("production", service)

		// then
		assert.Equal(t, apperrors.CodeNotFound, err.Code())
	})
}

func createApplication(name string) *v1alpha1.Application {

	reService1Entry := v1alpha1.Entry{
		Type:        "API",
		GatewayUrl:  "https://orders-gateway.production.svc.cluster.local/",
		AccessLabel: "access-label-1",
		TargetUrl:   "https://192.168.1.2",
		Credentials: v1alpha1.Credentials{
			AuthenticationUrl: "https://192.168.1.3/token",
			SecretName:        "app-ac031e8c-9aa4-4cb7-8999-0d358726ffaa",
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
			SecretName:        "app-bc031e8c-9aa4-4cb7-8999-0d358726ffab",
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

func createService() applications.Service {
	return applications.Service{
		ID:                  "id3",
		DisplayName:         "Promotions API",
		LongDescription:     "This is Promotions API",
		ProviderDisplayName: "SAP Hybris",
		Tags:                []string{"promotions"},
		API: &applications.ServiceAPI{
			GatewayURL:  "https://promotions-gateway.production.svc.cluster.local/",
			AccessLabel: "access-label-1",
			TargetUrl:   "https://192.168.1.2",
			Credentials: applications.Credentials{
				AuthenticationUrl: "https://192.168.1.3/token",
				SecretName:        "app-ac031e8c-9aa4-4cb7-8999-0d358726ffaa",
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
			SecretName:        "app-ac031e8c-9aa4-4cb7-8999-0d358726ffaa",
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
