package externalapi

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-project/kyma/components/connector-service/internal/httperrors"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManagementInfoHandler_GetManagementInfo(t *testing.T) {
	protectedBaseURL := "https://gateway.kyma.local/v1/applications"
	expectedRenewalsURL := "https://gateway.kyma.local/v1/applications/certificates/renewals"

	t.Run("should successfully get management info urls for application", func(t *testing.T) {
		//given
		expectedMetadataURL := "https://metadata.base.path/application/v1/metadata/services"
		expectedEventsURL := "https://events.base.path/application/v1/events"

		extClientCtx := &clientcontext.ExtendedApplicationContext{
			ApplicationContext: clientcontext.ApplicationContext{},
			RuntimeURLs: clientcontext.RuntimeURLs{
				MetadataURL: "https://metadata.base.path/application/v1/metadata/services",
				EventsURL:   "https://events.base.path/application/v1/events",
			},
		}

		connectorClientExtractor := func(ctx context.Context) (clientcontext.ContextServiceProvider, apperrors.AppError) {
			return *extClientCtx, nil
		}

		req, err := http.NewRequest(http.MethodGet, "/v1/applications/management/info", nil)
		require.NoError(t, err)

		infoHandler := NewManagementInfoHandler(connectorClientExtractor, protectedBaseURL)

		rr := httptest.NewRecorder()

		//when
		infoHandler.GetManagementInfo(rr, req)

		//then
		responseBody, err := ioutil.ReadAll(rr.Body)
		require.NoError(t, err)

		var infoResponse mgmtInfoReponse
		err = json.Unmarshal(responseBody, &infoResponse)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rr.Code)

		urls := infoResponse.URLs

		assert.Equal(t, expectedMetadataURL, urls.MetadataURL)
		assert.Equal(t, expectedEventsURL, urls.EventsURL)
		assert.Equal(t, expectedRenewalsURL, urls.RenewCertURL)
	})

	t.Run("should successfully get management info urls for runtime", func(t *testing.T) {
		//given
		contextServiceProvider := func(ctx context.Context) (clientcontext.ContextServiceProvider, apperrors.AppError) {
			return &clientcontext.ClusterContext{}, nil
		}

		req, err := http.NewRequest(http.MethodGet, "/v1/runtimes/management/info", nil)
		require.NoError(t, err)

		infoHandler := NewManagementInfoHandler(contextServiceProvider, protectedBaseURL)

		rr := httptest.NewRecorder()

		//when
		infoHandler.GetManagementInfo(rr, req)

		//then
		responseBody, err := ioutil.ReadAll(rr.Body)
		require.NoError(t, err)

		var infoResponse mgmtInfoReponse
		err = json.Unmarshal(responseBody, &infoResponse)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rr.Code)

		urls := infoResponse.URLs

		assert.Equal(t, expectedRenewalsURL, urls.RenewCertURL)
	})

	t.Run("should return 500 when failed to extract context", func(t *testing.T) {
		//given
		contextServiceProvider := func(ctx context.Context) (clientcontext.ContextServiceProvider, apperrors.AppError) {
			return nil, apperrors.Internal("error")
		}

		req, err := http.NewRequest(http.MethodGet, "/v1/applications/management/info", nil)
		require.NoError(t, err)

		infoHandler := NewManagementInfoHandler(contextServiceProvider, protectedBaseURL)

		rr := httptest.NewRecorder()

		//when
		infoHandler.GetManagementInfo(rr, req)

		//then
		responseBody, err := ioutil.ReadAll(rr.Body)
		require.NoError(t, err)

		var errorResposne httperrors.ErrorResponse
		err = json.Unmarshal(responseBody, &errorResposne)
		require.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}
