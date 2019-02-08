package internalapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/httperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/tokens/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	appName       = "appName"
	token         = "token"
	group         = "group"
	tenant        = "tenant"
	url           = "/v1/applications/tokens"
	dummyCtxKey   = "dummyKey"
	dummyCtxValue = "dummyValue"
)

func TestTokenHandler_CreateToken(t *testing.T) {

	clusterContext := clientcontext.ClusterContext{
		Tenant: tenant,
		Group:  group,
	}

	applicationContext := clientcontext.ApplicationContext{
		ClusterContext: clusterContext,
		Application:    appName,
	}

	connectorClientExtractor := func(ctx context.Context) (clientcontext.ConnectorClientContext, apperrors.AppError) {
		assert.Equal(t, dummyCtxValue, ctx.Value(dummyCtxKey))
		return applicationContext, nil
	}

	ctx := context.WithValue(context.Background(), dummyCtxKey, dummyCtxValue)

	t.Run("should create token", func(t *testing.T) {
		// given
		csrURL := "domain.local/v1/application/csr/info"
		expectedTokenResponse := tokenResponse{
			URL:   fmt.Sprintf(TokenURLFormat, csrURL, token),
			Token: token,
		}

		tokenManager := &mocks.Manager{}
		tokenManager.On("Save", applicationContext).Return(token, nil)

		tokenHandler := NewTokenHandler(tokenManager, csrURL, connectorClientExtractor)

		req, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		// when
		tokenHandler.CreateToken(rr, req)

		// then
		responseBody, err := ioutil.ReadAll(rr.Body)
		require.NoError(t, err)

		var tokenResponse tokenResponse
		err = json.Unmarshal(responseBody, &tokenResponse)
		require.NoError(t, err)

		assert.Equal(t, http.StatusCreated, rr.Code)
		assert.EqualValues(t, expectedTokenResponse, tokenResponse)
		tokenManager.AssertExpectations(t)
	})

	t.Run("should create token for cluster context", func(t *testing.T) {
		// given
		connectorClientExtractor := func(ctx context.Context) (clientcontext.ConnectorClientContext, apperrors.AppError) {
			assert.Equal(t, dummyCtxValue, ctx.Value(dummyCtxKey))
			return clusterContext, nil
		}

		csrURL := "domain.local/v1/application/csr/info"
		expectedTokenResponse := tokenResponse{
			URL:   fmt.Sprintf(TokenURLFormat, csrURL, token),
			Token: token,
		}

		tokenManager := &mocks.Manager{}
		tokenManager.On("Save", clusterContext).Return(token, nil)

		tokenHandler := NewTokenHandler(tokenManager, csrURL, connectorClientExtractor)

		req, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		// when
		tokenHandler.CreateToken(rr, req)

		// then
		responseBody, err := ioutil.ReadAll(rr.Body)
		require.NoError(t, err)

		var tokenResponse tokenResponse
		err = json.Unmarshal(responseBody, &tokenResponse)
		require.NoError(t, err)

		assert.Equal(t, http.StatusCreated, rr.Code)
		assert.EqualValues(t, expectedTokenResponse, tokenResponse)
		tokenManager.AssertExpectations(t)
	})

	t.Run("should return 500 when failed to parse context", func(t *testing.T) {
		// given
		errorExtractor := func(ctx context.Context) (clientcontext.ConnectorClientContext, apperrors.AppError) {
			return nil, apperrors.Internal("error")
		}

		tokenHandler := NewTokenHandler(nil, "", errorExtractor)

		req, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		// when
		tokenHandler.CreateToken(rr, req)

		// then
		responseBody, err := ioutil.ReadAll(rr.Body)
		require.NoError(t, err)

		var errorResponse httperrors.ErrorResponse
		err = json.Unmarshal(responseBody, &errorResponse)
		require.NoError(t, err)

		assert.Equal(t, http.StatusInternalServerError, errorResponse.Code)
		assert.Equal(t, "error", errorResponse.Error)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("should return 500 when failed to save", func(t *testing.T) {
		// given
		tokenManager := &mocks.Manager{}
		tokenManager.On("Save", applicationContext).Return("", apperrors.Internal("error"))

		tokenHandler := NewTokenHandler(tokenManager, "", connectorClientExtractor)

		req, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		// when
		tokenHandler.CreateToken(rr, req)

		// then
		responseBody, err := ioutil.ReadAll(rr.Body)
		require.NoError(t, err)

		var errorResponse httperrors.ErrorResponse
		err = json.Unmarshal(responseBody, &errorResponse)
		require.NoError(t, err)

		assert.Equal(t, http.StatusInternalServerError, errorResponse.Code)
		assert.Equal(t, "error", errorResponse.Error)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}
