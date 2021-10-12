package internalapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-project/kyma/components/connector-service/internal/certificates"

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

type dummyClientCertsContext struct {
	clientcontext.ClientContextService
}

func (ctx dummyClientCertsContext) GetSubject() certificates.CSRSubject {
	return certificates.CSRSubject{}
}

func (ctx dummyClientCertsContext) ClientContext() clientcontext.ClientContextService {
	return ctx.ClientContextService
}

func TestTokenHandler_CreateToken(t *testing.T) {

	clusterContext := clientcontext.ClusterContext{
		Tenant: tenant,
		Group:  group,
	}

	clusterClientContext := dummyClientCertsContext{clusterContext}

	applicationContext := clientcontext.ApplicationContext{
		ClusterContext: clusterContext,
		Application:    appName,
	}

	applicationClientContext := dummyClientCertsContext{applicationContext}

	connectorClientExtractor := func(ctx context.Context) (clientcontext.ClientCertContextService, apperrors.AppError) {
		assert.Equal(t, dummyCtxValue, ctx.Value(dummyCtxKey))
		return applicationClientContext, nil
	}

	ctx := context.WithValue(context.Background(), dummyCtxKey, dummyCtxValue)

	t.Run("should create token", func(t *testing.T) {
		// given
		csrURL := "domain.local/v1/application/csr/info"
		expectedTokenResponse := tokenResponse{
			URL:   fmt.Sprintf(TokenURLFormat, csrURL, token),
			Token: token,
		}

		tokenCreator := &mocks.Creator{}
		tokenCreator.On("Save", applicationContext).Return(token, nil)

		tokenHandler := NewTokenHandler(tokenCreator, csrURL, connectorClientExtractor)

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
		tokenCreator.AssertExpectations(t)
	})

	t.Run("should create token for cluster context", func(t *testing.T) {
		// given
		connectorClientExtractor := func(ctx context.Context) (clientcontext.ClientCertContextService, apperrors.AppError) {
			assert.Equal(t, dummyCtxValue, ctx.Value(dummyCtxKey))
			return clusterClientContext, nil
		}

		csrURL := "domain.local/v1/application/csr/info"
		expectedTokenResponse := tokenResponse{
			URL:   fmt.Sprintf(TokenURLFormat, csrURL, token),
			Token: token,
		}

		tokenCreator := &mocks.Creator{}
		tokenCreator.On("Save", clusterContext).Return(token, nil)

		tokenHandler := NewTokenHandler(tokenCreator, csrURL, connectorClientExtractor)

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
		tokenCreator.AssertExpectations(t)
	})

	t.Run("should return 500 when failed to parse context", func(t *testing.T) {
		// given
		errorExtractor := func(ctx context.Context) (clientcontext.ClientCertContextService, apperrors.AppError) {
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
		tokenCreator := &mocks.Creator{}
		tokenCreator.On("Save", applicationContext).Return("", apperrors.Internal("error"))

		tokenHandler := NewTokenHandler(tokenCreator, "", connectorClientExtractor)

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
