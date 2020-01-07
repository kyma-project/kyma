package specification

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/rafter/clusterassetgroup"

	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/model"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/rafter/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	serviceId                    = "1234"
	gatewayUrl                   = "http://1234.io"
	defaultSpecRequestTimeout    = 5
	defaultSpecRequestSkipVerify = true
)

var (
	baseApiSpec   = []byte("{\"api\":\"spec\"}")
	baseEventSpec = []byte("{\"event\":\"spec\"}")
	baseDocs      = []byte("{\"baseDocs\":\"baseDocs\"}")

	swaggerApiSpec      = []byte("{\"swagger\":\"2.0\"}")
	modifiedSwaggerSpec = []byte("{\"schemes\":[\"http\"],\"swagger\":\"2.0\",\"host\":\"1234.io\",\"paths\":null}")
)

func TestSpecService_PutSpec(t *testing.T) {

	t.Run("should save inline spec", func(t *testing.T) {
		// given
		serviceDef := defaultServiceDefWithAPI(&model.API{Spec: baseApiSpec, ApiType: ""})

		rafterSvc := &mocks.Service{}
		rafterSvc.On("Put", serviceId, clusterassetgroup.OpenApiType, baseDocs, baseApiSpec, baseEventSpec).Return(nil)

		specService := NewSpecService(rafterSvc, defaultSpecRequestTimeout, defaultSpecRequestSkipVerify)

		// when
		err := specService.PutSpec(serviceDef, gatewayUrl)

		// then
		require.NoError(t, err)
		rafterSvc.AssertExpectations(t)
	})

	t.Run("should modify and save inline swagger spec", func(t *testing.T) {
		// given
		serviceDef := defaultServiceDefWithAPI(&model.API{Spec: swaggerApiSpec})

		rafterSvc := &mocks.Service{}
		rafterSvc.On("Put", serviceId, clusterassetgroup.OpenApiType, baseDocs, modifiedSwaggerSpec, baseEventSpec).Return(nil)

		specService := NewSpecService(rafterSvc, defaultSpecRequestTimeout, defaultSpecRequestSkipVerify)

		// when
		err := specService.PutSpec(serviceDef, gatewayUrl)

		// then
		require.NoError(t, err)
		rafterSvc.AssertExpectations(t)
	})

	t.Run("should not modify spec if ApiType set to oData", func(t *testing.T) {
		// given
		serviceDef := defaultServiceDefWithAPI(&model.API{Spec: swaggerApiSpec, ApiType: oDataSpecType})

		rafterSvc := &mocks.Service{}
		rafterSvc.On("Put", serviceId, clusterassetgroup.ODataApiType, baseDocs, swaggerApiSpec, baseEventSpec).Return(nil)

		specService := NewSpecService(rafterSvc, defaultSpecRequestTimeout, defaultSpecRequestSkipVerify)

		// when
		err := specService.PutSpec(serviceDef, gatewayUrl)

		// then
		require.NoError(t, err)
		rafterSvc.AssertExpectations(t)
	})

	t.Run("should fetch and save spec", func(t *testing.T) {
		// given
		specServer := newSpecServer(baseApiSpec, func(req *http.Request) {
			assert.Equal(t, http.MethodGet, req.Method)
			assert.Equal(t, "/path", req.URL.Path)
		})

		serviceDef := defaultServiceDefWithAPI(&model.API{SpecificationUrl: specServer.URL + "/path"})

		rafterSvc := &mocks.Service{}
		rafterSvc.On("Put", serviceId, clusterassetgroup.OpenApiType, baseDocs, baseApiSpec, baseEventSpec).Return(nil)

		specService := NewSpecService(rafterSvc, defaultSpecRequestTimeout, defaultSpecRequestSkipVerify)

		// when
		err := specService.PutSpec(serviceDef, gatewayUrl)

		// then
		require.NoError(t, err)
		rafterSvc.AssertExpectations(t)
	})

	t.Run("should fetch spec if spec equal to null", func(t *testing.T) {
		// given
		specServer := newSpecServer(baseApiSpec, func(req *http.Request) {
			assert.Equal(t, http.MethodGet, req.Method)
			assert.Equal(t, "/path", req.URL.Path)
		})

		serviceDef := defaultServiceDefWithAPI(&model.API{Spec: []byte("null"), SpecificationUrl: specServer.URL + "/path"})

		rafterSvc := &mocks.Service{}
		rafterSvc.On("Put", serviceId, clusterassetgroup.OpenApiType, baseDocs, baseApiSpec, baseEventSpec).Return(nil)

		specService := NewSpecService(rafterSvc, defaultSpecRequestTimeout, defaultSpecRequestSkipVerify)

		// when
		err := specService.PutSpec(serviceDef, gatewayUrl)

		// then
		require.NoError(t, err)
		rafterSvc.AssertExpectations(t)
	})

	t.Run("should fetch, modify and save spec", func(t *testing.T) {
		// given
		specServer := newSpecServer(swaggerApiSpec, func(req *http.Request) {
			assert.Equal(t, http.MethodGet, req.Method)
			assert.Equal(t, "/path", req.URL.Path)
		})

		serviceDef := defaultServiceDefWithAPI(&model.API{SpecificationUrl: specServer.URL + "/path"})

		rafterSvc := &mocks.Service{}
		rafterSvc.On("Put", serviceId, clusterassetgroup.OpenApiType, baseDocs, modifiedSwaggerSpec, baseEventSpec).Return(nil)

		specService := NewSpecService(rafterSvc, defaultSpecRequestTimeout, defaultSpecRequestSkipVerify)

		// when
		err := specService.PutSpec(serviceDef, gatewayUrl)

		// then
		require.NoError(t, err)
		rafterSvc.AssertExpectations(t)
	})

	t.Run("should return UpstreamServerCallFailed error when failed to fetch spec", func(t *testing.T) {
		// given
		specServer := new404server()

		serviceDef := defaultServiceDefWithAPI(&model.API{SpecificationUrl: specServer.URL})

		rafterSvc := &mocks.Service{}

		specService := NewSpecService(rafterSvc, defaultSpecRequestTimeout, defaultSpecRequestSkipVerify)

		// when
		err := specService.PutSpec(serviceDef, gatewayUrl)

		// then
		require.Error(t, err)
		assert.Equal(t, apperrors.CodeUpstreamServerCallFailed, err.Code())
	})

	t.Run("should fetch OData spec from /$metadata when no url provided", func(t *testing.T) {
		// given
		specServer := newSpecServer(baseApiSpec, func(req *http.Request) {
			assert.Equal(t, http.MethodGet, req.Method)
			assert.Equal(t, "/$metadata", req.URL.Path)
		})

		serviceDef := defaultServiceDefWithAPI(&model.API{TargetUrl: specServer.URL, ApiType: oDataSpecType})

		rafterSvc := &mocks.Service{}
		rafterSvc.On("Put", serviceId, clusterassetgroup.ODataApiType, baseDocs, baseApiSpec, baseEventSpec).Return(nil)

		specService := NewSpecService(rafterSvc, defaultSpecRequestTimeout, defaultSpecRequestSkipVerify)

		// when
		err := specService.PutSpec(serviceDef, gatewayUrl)

		// then
		require.NoError(t, err)
		rafterSvc.AssertExpectations(t)
	})

	t.Run("should save empty spec when no spec url provided and api type is not OData", func(t *testing.T) {
		// given
		serviceDef := defaultServiceDefWithAPI(&model.API{})

		rafterSvc := &mocks.Service{}
		rafterSvc.On("Put", serviceId, clusterassetgroup.OpenApiType, baseDocs, []byte(nil), baseEventSpec).Return(nil)

		specService := NewSpecService(rafterSvc, defaultSpecRequestTimeout, defaultSpecRequestSkipVerify)

		// when
		err := specService.PutSpec(serviceDef, gatewayUrl)

		// then
		require.NoError(t, err)
		rafterSvc.AssertExpectations(t)
	})

	t.Run("should skip processing api spec if api not specified", func(t *testing.T) {
		// given
		serviceDef := defaultServiceDefWithAPI(nil)

		assetRafterSvc := &mocks.Service{}
		assetRafterSvc.On("Put", serviceId, clusterassetgroup.NoneApiType, baseDocs, []byte(nil), baseEventSpec).Return(nil)

		specService := NewSpecService(assetRafterSvc, defaultSpecRequestTimeout, defaultSpecRequestSkipVerify)

		// when
		err := specService.PutSpec(serviceDef, gatewayUrl)

		// then
		require.NoError(t, err)
		assetRafterSvc.AssertExpectations(t)
	})

	t.Run("should return error when saving to Minio failed", func(t *testing.T) {
		// given
		serviceDef := defaultServiceDefWithAPI(&model.API{Spec: baseApiSpec})

		rafterSvc := &mocks.Service{}
		rafterSvc.On("Put", serviceId, clusterassetgroup.OpenApiType, baseDocs, baseApiSpec, baseEventSpec).Return(apperrors.Internal("Error"))

		specService := NewSpecService(rafterSvc, defaultSpecRequestTimeout, defaultSpecRequestSkipVerify)

		// when
		err := specService.PutSpec(serviceDef, gatewayUrl)

		// then
		require.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		rafterSvc.AssertExpectations(t)
	})
}

func TestSpecService_GetSpec(t *testing.T) {

	t.Run("should get spec", func(t *testing.T) {
		// given
		rafterSvc := &mocks.Service{}
		rafterSvc.On("Get", serviceId).Return(baseDocs, baseApiSpec, baseEventSpec, nil)

		specService := NewSpecService(rafterSvc, defaultSpecRequestTimeout, defaultSpecRequestSkipVerify)

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
		rafterSvc := &mocks.Service{}
		rafterSvc.On("Get", serviceId).Return(nil, nil, nil, apperrors.Internal("Error"))

		specService := NewSpecService(rafterSvc, defaultSpecRequestTimeout, defaultSpecRequestSkipVerify)

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
		rafterSvc := &mocks.Service{}
		rafterSvc.On("Remove", serviceId).Return(nil)

		specService := NewSpecService(rafterSvc, defaultSpecRequestTimeout, defaultSpecRequestSkipVerify)

		// when
		err := specService.RemoveSpec(serviceId)

		// then
		require.NoError(t, err)
	})

	t.Run("should return error when failed to remove spec", func(t *testing.T) {
		// given
		rafterSvc := &mocks.Service{}
		rafterSvc.On("Remove", serviceId).Return(apperrors.Internal("Error"))

		specService := NewSpecService(rafterSvc, defaultSpecRequestTimeout, defaultSpecRequestSkipVerify)

		// when
		err := specService.RemoveSpec(serviceId)

		// then
		require.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
	})
}

func defaultServiceDefWithAPI(api *model.API) *model.ServiceDefinition {
	return &model.ServiceDefinition{
		ID:            serviceId,
		Identifier:    "identifier",
		Provider:      "provider",
		Description:   "description",
		Api:           api,
		Events:        &model.Events{Spec: baseEventSpec},
		Documentation: baseDocs,
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
