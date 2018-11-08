package specification

import (
	"testing"
	"github.com/kyma-project/kyma/components/metadata-service/internal/metadata/specification/minio/mocks"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/assert"
	"github.com/kyma-project/kyma/components/metadata-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/metadata-service/internal/metadata/serviceapi"
)

const (
	serviceId = "abcd"
	gatewayUrl = "http://re-1234.io"
)

var (
	baseApiSpec   = []byte("{\"api\":\"spec\"}")
	baseEventSpec = []byte("{\"event\":\"spec\"}")
	baseDocs      = []byte("{\"baseDocs\":\"baseDocs\"}")

	swaggerApiSpec   = []byte("{\"swagger\":\"2.0\"}")
	modifiedSwaggerSpec = []byte("{\"schemes\":[\"http\"],\"swagger\":\"2.0\",\"host\":\"re-1234.io\",\"paths\":null}")
)

func TestSpecService_SaveServiceSpecs(t *testing.T) {

	events := &Events{ Spec: baseEventSpec }

	t.Run("should save inline spec", func(t *testing.T) {
		// given
		specData := SpecData{
			Id: serviceId,
			API: &serviceapi.API{ Spec: baseApiSpec },
			Events: events,
			Docs: baseDocs,
			GatewayUrl: gatewayUrl,
		}

		minioSvc := &mocks.Service{}
		minioSvc.On("Put", serviceId, baseDocs, baseApiSpec, baseEventSpec).Return(nil)

		specService := NewSpecService(minioSvc)

		// when
		err := specService.SaveServiceSpecs(specData)

		// then
		require.NoError(t, err)
	})

	t.Run("should modify and save inline swagger spec", func(t *testing.T) {
		// given
		specData := SpecData{
			Id: serviceId,
			API: &serviceapi.API{ Spec: swaggerApiSpec },
			Events: events,
			Docs: baseDocs,
			GatewayUrl: gatewayUrl,
		}

		minioSvc := &mocks.Service{}
		minioSvc.On("Put", serviceId, baseDocs, modifiedSwaggerSpec, baseEventSpec).Return(nil)

		specService := NewSpecService(minioSvc)

		// when
		err := specService.SaveServiceSpecs(specData)

		// then
		require.NoError(t, err)
	})
}

func TestSpecService_GetSpec(t *testing.T) {

	t.Run("should get spec", func(t *testing.T) {
		// given
		minioSvc := &mocks.Service{}
		minioSvc.On("Get", serviceId).Return(baseDocs, baseApiSpec, baseEventSpec, nil)

		specService := NewSpecService(minioSvc)

		// when
		docs, apiSpec, eventsSpec, err := specService.GetSpec(serviceId)

		// then
		require.NoError(t, err)
		assert.Equal(t, baseApiSpec, apiSpec)
		assert.Equal(t, baseEventSpec, eventsSpec)
		assert.Equal(t, baseDocs, docs)
	})

	t.Run("should return error if getting speec failed", func(t *testing.T) {
		// given
		minioSvc := &mocks.Service{}
		minioSvc.On("Get", serviceId).Return(nil, nil, nil, apperrors.Internal("Error"))

		specService := NewSpecService(minioSvc)

		// when
		docs, apiSpec, eventsSpec, err := specService.GetSpec(serviceId)

		// then
		require.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		assert.Nil(t, docs)
		assert.Nil(t, apiSpec)
		assert.Nil(t, eventsSpec)
	})
}

func TestSpecService_RemoveSpec(t *testing.T) {

	t.Run("should delete spec", func(t *testing.T) {
		// given
		minioSvc := &mocks.Service{}
		minioSvc.On("Remove", serviceId).Return(nil)

		specService := NewSpecService(minioSvc)

		// when
		err := specService.RemoveSpec(serviceId)

		// then
		require.NoError(t, err)
	})

	t.Run("should return error when failed to remove spec", func(t *testing.T) {
		// given
		minioSvc := &mocks.Service{}
		minioSvc.On("Remove", serviceId).Return(apperrors.Internal("Error"))

		specService := NewSpecService(minioSvc)

		// when
		err := specService.RemoveSpec(serviceId)

		// then
		require.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
	})
}