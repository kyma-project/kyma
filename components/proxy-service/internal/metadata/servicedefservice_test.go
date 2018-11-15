package metadata

import (
	"testing"

	"github.com/kyma-project/kyma/components/proxy-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/proxy-service/internal/metadata/remoteenv"
	remoteenvmocks "github.com/kyma-project/kyma/components/proxy-service/internal/metadata/remoteenv/mocks"
	"github.com/kyma-project/kyma/components/proxy-service/internal/metadata/serviceapi"
	serviceapimocks "github.com/kyma-project/kyma/components/proxy-service/internal/metadata/serviceapi/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	empty []byte
)

func TestServiceDefinitionService_GetAPI(t *testing.T) {

	t.Run("should get API", func(t *testing.T) {
		// given
		remoteEnvServiceAPI := &remoteenv.ServiceAPI{}
		remoteEnvService := remoteenv.Service{API: remoteEnvServiceAPI}
		serviceAPI := &serviceapi.API{}

		serviceRepository := new(remoteenvmocks.ServiceRepository)
		serviceRepository.On("Get", "uuid-1").Return(remoteEnvService, nil)

		serviceAPIService := new(serviceapimocks.Service)
		serviceAPIService.On("Read", remoteEnvServiceAPI).Return(serviceAPI, nil)

		service := NewServiceDefinitionService(serviceAPIService, serviceRepository)

		// when
		result, err := service.GetAPI("uuid-1")

		// then
		require.NoError(t, err)

		assert.Equal(t, serviceAPI, result)
	})

	t.Run("should return not found error if service does not exist", func(t *testing.T) {
		// given
		serviceRepository := new(remoteenvmocks.ServiceRepository)
		serviceRepository.On("Get", "uuid-1").Return(remoteenv.Service{}, apperrors.NotFound("missing"))

		service := NewServiceDefinitionService(nil, serviceRepository)

		// when
		result, err := service.GetAPI("uuid-1")

		// then
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, apperrors.CodeNotFound, err.Code())
	})

	t.Run("should return internal error if service does not exist", func(t *testing.T) {
		// given
		serviceRepository := new(remoteenvmocks.ServiceRepository)
		serviceRepository.On("Get", "uuid-1").Return(remoteenv.Service{}, apperrors.Internal("some error"))

		service := NewServiceDefinitionService(nil, serviceRepository)

		// when
		result, err := service.GetAPI("uuid-1")

		// then
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		assert.Contains(t, err.Error(), "some error")
	})

	t.Run("should return bad request if service does not have API", func(t *testing.T) {
		// given
		serviceRepository := new(remoteenvmocks.ServiceRepository)
		serviceRepository.On("Get", "uuid-1").Return(remoteenv.Service{}, nil)

		service := NewServiceDefinitionService(nil, serviceRepository)

		// when
		result, err := service.GetAPI("uuid-1")

		// then
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, apperrors.CodeWrongInput, err.Code())
	})

	t.Run("should return internal error if reading service API fails", func(t *testing.T) {
		// given
		remoteEnvServiceAPI := &remoteenv.ServiceAPI{}
		remoteEnvService := remoteenv.Service{API: remoteEnvServiceAPI}

		serviceRepository := new(remoteenvmocks.ServiceRepository)
		serviceRepository.On("Get", "uuid-1").Return(remoteEnvService, nil)

		serviceAPIService := new(serviceapimocks.Service)
		serviceAPIService.On("Read", remoteEnvServiceAPI).Return(nil, apperrors.Internal("some error"))

		service := NewServiceDefinitionService(serviceAPIService, serviceRepository)

		// when
		result, err := service.GetAPI("uuid-1")

		// then
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		assert.Contains(t, err.Error(), "some error")
	})
}
