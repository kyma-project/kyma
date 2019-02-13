package externalapi

import (
	"context"
	"encoding/json"
	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMngmtInfoHandler_GetCSRInfo(t *testing.T) {
	host := "connector-service.kyma.cx"

	t.Run("should successfully get management info urls for application", func(t *testing.T) {
		//given
		url := "/v1/applications/management/info"

		expectedMetadataURL := "https://metadata.base.path/application/v1/metadata/services"
		expectedEventsURL := "https://events.base.path/application/v1/events"
		expectedRenewalsURL := "https://connector-service.kyma.cx/v1/applications/certificates/renewals"

		extClientCtx := &clientcontext.ExtendedApplicationContext{
			ApplicationContext: clientcontext.ApplicationContext{},
			RuntimeURLs: clientcontext.RuntimeURLs{
				MetadataURL: "https://metadata.base.path/application/v1/metadata/services",
				EventsURL:   "https://events.base.path/application/v1/events",
			},
		}

		connectorClientExtractor := func(ctx context.Context) (clientcontext.ConnectorClientContext, apperrors.AppError) {
			return *extClientCtx, nil
		}

		req, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err)

		infoHandler := NewManagementInfoHandler(connectorClientExtractor, host)

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
		url := "/v1/runtimes/management/info"

		expectedRenewalsURL := "https://connector-service.kyma.cx/v1/applications/certificates/renewals"

		clusterCtx := &clientcontext.ClusterContext{}

		connectorClientExtractor := func(ctx context.Context) (clientcontext.ConnectorClientContext, apperrors.AppError) {
			return *clusterCtx, nil
		}

		req, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err)

		infoHandler := NewManagementInfoHandler(connectorClientExtractor, host)

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
}
