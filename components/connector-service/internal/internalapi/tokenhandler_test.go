package internalapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/httperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/tokens"
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

	tokenParams := tokens.ApplicationTokenParams{
		ClusterTokenParams: tokens.ClusterTokenParams{
			Tenant: tenant,
			Group:  group,
		},
		Application: appName,
	}

	tokenParamsParser := func(ctx context.Context) (tokens.TokenParams, apperrors.AppError) {
		assert.Equal(t, dummyCtxValue, ctx.Value(dummyCtxKey))
		return tokenParams, nil
	}

	ctx := context.WithValue(context.Background(), dummyCtxKey, dummyCtxValue)

	t.Run("should create token", func(t *testing.T) {
		// given
		csrURL := "domain.local/v1/application/csr/info"
		expectedTokenResponse := tokenResponse{
			URL:   fmt.Sprintf("https://%s?token=%s", csrURL, token),
			Token: token,
		}

		tokenService := &mocks.Service{}
		tokenService.On("Save", tokenParams).Return(token, nil)

		tokenHandler := NewTokenHandler(tokenService, csrURL, tokenParamsParser)

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
		tokenService.AssertExpectations(t)
	})

	t.Run("should return 500 when failed to parse context", func(t *testing.T) {
		// given
		errorParamsParser := func(ctx context.Context) (tokens.TokenParams, apperrors.AppError) {
			return nil, apperrors.Internal("error")
		}

		tokenHandler := NewTokenHandler(nil, "", errorParamsParser)

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
		tokenService := &mocks.Service{}
		tokenService.On("Save", tokenParams).Return("", apperrors.Internal("error"))

		tokenHandler := NewTokenHandler(tokenService, "", tokenParamsParser)

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
