package internalapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-project/kyma/components/connector-service/internal/api"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/httperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/tokens/mocks"
	uuidmocks "github.com/kyma-project/kyma/components/connector-service/internal/uuid/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	host       = "host"
	token      = "token"
	identifier = "identifier"
)

func TestTokenHandler_CreateToken(t *testing.T) {

	t.Run("should create token", func(t *testing.T) {
		// given
		url := "/v1/clusters"

		expectedTokenResponse := api.TokenResponse{
			URL:   fmt.Sprintf("https://%s/v1/clusters/%s/info?token=%s", host, identifier, token),
			Token: token,
		}

		tokenService := &mocks.ClusterService{}
		tokenService.On("CreateClusterToken", identifier).Return(token, nil)

		uuidGenerator := &uuidmocks.Generator{}
		uuidGenerator.On("NewUUID").Return("identifier")

		tokenHandler := NewTokenHandler(tokenService, host, uuidGenerator)

		req, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err)
		rr := httptest.NewRecorder()

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

	t.Run("should return 500 when failed to generate token", func(t *testing.T) {
		// given
		url := fmt.Sprintf("/v1/applications/%s/tokens", identifier)

		tokenService := &mocks.ClusterService{}
		tokenService.On("CreateClusterToken", identifier).Return("", apperrors.Internal("error"))

		uuidGenerator := &uuidmocks.Generator{}
		uuidGenerator.On("NewUUID").Return("identifier")

		tokenHandler := NewTokenHandler(tokenService, host, uuidGenerator)

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
}
