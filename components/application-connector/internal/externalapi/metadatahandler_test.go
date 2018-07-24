package externalapi

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/application-connector/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-connector/internal/httperrors"
	"github.com/kyma-project/kyma/components/application-connector/internal/metadata"
	metadataMock "github.com/kyma-project/kyma/components/application-connector/internal/metadata/mocks"
	"github.com/kyma-project/kyma/components/application-connector/internal/metadata/serviceapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type testSpec struct {
	Name string
}

var (
	apiRawSpec       = compact([]byte("{\"name\":\"api\"}"))
	eventsRawSpec    = compact([]byte("{\"name\":\"events\"}"))
	documentationRaw = compact([]byte("{\"displayName\":\"documentation name\",\"description\":\"documentation description\",\"type\":\"documentation type\",\"docs\":[{\"title\":\"doc title\",\"type\":\"doc type\",\"source\":\"doc source\"}]}"))
	apiSpec          = testSpec{Name: "api"}
	eventsSpec       = testSpec{Name: "events"}
)

func TestMetadataHandler_CreateService(t *testing.T) {
	t.Run("should create a service", func(t *testing.T) {

		// given
		serviceDetails := ServiceDetails{
			Name:        "service name",
			Provider:    "service provider",
			Description: "service description",
			Api: &API{
				TargetUrl: "http://service.com",
				Credentials: &Credentials{
					Oauth: Oauth{
						URL:          "http://oauth.com",
						ClientID:     "clientId",
						ClientSecret: "clientSecret",
					},
				},
				Spec: apiRawSpec,
			},
			Events: &Events{
				Spec: eventsRawSpec,
			},
			Documentation: &Documentation{
				DisplayName: "documentation name",
				Description: "documentation description",
				Type:        "documentation type",
				Docs:        []DocsObject{{Title: "doc title", Type: "doc type", Source: "doc source"}},
			},
		}

		serviceDefinition := &metadata.ServiceDefinition{
			Name:        "service name",
			Provider:    "service provider",
			Description: "service description",
			Api: &serviceapi.API{
				TargetUrl: "http://service.com",
				Credentials: &serviceapi.Credentials{
					Oauth: serviceapi.Oauth{
						URL:          "http://oauth.com",
						ClientID:     "clientId",
						ClientSecret: "clientSecret",
					},
				},
				Spec: apiRawSpec,
			},
			Events: &metadata.Events{
				Spec: eventsRawSpec,
			},
			Documentation: documentationRaw,
		}

		validator := ServiceDetailsValidatorFunc(func(sd ServiceDetails) apperrors.AppError {
			return nil
		})
		serviceDefinitionService := &metadataMock.ServiceDefinitionService{}
		serviceDefinitionService.On("Create", "re", serviceDefinition).Return("1", nil)

		metadataHandler := NewMetadataHandler(validator, serviceDefinitionService)

		serviceDetailsData, err := json.Marshal(serviceDetails)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, "/re/v1/metadata/services", bytes.NewReader(serviceDetailsData))
		require.NoError(t, err)

		req = mux.SetURLVars(req, map[string]string{"remoteEnvironment": "re"})
		rr := httptest.NewRecorder()

		// when
		metadataHandler.CreateService(rr, req)

		// then
		responseBody, err := ioutil.ReadAll(rr.Body)
		require.NoError(t, err)

		var postResponse CreateServiceResponse
		json.Unmarshal(responseBody, &postResponse)

		require.NoError(t, err)
		assert.Equal(t, "1", postResponse.ID)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("should create a service with API without credentials", func(t *testing.T) {

		// given
		serviceDetails := ServiceDetails{
			Name:        "service name",
			Provider:    "service provider",
			Description: "service description",
			Api: &API{
				TargetUrl: "http://service.com",
			},
		}

		serviceDefinition := &metadata.ServiceDefinition{
			Name:        "service name",
			Provider:    "service provider",
			Description: "service description",
			Api: &serviceapi.API{
				TargetUrl: "http://service.com",
			},
		}

		validator := ServiceDetailsValidatorFunc(func(sd ServiceDetails) apperrors.AppError {
			return nil
		})
		serviceDefinitionService := &metadataMock.ServiceDefinitionService{}
		serviceDefinitionService.On("Create", "re", serviceDefinition).Return("1", nil)

		metadataHandler := NewMetadataHandler(validator, serviceDefinitionService)

		serviceDetailsData, err := json.Marshal(serviceDetails)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, "/re/v1/metadata/services", bytes.NewReader(serviceDetailsData))
		require.NoError(t, err)

		req = mux.SetURLVars(req, map[string]string{"remoteEnvironment": "re"})
		rr := httptest.NewRecorder()

		// when
		metadataHandler.CreateService(rr, req)

		// then
		responseBody, err := ioutil.ReadAll(rr.Body)
		require.NoError(t, err)

		var postResponse CreateServiceResponse
		json.Unmarshal(responseBody, &postResponse)

		require.NoError(t, err)
		assert.Equal(t, "1", postResponse.ID)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("should return 400 when validation fails", func(t *testing.T) {

		// given
		serviceDetails := ServiceDetails{}

		validator := ServiceDetailsValidatorFunc(func(sd ServiceDetails) apperrors.AppError {
			return apperrors.WrongInput("failed")
		})
		serviceDefinitionService := &metadataMock.ServiceDefinitionService{}

		metadataHandler := NewMetadataHandler(validator, serviceDefinitionService)

		serviceDetailsData, err := json.Marshal(serviceDetails)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, "/re/v1/metadata/services", bytes.NewReader(serviceDetailsData))
		require.NoError(t, err)

		req = mux.SetURLVars(req, map[string]string{"remoteEnvironment": "re"})
		rr := httptest.NewRecorder()

		// when
		metadataHandler.CreateService(rr, req)

		// then
		responseBody, err := ioutil.ReadAll(rr.Body)
		require.NoError(t, err)

		var errorResponse httperrors.ErrorResponse
		json.Unmarshal(responseBody, &errorResponse)

		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, errorResponse.Code)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		serviceDefinitionService.AssertNotCalled(t, "Create", "re", mock.AnythingOfType("*metadata.ServiceDefinition"))
	})

	t.Run("should handle internal errors", func(t *testing.T) {

		// given
		serviceDetails := ServiceDetails{
			Name:        "service name",
			Provider:    "service provider",
			Description: "service description",
			Api: &API{
				TargetUrl: "http://service.com",
			},
		}

		validator := ServiceDetailsValidatorFunc(func(sd ServiceDetails) apperrors.AppError {
			return nil
		})
		serviceDefinitionService := &metadataMock.ServiceDefinitionService{}
		serviceDefinitionService.On("Create", "re", mock.AnythingOfType("*metadata.ServiceDefinition")).Return(
			"", apperrors.Internal(""))

		metadataHandler := NewMetadataHandler(validator, serviceDefinitionService)

		serviceDetailsData, err := json.Marshal(serviceDetails)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, "/re/v1/metadata/services", bytes.NewReader(serviceDetailsData))
		require.NoError(t, err)

		req = mux.SetURLVars(req, map[string]string{"remoteEnvironment": "re"})
		rr := httptest.NewRecorder()

		// when
		metadataHandler.CreateService(rr, req)

		// then
		responseBody, err := ioutil.ReadAll(rr.Body)
		require.NoError(t, err)

		var errorResponse httperrors.ErrorResponse
		json.Unmarshal(responseBody, &errorResponse)

		require.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, errorResponse.Code)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestMetadataHandler_GetService(t *testing.T) {
	t.Run("should return requested service", func(t *testing.T) {
		// given
		serviceDefinition := metadata.ServiceDefinition{
			Name:        "service name",
			Provider:    "service provider",
			Description: "service description",
			Api: &serviceapi.API{
				TargetUrl: "http://service.com",
				Credentials: &serviceapi.Credentials{
					Oauth: serviceapi.Oauth{
						URL:          "http://oauth.com",
						ClientID:     "clientId",
						ClientSecret: "clientSecret",
					},
				},
				Spec: apiRawSpec,
			},
			Events: &metadata.Events{
				Spec: eventsRawSpec,
			},
			Documentation: documentationRaw,
		}

		serviceDefinitionService := &metadataMock.ServiceDefinitionService{}
		serviceDefinitionService.On("GetByID", "re", "123456").Return(serviceDefinition, nil)

		metadataHandler := NewMetadataHandler(nil, serviceDefinitionService)

		req, err := http.NewRequest(http.MethodGet, "/re/v1/metadata/services/123456", nil)
		require.NoError(t, err)

		req = mux.SetURLVars(req, map[string]string{"remoteEnvironment": "re", "serviceId": "123456"})
		rr := httptest.NewRecorder()

		// when
		metadataHandler.GetService(rr, req)

		// then
		responseBody, err := ioutil.ReadAll(rr.Body)
		require.NoError(t, err)

		var serviceDetails ServiceDetails
		json.Unmarshal(responseBody, &serviceDetails)

		require.NoError(t, err)
		serviceDefinitionService.AssertCalled(t, "GetByID", "re", "123456")
		assert.Equal(t, "service name", serviceDetails.Name)
		assert.Equal(t, "service provider", serviceDetails.Provider)
		assert.Equal(t, "service description", serviceDetails.Description)
		assert.Equal(t, "http://service.com", serviceDetails.Api.TargetUrl)
		assert.Equal(t, "http://oauth.com", serviceDetails.Api.Credentials.Oauth.URL)
		assert.Equal(t, stars, serviceDetails.Api.Credentials.Oauth.ClientID)
		assert.Equal(t, stars, serviceDetails.Api.Credentials.Oauth.ClientSecret)
		assert.Equal(t, apiSpec, raw2Json(t, serviceDetails.Api.Spec))
		assert.Equal(t, eventsSpec, raw2Json(t, serviceDetails.Events.Spec))
		assert.Equal(t, "documentation name", serviceDetails.Documentation.DisplayName)
		assert.Equal(t, "documentation description", serviceDetails.Documentation.Description)
		assert.Equal(t, "documentation type", serviceDetails.Documentation.Type)
		assert.Len(t, serviceDetails.Documentation.Docs, 1)
		assert.Equal(t, DocsObject{Title: "doc title", Type: "doc type", Source: "doc source"}, serviceDetails.Documentation.Docs[0])

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("should return requested service only with API", func(t *testing.T) {
		// given
		serviceDefinition := metadata.ServiceDefinition{
			Name:        "service name",
			Provider:    "service provider",
			Description: "service description",
			Api: &serviceapi.API{
				TargetUrl: "http://service.com",
				Credentials: &serviceapi.Credentials{
					Oauth: serviceapi.Oauth{
						URL:          "http://oauth.com",
						ClientID:     "clientId",
						ClientSecret: "clientSecret",
					},
				},
				Spec: apiRawSpec,
			},
		}

		serviceDefinitionService := &metadataMock.ServiceDefinitionService{}
		serviceDefinitionService.On("GetByID", "re", "123456").Return(serviceDefinition, nil)

		metadataHandler := NewMetadataHandler(nil, serviceDefinitionService)

		req, err := http.NewRequest(http.MethodGet, "/re/v1/metadata/services/123456", nil)
		require.NoError(t, err)

		req = mux.SetURLVars(req, map[string]string{"remoteEnvironment": "re", "serviceId": "123456"})
		rr := httptest.NewRecorder()

		// when
		metadataHandler.GetService(rr, req)

		// then
		responseBody, err := ioutil.ReadAll(rr.Body)
		require.NoError(t, err)

		var serviceDetails ServiceDetails
		json.Unmarshal(responseBody, &serviceDetails)

		require.NoError(t, err)
		serviceDefinitionService.AssertCalled(t, "GetByID", "re", "123456")
		assert.Equal(t, "service name", serviceDetails.Name)
		assert.Equal(t, "service provider", serviceDetails.Provider)
		assert.Equal(t, "service description", serviceDetails.Description)
		assert.Equal(t, "http://service.com", serviceDetails.Api.TargetUrl)
		assert.Equal(t, "http://oauth.com", serviceDetails.Api.Credentials.Oauth.URL)
		assert.Equal(t, stars, serviceDetails.Api.Credentials.Oauth.ClientID)
		assert.Equal(t, stars, serviceDetails.Api.Credentials.Oauth.ClientSecret)
		assert.Equal(t, apiSpec, raw2Json(t, serviceDetails.Api.Spec))
		assert.Nil(t, serviceDetails.Events)
		assert.Nil(t, serviceDetails.Documentation)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("should return requested service only with Events", func(t *testing.T) {
		// given
		serviceDefinition := metadata.ServiceDefinition{
			Name:        "service name",
			Provider:    "service provider",
			Description: "service description",
			Events: &metadata.Events{
				Spec: eventsRawSpec,
			},
		}

		serviceDefinitionService := &metadataMock.ServiceDefinitionService{}
		serviceDefinitionService.On("GetByID", "re", "123456").Return(serviceDefinition, nil)

		metadataHandler := NewMetadataHandler(nil, serviceDefinitionService)

		req, err := http.NewRequest(http.MethodGet, "/re/v1/metadata/services/123456", nil)
		require.NoError(t, err)

		req = mux.SetURLVars(req, map[string]string{"remoteEnvironment": "re", "serviceId": "123456"})
		rr := httptest.NewRecorder()

		// when
		metadataHandler.GetService(rr, req)

		// then
		responseBody, err := ioutil.ReadAll(rr.Body)
		require.NoError(t, err)

		var serviceDetails ServiceDetails
		json.Unmarshal(responseBody, &serviceDetails)

		require.NoError(t, err)
		serviceDefinitionService.AssertCalled(t, "GetByID", "re", "123456")
		assert.Equal(t, "service name", serviceDetails.Name)
		assert.Equal(t, "service provider", serviceDetails.Provider)
		assert.Equal(t, "service description", serviceDetails.Description)
		assert.Nil(t, serviceDetails.Api)
		assert.Equal(t, eventsSpec, raw2Json(t, serviceDetails.Events.Spec))
		assert.Nil(t, serviceDetails.Documentation)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("should return 404 when service was not found", func(t *testing.T) {

		// given
		serviceDefinitionService := &metadataMock.ServiceDefinitionService{}
		serviceDefinitionService.On("GetByID", "re", "654321").Return(
			metadata.ServiceDefinition{},
			apperrors.NotFound("Service with ID %d not found", 654321),
		)

		metadataHandler := NewMetadataHandler(nil, serviceDefinitionService)

		req, err := http.NewRequest(http.MethodGet, "/re/v1/metadata/services/654321", nil)
		require.NoError(t, err)

		req = mux.SetURLVars(req, map[string]string{"remoteEnvironment": "re", "serviceId": "654321"})
		rr := httptest.NewRecorder()

		// when
		metadataHandler.GetService(rr, req)

		// then
		require.NoError(t, err)
		serviceDefinitionService.AssertCalled(t, "GetByID", "re", "654321")
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestMetadataHandler_GetServices(t *testing.T) {
	t.Run("should return list of available services", func(t *testing.T) {
		// given
		serviceDefinitionService := &metadataMock.ServiceDefinitionService{}
		serviceDefinitionService.On("GetAll", "re").Return([]metadata.ServiceDefinition{{
			Name:        "service name",
			Provider:    "service provider",
			Description: "service description",
		}}, nil)

		metadataHandler := NewMetadataHandler(nil, serviceDefinitionService)

		req, err := http.NewRequest(http.MethodGet, "/re/v1/metadata/services", nil)
		require.NoError(t, err)
		req = mux.SetURLVars(req, map[string]string{"remoteEnvironment": "re"})
		rr := httptest.NewRecorder()

		// when
		metadataHandler.GetServices(rr, req)

		// then
		responseBody, err := ioutil.ReadAll(rr.Body)
		require.NoError(t, err)

		var services []Service
		json.Unmarshal(responseBody, &services)

		require.NoError(t, err)
		serviceDefinitionService.AssertExpectations(t)
		assert.Len(t, services, 1)
		assert.Equal(t, "service name", services[0].Name)
		assert.Equal(t, "service provider", services[0].Provider)
		assert.Equal(t, "service description", services[0].Description)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("should empty list when no services found", func(t *testing.T) {
		// given
		var empty []metadata.ServiceDefinition
		serviceDefinitionService := &metadataMock.ServiceDefinitionService{}
		serviceDefinitionService.On("GetAll", "re").Return(empty, nil)

		metadataHandler := NewMetadataHandler(nil, serviceDefinitionService)

		req, err := http.NewRequest(http.MethodGet, "/re/v1/metadata/services", nil)
		require.NoError(t, err)

		req = mux.SetURLVars(req, map[string]string{"remoteEnvironment": "re"})
		rr := httptest.NewRecorder()

		// when
		metadataHandler.GetServices(rr, req)

		// then
		responseBody, err := ioutil.ReadAll(rr.Body)
		require.NoError(t, err)

		var services []Service
		json.Unmarshal(responseBody, &services)

		require.NoError(t, err)
		serviceDefinitionService.AssertExpectations(t)
		assert.Len(t, services, 0)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("should handle internal errors", func(t *testing.T) {
		// given
		serviceDefinitionService := &metadataMock.ServiceDefinitionService{}
		serviceDefinitionService.On("GetAll", "re").Return(nil, apperrors.Internal(""))

		metadataHandler := NewMetadataHandler(nil, serviceDefinitionService)

		req, err := http.NewRequest(http.MethodGet, "/re/v1/metadata/services", nil)
		require.NoError(t, err)

		req = mux.SetURLVars(req, map[string]string{"remoteEnvironment": "re"})
		rr := httptest.NewRecorder()

		// when
		metadataHandler.GetServices(rr, req)

		// then
		responseBody, err := ioutil.ReadAll(rr.Body)
		require.NoError(t, err)

		var errorResponse httperrors.ErrorResponse
		json.Unmarshal(responseBody, &errorResponse)

		require.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, errorResponse.Code)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		serviceDefinitionService.AssertExpectations(t)
	})
}

func TestMetadataHandler_UpdateService(t *testing.T) {
	t.Run("should update a service", func(t *testing.T) {
		// given
		serviceDetails := ServiceDetails{
			Name:        "service name",
			Provider:    "service provider",
			Description: "service description",
			Api: &API{
				TargetUrl: "http://service.com",
				Credentials: &Credentials{
					Oauth: Oauth{
						URL:          "http://oauth.com",
						ClientID:     "clientId",
						ClientSecret: "clientSecret",
					},
				},
				Spec: apiRawSpec,
			},
			Events: &Events{
				Spec: eventsRawSpec,
			},
			Documentation: &Documentation{
				DisplayName: "documentation name",
				Description: "documentation description",
				Type:        "documentation type",
				Docs:        []DocsObject{{Title: "doc title", Type: "doc type", Source: "doc source"}},
			},
		}

		serviceDefinition := &metadata.ServiceDefinition{
			Name:        "service name",
			Provider:    "service provider",
			Description: "service description",
			Api: &serviceapi.API{
				TargetUrl: "http://service.com",
				Credentials: &serviceapi.Credentials{
					Oauth: serviceapi.Oauth{
						URL:          "http://oauth.com",
						ClientID:     "clientId",
						ClientSecret: "clientSecret",
					},
				},
				Spec: apiRawSpec,
			},
			Events: &metadata.Events{
				Spec: eventsRawSpec,
			},
			Documentation: documentationRaw,
		}

		validator := ServiceDetailsValidatorFunc(func(sd ServiceDetails) apperrors.AppError {
			return nil
		})
		serviceDefinitionService := &metadataMock.ServiceDefinitionService{}
		serviceDefinitionService.On("Update", "re", "1234", serviceDefinition).Return(nil)

		metadataHandler := NewMetadataHandler(validator, serviceDefinitionService)

		serviceDetailsData, err := json.Marshal(serviceDetails)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPut, "/re/v1/metadata/services/1234", bytes.NewReader(serviceDetailsData))
		require.NoError(t, err)

		req = mux.SetURLVars(req, map[string]string{"remoteEnvironment": "re", "serviceId": "1234"})
		rr := httptest.NewRecorder()

		// when
		metadataHandler.UpdateService(rr, req)

		// then
		responseBody, err := ioutil.ReadAll(rr.Body)
		require.NoError(t, err)

		var serviceDetailsResponse ServiceDetails
		json.Unmarshal(responseBody, &serviceDetailsResponse)

		assert.Equal(t, "service name", serviceDetailsResponse.Name)
		assert.Equal(t, "service provider", serviceDetailsResponse.Provider)
		assert.Equal(t, "service description", serviceDetailsResponse.Description)
		assert.Equal(t, "http://service.com", serviceDetailsResponse.Api.TargetUrl)
		assert.Equal(t, "http://oauth.com", serviceDetailsResponse.Api.Credentials.Oauth.URL)
		assert.Equal(t, stars, serviceDetailsResponse.Api.Credentials.Oauth.ClientID)
		assert.Equal(t, stars, serviceDetailsResponse.Api.Credentials.Oauth.ClientSecret)
		assert.Equal(t, apiSpec, raw2Json(t, serviceDetailsResponse.Api.Spec))
		assert.Equal(t, eventsSpec, raw2Json(t, serviceDetailsResponse.Events.Spec))
		assert.Equal(t, "documentation name", serviceDetailsResponse.Documentation.DisplayName)
		assert.Equal(t, "documentation description", serviceDetailsResponse.Documentation.Description)
		assert.Equal(t, "documentation type", serviceDetailsResponse.Documentation.Type)
		assert.Len(t, serviceDetailsResponse.Documentation.Docs, 1)
		assert.Equal(t, DocsObject{Title: "doc title", Type: "doc type", Source: "doc source"}, serviceDetailsResponse.Documentation.Docs[0])

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("should return 400 when validation fails", func(t *testing.T) {
		// given
		serviceDetails := ServiceDetails{}

		validator := ServiceDetailsValidatorFunc(func(sd ServiceDetails) apperrors.AppError {
			return apperrors.WrongInput("failed")
		})
		serviceDefinitionService := &metadataMock.ServiceDefinitionService{}

		metadataHandler := NewMetadataHandler(validator, serviceDefinitionService)

		serviceDetailsData, err := json.Marshal(serviceDetails)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPut, "/re/v1/metadata/services/1234", bytes.NewReader(serviceDetailsData))
		require.NoError(t, err)

		req = mux.SetURLVars(req, map[string]string{"remoteEnvironment": "re", "serviceId": "1234"})
		rr := httptest.NewRecorder()

		// when
		metadataHandler.UpdateService(rr, req)

		// then
		responseBody, err := ioutil.ReadAll(rr.Body)
		require.NoError(t, err)

		var errorResponse httperrors.ErrorResponse
		json.Unmarshal(responseBody, &errorResponse)

		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, errorResponse.Code)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		serviceDefinitionService.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
	})

	t.Run("should handle internal errors", func(t *testing.T) {
		// given
		serviceDetails := ServiceDetails{
			Name:        "service name",
			Provider:    "service provider",
			Description: "service description",
		}

		validator := ServiceDetailsValidatorFunc(func(sd ServiceDetails) apperrors.AppError {
			return nil
		})
		serviceDefinitionService := &metadataMock.ServiceDefinitionService{}
		serviceDefinitionService.On("Update", "re", "1234", mock.Anything).Return(apperrors.Internal(""))

		metadataHandler := NewMetadataHandler(validator, serviceDefinitionService)

		serviceDetailsData, err := json.Marshal(serviceDetails)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPut, "/re/v1/metadata/services/1234", bytes.NewReader(serviceDetailsData))
		require.NoError(t, err)

		req = mux.SetURLVars(req, map[string]string{"remoteEnvironment": "re", "serviceId": "1234"})
		rr := httptest.NewRecorder()

		// when
		metadataHandler.UpdateService(rr, req)

		// then
		responseBody, err := ioutil.ReadAll(rr.Body)
		require.NoError(t, err)

		var errorResponse httperrors.ErrorResponse
		json.Unmarshal(responseBody, &errorResponse)

		require.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, errorResponse.Code)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestMetadataHandler_DeleteService(t *testing.T) {
	t.Run("should delete service", func(t *testing.T) {
		// given
		serviceDefinitionService := &metadataMock.ServiceDefinitionService{}
		serviceDefinitionService.On("Delete", "re", "1234").Return(nil)

		metadataHandler := NewMetadataHandler(nil, serviceDefinitionService)

		req, err := http.NewRequest(http.MethodDelete, "/re/v1/metadata/services/1234", nil)
		require.NoError(t, err)

		req = mux.SetURLVars(req, map[string]string{"remoteEnvironment": "re", "serviceId": "1234"})
		rr := httptest.NewRecorder()

		// when
		metadataHandler.DeleteService(rr, req)

		// then
		assert.Equal(t, http.StatusNoContent, rr.Code)
	})

	t.Run("should handle errors when deleting service", func(t *testing.T) {
		// given
		serviceDefinitionService := &metadataMock.ServiceDefinitionService{}
		serviceDefinitionService.On("Delete", "re", "1234").Return(apperrors.Internal("error"))

		metadataHandler := NewMetadataHandler(nil, serviceDefinitionService)

		req, err := http.NewRequest(http.MethodDelete, "/re/v1/metadata/services/1234", nil)
		require.NoError(t, err)

		req = mux.SetURLVars(req, map[string]string{"remoteEnvironment": "re", "serviceId": "1234"})
		rr := httptest.NewRecorder()

		// when
		metadataHandler.DeleteService(rr, req)

		// then
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func raw2Json(t *testing.T, rawMsg json.RawMessage) testSpec {
	spec := testSpec{}
	err := json.Unmarshal(rawMsg, &spec)
	require.NoError(t, err)
	return spec
}
