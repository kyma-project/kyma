package applications_test

import (
	"context"
	"testing"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/central-application-gateway/internal/metadata/applications"
	"github.com/kyma-project/kyma/components/central-application-gateway/internal/metadata/applications/mocks"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/apperrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetServices(t *testing.T) {

	type testFunc func(applications.ServiceRepository) (applications.Service, apperrors.AppError)

	type testcase struct {
		description        string
		application        *v1alpha1.Application
		testFunc           testFunc
		expectedServiceAPI applications.ServiceAPI
	}

	expectedServiceAPI := applications.ServiceAPI{
		TargetURL: "https://192.168.1.2",
		Credentials: &applications.Credentials{
			Type:       "OAuth",
			SecretName: "SecretName",
			URL:        "www.example.com/token",
		},
	}

	for _, testCase := range []testcase{
		{
			description: "should get service by service name",
			application: createApplication("production"),
			testFunc: func(repository applications.ServiceRepository) (applications.Service, apperrors.AppError) {
				return repository.GetByServiceName("production", "service-1")
			},
			expectedServiceAPI: expectedServiceAPI,
		},
		{
			description: "should get service by service and entry name",
			application: createApplication("production"),
			testFunc: func(repository applications.ServiceRepository) (applications.Service, apperrors.AppError) {
				return repository.GetByEntryName("production", "service-1", "service-entry-1")
			},
			expectedServiceAPI: expectedServiceAPI,
		},
	} {
		t.Run(testCase.description, func(t *testing.T) {
			// given
			managerMock := &mocks.Manager{}
			managerMock.On("Get", context.Background(), "production", metav1.GetOptions{}).
				Return(testCase.application, nil)

			repository := applications.NewServiceRepository(managerMock)
			require.NotNil(t, repository)

			// when
			service, err := testCase.testFunc(repository)

			// then
			require.NotNil(t, service)
			require.NoError(t, err)

			assert.Equal(t, service.ProviderDisplayName, "SAP Hybris")
			assert.Equal(t, service.DisplayName, "Service 1")
			assert.Equal(t, service.LongDescription, "This is Orders API")
			assert.Equal(t, service.API, &testCase.expectedServiceAPI)
		})
	}

	for _, testCase := range []testcase{
		{
			description: "should return not found error if service doesn't exist",
			application: createApplication("production"),
			testFunc: func(repository applications.ServiceRepository) (applications.Service, apperrors.AppError) {
				return repository.GetByServiceName("production", "not-exists")
			},
		},
		{
			description: "should return not found error if service doesn't exist",
			application: createApplication("production"),
			testFunc: func(repository applications.ServiceRepository) (applications.Service, apperrors.AppError) {
				return repository.GetByEntryName("production", "not-exists", "service-entry-1")
			},
			expectedServiceAPI: expectedServiceAPI,
		},
		{
			description: "should return not found error if service entry doesn't exist",
			application: createApplication("production"),
			testFunc: func(repository applications.ServiceRepository) (applications.Service, apperrors.AppError) {
				return repository.GetByEntryName("production", "service-1", "not-exists")
			},
			expectedServiceAPI: expectedServiceAPI,
		},
	} {
		t.Run("should return not found error if service doesn't exist", func(t *testing.T) {
			// given
			managerMock := &mocks.Manager{}
			managerMock.On("Get", context.Background(), "production", metav1.GetOptions{}).
				Return(testCase.application, nil)

			repository := applications.NewServiceRepository(managerMock)
			require.NotNil(t, repository)

			// when
			service, err := testCase.testFunc(repository)

			// then
			assert.Equal(t, applications.Service{}, service)
			assert.Equal(t, apperrors.CodeNotFound, err.Code())
		})
	}

	for _, testCase := range []testcase{

		{
			description: "should return bad request error if service has multiple entries with the same names",
			application: createApplication("production"),
			testFunc: func(repository applications.ServiceRepository) (applications.Service, apperrors.AppError) {
				return repository.GetByEntryName("production", "service-3", "service-entry-duplicate")
			},
		},
		{
			description: "should return bad request error if there are multiple services with the same name",
			application: createApplication("production"),
			testFunc: func(repository applications.ServiceRepository) (applications.Service, apperrors.AppError) {
				return repository.GetByServiceName("production", "service-4")
			},
		},
	} {
		t.Run("should return bad request error if multiple services were found", func(t *testing.T) {
			// given
			managerMock := &mocks.Manager{}
			managerMock.On("Get", context.Background(), "production", metav1.GetOptions{}).
				Return(testCase.application, nil)

			repository := applications.NewServiceRepository(managerMock)
			require.NotNil(t, repository)

			// when
			service, err := testCase.testFunc(repository)

			// then
			assert.Equal(t, applications.Service{}, service)
			assert.Equal(t, apperrors.CodeWrongInput, err.Code())
		})
	}

}

func createApplication(name string) *v1alpha1.Application {

	service1Entry := v1alpha1.Entry{
		Type:      "API",
		Name:      "Service entry 1",
		TargetUrl: "https://192.168.1.2",
		Credentials: v1alpha1.Credentials{
			Type:              "OAuth",
			SecretName:        "SecretName",
			AuthenticationUrl: "www.example.com/token",
		},
	}
	service1 := v1alpha1.Service{
		ID:                  "id1",
		Name:                "service-1",
		DisplayName:         "Service 1",
		LongDescription:     "This is Orders API",
		ProviderDisplayName: "SAP Hybris",
		Tags:                []string{"orders"},
		Entries:             []v1alpha1.Entry{service1Entry},
	}

	service2Entry := v1alpha1.Entry{
		Type:      "API",
		TargetUrl: "https://192.168.1.3",
		Credentials: v1alpha1.Credentials{
			Type:              "OAuth",
			SecretName:        "SecretName",
			AuthenticationUrl: "www.example.com/token",
		},
	}

	service2 := v1alpha1.Service{
		ID:                  "id2",
		DisplayName:         "Products API",
		LongDescription:     "This is Products API",
		ProviderDisplayName: "SAP Hybris",
		Tags:                []string{"products"},
		Entries:             []v1alpha1.Entry{service2Entry},
	}

	service3Entry1 := v1alpha1.Entry{
		Name:      "Service entry duplicate",
		Type:      "API",
		TargetUrl: "https://192.168.1.3",
		Credentials: v1alpha1.Credentials{
			Type:              "OAuth",
			SecretName:        "SecretName",
			AuthenticationUrl: "www.example.com/token",
		},
	}

	service3Entry2 := v1alpha1.Entry{
		Name:      "Service entry duplicate",
		Type:      "API",
		TargetUrl: "https://192.168.1.3",
		Credentials: v1alpha1.Credentials{
			Type:              "OAuth",
			SecretName:        "SecretName",
			AuthenticationUrl: "www.example.com/token",
		},
	}
	service3Entry3 := v1alpha1.Entry{
		Name:      "Service entry 3",
		Type:      "API",
		TargetUrl: "https://192.168.1.3",
		Credentials: v1alpha1.Credentials{
			Type:              "OAuth",
			SecretName:        "SecretName",
			AuthenticationUrl: "www.example.com/token",
		},
	}

	service3 := v1alpha1.Service{
		Name:        "service-3",
		DisplayName: "Service 3",
		Entries:     []v1alpha1.Entry{service3Entry1, service3Entry2, service3Entry3},
	}

	service4Entry := v1alpha1.Entry{
		Name:      "Service entry 4",
		Type:      "API",
		TargetUrl: "https://192.168.1.3",
		Credentials: v1alpha1.Credentials{
			Type:              "OAuth",
			SecretName:        "SecretName",
			AuthenticationUrl: "www.example.com/token",
		},
	}
	service4 := v1alpha1.Service{
		Name:        "service-4",
		DisplayName: "Service 4",
		Entries:     []v1alpha1.Entry{service4Entry},
	}

	spec1 := v1alpha1.ApplicationSpec{
		Description: "test_1",
		Services: []v1alpha1.Service{
			service1,
			service2,
			service3,
			// duplicate services
			service4,
			service4,
		},
	}

	return &v1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec:       spec1,
	}
}
