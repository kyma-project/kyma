package internalapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-project/kyma/components/connector-service/internal/api"
	"github.com/kyma-project/kyma/components/connector-service/internal/tokens"

	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/httperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/tokens/mocks"
	uuidmocks "github.com/kyma-project/kyma/components/connector-service/internal/uuid/mocks"
	vermocks "github.com/kyma-project/kyma/components/connector-service/internal/verification/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	host    = "host"
	appName = "appName"
	token   = "token"
	group   = "group"
	tenant  = "tenant"
)

func TestTokenHandler_CreateToken(t *testing.T) {

	tokenData := &tokens.TokenData{
		Group:  group,
		Tenant: tenant,
	}

	t.Run("should create token when app name specified", func(t *testing.T) {
		// given
		url := fmt.Sprintf("/v1/applications/%s/tokens", appName)

		expectedTokenResponse := api.TokenResponse{
			URL:   fmt.Sprintf("https://%s/v1/applications/%s/info?token=%s", host, appName, token),
			Token: token,
		}

		req, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		req = mux.SetURLVars(req, map[string]string{"appName": appName})

		uuidGenerator := &uuidmocks.Generator{}

		verificationService := &vermocks.Service{}
		verificationService.On("Verify", req, appName).Return(tokenData, nil)

		tokenService := &mocks.Service{}
		tokenService.On("CreateToken", appName, tokenData).Return(token, nil)

		tokenHandler := NewTokenHandler(verificationService, tokenService, host, uuidGenerator)

		// when
		tokenHandler.CreateToken(rr, req)

		// then
		responseBody, err := ioutil.ReadAll(rr.Body)
		require.NoError(t, err)

		var tokenResponse api.TokenResponse
		err = json.Unmarshal(responseBody, &tokenResponse)
		require.NoError(t, err)

		assert.Equal(t, http.StatusCreated, rr.Code)
		assert.EqualValues(t, expectedTokenResponse, tokenResponse)
	})

	t.Run("should create token when app not specified", func(t *testing.T) {
		// given
		url := "/v1/applications/tokens"

		expectedTokenResponse := api.TokenResponse{
			URL:   fmt.Sprintf("https://%s/v1/applications/%s/info?token=%s", host, appName, token),
			Token: token,
		}

		req, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		uuidGenerator := &uuidmocks.Generator{}
		uuidGenerator.On("NewUUID").Return(appName)

		verificationService := &vermocks.Service{}
		verificationService.On("Verify", req, appName).Return(tokenData, nil)

		tokenService := &mocks.Service{}
		tokenService.On("CreateToken", appName, tokenData).Return(token, nil)

		tokenHandler := NewTokenHandler(verificationService, tokenService, host, uuidGenerator)

		// when
		tokenHandler.CreateToken(rr, req)

		// then
		responseBody, err := ioutil.ReadAll(rr.Body)
		require.NoError(t, err)

		var tokenResponse api.TokenResponse
		err = json.Unmarshal(responseBody, &tokenResponse)
		require.NoError(t, err)

		assert.Equal(t, http.StatusCreated, rr.Code)
		assert.EqualValues(t, expectedTokenResponse, tokenResponse)
	})

	t.Run("should return 500 when failed to verify request", func(t *testing.T) {
		// given
		url := "/v1/applications/tokens"

		req, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		uuidGenerator := &uuidmocks.Generator{}
		uuidGenerator.On("NewUUID").Return(appName)

		verificationService := &vermocks.Service{}
		verificationService.On("Verify", req, appName).Return(nil, apperrors.Internal("error"))

		tokenService := &mocks.Service{}

		tokenHandler := NewTokenHandler(verificationService, tokenService, host, uuidGenerator)

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

	t.Run("should return 500 when failed to generate token", func(t *testing.T) {
		// given
		url := "/v1/applications/tokens"

		req, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		uuidGenerator := &uuidmocks.Generator{}
		uuidGenerator.On("NewUUID").Return(appName)

		verificationService := &vermocks.Service{}
		verificationService.On("Verify", req, appName).Return(tokenData, nil)

		tokenService := &mocks.Service{}
		tokenService.On("CreateToken", appName, tokenData).Return("", apperrors.Internal("error"))

		tokenHandler := NewTokenHandler(verificationService, tokenService, host, uuidGenerator)

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
