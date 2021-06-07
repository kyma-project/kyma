package metadata

import (
	"testing"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/central-application-gateway/internal/metadata/applications"
	applicationmocks "github.com/kyma-project/kyma/components/central-application-gateway/internal/metadata/applications/mocks"
	"github.com/kyma-project/kyma/components/central-application-gateway/internal/metadata/model"
	serviceapimocks "github.com/kyma-project/kyma/components/central-application-gateway/internal/metadata/serviceapi/mocks"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/apperrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceDefinitionService_GetAPI(t *testing.T) {

	type testFunc func(ServiceDefinitionService) (*model.API, apperrors.AppError)
	type getRepositoryMockFunction func() *applicationmocks.ServiceRepository

	type testcase struct {
		description               string
		application               *v1alpha1.Application
		testFunc                  testFunc
		getRepositoryMockFunction getRepositoryMockFunction
		checkExpectedErrorMessage bool
		expectedError             apperrors.AppError
	}

	applicationServiceAPI := &applications.ServiceAPI{}
	applicationService := applications.Service{API: applicationServiceAPI}

	for _, testCase := range []testcase{
		{
			description: "should get API by service name",
			testFunc: func(serviceDefService ServiceDefinitionService) (*model.API, apperrors.AppError) {
				return serviceDefService.GetAPIByServiceName("app", "service")
			},
			getRepositoryMockFunction: func() *applicationmocks.ServiceRepository {
				serviceRepository := applicationmocks.ServiceRepository{}
				serviceRepository.On("GetByServiceName", "app", "service").Return(applicationService, nil)

				return &serviceRepository
			},
		},
		{
			description: "should get API by entry name",
			testFunc: func(serviceDefService ServiceDefinitionService) (*model.API, apperrors.AppError) {
				return serviceDefService.GetAPIByEntryName("app", "service", "entry")
			},
			getRepositoryMockFunction: func() *applicationmocks.ServiceRepository {
				serviceRepository := applicationmocks.ServiceRepository{}
				serviceRepository.On("GetByEntryName", "app", "service", "entry").Return(applicationService, nil)

				return &serviceRepository
			},
		},
	} {
		t.Run(testCase.description, func(t *testing.T) {
			// given
			serviceAPI := &model.API{}

			serviceAPIService := new(serviceapimocks.Service)
			serviceAPIService.On("Read", applicationServiceAPI).Return(serviceAPI, nil)

			service := NewServiceDefinitionService(serviceAPIService, testCase.getRepositoryMockFunction())

			// when
			result, err := testCase.testFunc(service)

			// then
			require.NoError(t, err)

			assert.Equal(t, serviceAPI, result)
		})
	}

	testGetByServiceNameFunc := func(serviceDefService ServiceDefinitionService) (*model.API, apperrors.AppError) {
		return serviceDefService.GetAPIByServiceName("app", "service")
	}

	testGetByEntryNameFunc := func(serviceDefService ServiceDefinitionService) (*model.API, apperrors.AppError) {
		return serviceDefService.GetAPIByEntryName("app", "service", "entry")
	}

	for _, testCase := range []testcase{
		{
			description: "should return not found error if service does not exist",
			testFunc:    testGetByServiceNameFunc,
			getRepositoryMockFunction: func() *applicationmocks.ServiceRepository {
				serviceRepository := applicationmocks.ServiceRepository{}
				serviceRepository.On("GetByServiceName", "app", "service").Return(applications.Service{}, apperrors.NotFound("missing"))

				return &serviceRepository
			},
			checkExpectedErrorMessage: false,
			expectedError:             apperrors.NotFound("missing"),
		},
		{
			description: "should return not found error if service entry does not exist",
			testFunc:    testGetByEntryNameFunc,
			getRepositoryMockFunction: func() *applicationmocks.ServiceRepository {
				serviceRepository := applicationmocks.ServiceRepository{}
				serviceRepository.On("GetByEntryName", "app", "service", "entry").Return(applications.Service{}, apperrors.NotFound("missing"))

				return &serviceRepository
			},
			checkExpectedErrorMessage: false,
			expectedError:             apperrors.NotFound("missing"),
		},
		{
			description: "should return internal error if failed to get service",
			testFunc:    testGetByServiceNameFunc,
			getRepositoryMockFunction: func() *applicationmocks.ServiceRepository {
				serviceRepository := applicationmocks.ServiceRepository{}
				serviceRepository.On("GetByServiceName", "app", "service").Return(applications.Service{}, apperrors.Internal("some error"))

				return &serviceRepository
			},
			checkExpectedErrorMessage: true,
			expectedError:             apperrors.Internal("some error"),
		},
		{
			description: "should return internal error if failed to get service entry",
			testFunc:    testGetByEntryNameFunc,
			getRepositoryMockFunction: func() *applicationmocks.ServiceRepository {
				serviceRepository := applicationmocks.ServiceRepository{}
				serviceRepository.On("GetByEntryName", "app", "service", "entry").Return(applications.Service{}, apperrors.Internal("some error"))

				return &serviceRepository
			},
			checkExpectedErrorMessage: true,
			expectedError:             apperrors.Internal("some error"),
		},
		{
			description: "should return bad request if service does not have API",
			testFunc:    testGetByServiceNameFunc,
			getRepositoryMockFunction: func() *applicationmocks.ServiceRepository {
				serviceRepository := applicationmocks.ServiceRepository{}
				serviceRepository.On("GetByServiceName", "app", "service").Return(applications.Service{}, nil)

				return &serviceRepository
			},
			checkExpectedErrorMessage: false,
			expectedError:             apperrors.WrongInput("some error"),
		},
		{
			description: "should return bad request if service entry does not have API",
			testFunc:    testGetByEntryNameFunc,
			getRepositoryMockFunction: func() *applicationmocks.ServiceRepository {
				serviceRepository := applicationmocks.ServiceRepository{}
				serviceRepository.On("GetByEntryName", "app", "service", "entry").Return(applications.Service{}, nil)

				return &serviceRepository
			},
			checkExpectedErrorMessage: false,
			expectedError:             apperrors.WrongInput("some error"),
		},
	} {
		t.Run(testCase.description, func(t *testing.T) {
			// given
			service := NewServiceDefinitionService(nil, testCase.getRepositoryMockFunction())

			// when
			result, err := testCase.testFunc(service)

			// then
			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Equal(t, testCase.expectedError.Code(), err.Code())

			if testCase.checkExpectedErrorMessage {
				assert.Contains(t, err.Error(), testCase.expectedError.Error())
			}
		})
	}

	for _, testCase := range []testcase{
		{
			description: "should return internal error if reading service API fails",
			testFunc: func(serviceDefService ServiceDefinitionService) (*model.API, apperrors.AppError) {
				return serviceDefService.GetAPIByServiceName("app", "service")
			},
			getRepositoryMockFunction: func() *applicationmocks.ServiceRepository {
				serviceRepository := applicationmocks.ServiceRepository{}
				serviceRepository.On("GetByServiceName", "app", "service").Return(applicationService, nil)

				return &serviceRepository
			},
		},
		{
			description: "should return internal error if reading service API fails",
			testFunc: func(serviceDefService ServiceDefinitionService) (*model.API, apperrors.AppError) {
				return serviceDefService.GetAPIByEntryName("app", "service", "entry")
			},
			getRepositoryMockFunction: func() *applicationmocks.ServiceRepository {
				serviceRepository := applicationmocks.ServiceRepository{}
				serviceRepository.On("GetByEntryName", "app", "service", "entry").Return(applicationService, nil)

				return &serviceRepository
			},
		},
	} {
		t.Run("should return internal error if reading service API fails", func(t *testing.T) {
			// given
			applicationServiceAPI := &applications.ServiceAPI{}

			serviceAPIService := new(serviceapimocks.Service)
			serviceAPIService.On("Read", applicationServiceAPI).Return(nil, apperrors.Internal("some error"))

			service := NewServiceDefinitionService(serviceAPIService, testCase.getRepositoryMockFunction())

			// when
			result, err := testCase.testFunc(service)

			// then
			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Equal(t, apperrors.CodeInternal, err.Code())
			assert.Contains(t, err.Error(), "some error")
		})
	}
}
