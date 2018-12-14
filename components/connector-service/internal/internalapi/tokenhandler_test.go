package internalapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/httperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/tokens/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	host    = "host"
	appName = "appName"
	token   = "token"
)

func TestTokenHandler_CreateToken(t *testing.T) {

	t.Run("should create token", func(t *testing.T) {
		// given
		url := fmt.Sprintf("/v1/applications/%s/tokens", appName)

		expectedTokenResponse := tokenResponse{
			URL:   fmt.Sprintf("https://%s/v1/applications/%s/info?token=%s", host, appName, token),
			Token: token,
		}

		tokenGenerator := &mocks.TokenGenerator{}
		tokenGenerator.On("NewToken", appName).Return(token, nil)

		tokenHandler := NewTokenHandler(tokenGenerator, host)

		req, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		req = mux.SetURLVars(req, map[string]string{"appName": appName})

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
	})

	t.Run("should return 500 when failed to generate token", func(t *testing.T) {
		// given
		url := fmt.Sprintf("/v1/applications/%s/tokens", appName)

		tokenGenerator := &mocks.TokenGenerator{}
		tokenGenerator.On("NewToken", appName).Return("", apperrors.Internal("error"))

		tokenHandler := NewTokenHandler(tokenGenerator, host)

		req, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		req = mux.SetURLVars(req, map[string]string{"appName": appName})

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
