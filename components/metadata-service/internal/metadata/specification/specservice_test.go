package specification

import (
	"github.com/kyma-project/kyma/components/metadata-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/metadata-service/internal/metadata/serviceapi"
	"github.com/kyma-project/kyma/components/metadata-service/internal/metadata/specification/minio/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	serviceId  = "1234"
	gatewayUrl = "http://re-1234.io"
)

var (
	baseApiSpec   = []byte("{\"api\":\"spec\"}")
	baseEventSpec = []byte("{\"event\":\"spec\"}")
	baseDocs      = []byte("{\"baseDocs\":\"baseDocs\"}")

	swaggerApiSpec      = []byte("{\"swagger\":\"2.0\"}")
	modifiedSwaggerSpec = []byte("{\"schemes\":[\"http\"],\"swagger\":\"2.0\",\"host\":\"re-1234.io\",\"paths\":null}")
)

func TestSpecService_SaveServiceSpecs(t *testing.T) {

	t.Run("should save inline spec", func(t *testing.T) {
		// given
		specData := defaultSpecDataWithAPI(&serviceapi.API{Spec: baseApiSpec})

		minioSvc := &mocks.Service{}
		minioSvc.On("Put", serviceId, baseDocs, baseApiSpec, baseEventSpec).Return(nil)

		specService := NewSpecService(minioSvc)

		// when
		err := specService.SaveServiceSpecs(specData)

		// then
		require.NoError(t, err)
		minioSvc.AssertExpectations(t)
	})

	t.Run("should modify and save inline swagger spec", func(t *testing.T) {
		// given
		specData := defaultSpecDataWithAPI(&serviceapi.API{Spec: swaggerApiSpec})

		minioSvc := &mocks.Service{}
		minioSvc.On("Put", serviceId, baseDocs, modifiedSwaggerSpec, baseEventSpec).Return(nil)

		specService := NewSpecService(minioSvc)

		// when
		err := specService.SaveServiceSpecs(specData)

		// then
		require.NoError(t, err)
		minioSvc.AssertExpectations(t)
	})

	t.Run("should fetch and save spec", func(t *testing.T) {
		// given
		specServer := newSpecServer(baseApiSpec, func(req *http.Request) {
			assert.Equal(t, http.MethodGet, req.Method)
			assert.Equal(t, "/path", req.URL.Path)
		})

		specData := defaultSpecDataWithAPI(&serviceapi.API{SpecUrl: specServer.URL + "/path"})

		minioSvc := &mocks.Service{}
		minioSvc.On("Put", serviceId, baseDocs, baseApiSpec, baseEventSpec).Return(nil)

		specService := NewSpecService(minioSvc)

		// when
		err := specService.SaveServiceSpecs(specData)

		// then
		require.NoError(t, err)
		minioSvc.AssertExpectations(t)
	})

	t.Run("should fetch, modify and save spec", func(t *testing.T) {
		// given
		specServer := newSpecServer(swaggerApiSpec, func(req *http.Request) {
			assert.Equal(t, http.MethodGet, req.Method)
			assert.Equal(t, "/path", req.URL.Path)
		})

		specData := defaultSpecDataWithAPI(&serviceapi.API{SpecUrl: specServer.URL + "/path"})

		minioSvc := &mocks.Service{}
		minioSvc.On("Put", serviceId, baseDocs, modifiedSwaggerSpec, baseEventSpec).Return(nil)

		specService := NewSpecService(minioSvc)

		// when
		err := specService.SaveServiceSpecs(specData)

		// then
		require.NoError(t, err)
		minioSvc.AssertExpectations(t)
	})

	t.Run("should return UpstreamServerCallFailed error when failed to fetch spec", func(t *testing.T) {
		// given
		specServer := new404server()

		specData := defaultSpecDataWithAPI(&serviceapi.API{SpecUrl: specServer.URL})

		minioSvc := &mocks.Service{}

		specService := NewSpecService(minioSvc)

		// when
		err := specService.SaveServiceSpecs(specData)

		// then
		require.Error(t, err)
		assert.Equal(t, apperrors.CodeUpstreamServerCallFailed, err.Code())
	})

	t.Run("should fetch spec from /$metadata when no url provided", func(t *testing.T) {
		// given
		specServer := newSpecServer(baseApiSpec, func(req *http.Request) {
			assert.Equal(t, http.MethodGet, req.Method)
			assert.Equal(t, "/$metadata", req.URL.Path)
		})

		specData := defaultSpecDataWithAPI(&serviceapi.API{TargetUrl: specServer.URL})

		minioSvc := &mocks.Service{}
		minioSvc.On("Put", serviceId, baseDocs, baseApiSpec, baseEventSpec).Return(nil)

		specService := NewSpecService(minioSvc)

		// when
		err := specService.SaveServiceSpecs(specData)

		// then
		require.NoError(t, err)
		minioSvc.AssertExpectations(t)
	})

	t.Run("should return error when saving to Minio failed", func(t *testing.T) {
		// given
		specData := defaultSpecDataWithAPI(&serviceapi.API{Spec: baseApiSpec})

		minioSvc := &mocks.Service{}
		minioSvc.On("Put", serviceId, baseDocs, baseApiSpec, baseEventSpec).Return(apperrors.Internal("Error"))

		specService := NewSpecService(minioSvc)

		// when
		err := specService.SaveServiceSpecs(specData)

		// then
		require.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		minioSvc.AssertExpectations(t)
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

func defaultSpecDataWithAPI(api *serviceapi.API) SpecData {
	return SpecData{
		Id:         serviceId,
		API:        api,
		Events:     &Events{Spec: baseEventSpec},
		Docs:       baseDocs,
		GatewayUrl: gatewayUrl,
	}
}

func newSpecServer(spec []byte, check func(req *http.Request)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		check(r)
		w.WriteHeader(http.StatusOK)
		w.Write(spec)
	}))
}

func new404server() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
}
