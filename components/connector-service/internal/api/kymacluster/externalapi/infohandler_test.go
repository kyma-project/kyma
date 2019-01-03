package externalapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-project/kyma/components/connector-service/internal/api"

	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/certificates"
	"github.com/kyma-project/kyma/components/connector-service/internal/httperrors"
	tokenMocks "github.com/kyma-project/kyma/components/connector-service/internal/tokens/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	identifier = "identifier"
	token      = "token"

	host               = "host"
	country            = "country"
	organization       = "organization"
	organizationalUnit = "organizationalUnit"
	locality           = "locality"
	province           = "province"
)

var (
	subjectValues = certificates.CSRSubject{
		CName:              identifier,
		Country:            country,
		Organization:       organization,
		OrganizationalUnit: organizationalUnit,
		Locality:           locality,
		Province:           province,
	}

	getInfoUrl = fmt.Sprintf("/v1/clusters/%s/client-cert?token=%s", identifier, token)
)

func TestInfoHandler_GetInfo(t *testing.T) {

	newToken := "newToken"

	t.Run("should get info", func(t *testing.T) {
		// given
		expectedSignUrl := fmt.Sprintf("https://%s/v1/clusters/%s/client-certs?token=%s", host, identifier, newToken)

		expectedCertInfo := api.CertificateInfo{
			Subject:      fmt.Sprintf("OU=%s,O=%s,L=%s,ST=%s,C=%s,CN=%s", organizationalUnit, organization, locality, province, country, identifier),
			Extensions:   "",
			KeyAlgorithm: "rsa2048",
		}

		tokenService := &tokenMocks.ClusterService{}
		tokenService.On("GetClusterToken", identifier).Return(token, true)
		tokenService.On("CreateClusterToken", identifier).Return(newToken, nil)

		infoHandler := NewInfoHandler(tokenService, host, subjectValues)

		req, err := http.NewRequest(http.MethodPost, getInfoUrl, nil)
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		req = mux.SetURLVars(req, map[string]string{"identifier": identifier})

		// when
		infoHandler.GetInfo(rr, req)

		// then
		responseBody, err := ioutil.ReadAll(rr.Body)
		require.NoError(t, err)

		var infoResponse InfoResponse
		err = json.Unmarshal(responseBody, &infoResponse)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, expectedSignUrl, infoResponse.SignUrl)
		assert.EqualValues(t, expectedCertInfo, infoResponse.CertificateInfo)
	})

	t.Run("should return 403 when token not provided", func(t *testing.T) {
		// given
		url := fmt.Sprintf("/v1/applications/%s/client-cert", identifier)

		tokenService := &tokenMocks.ClusterService{}

		infoHandler := NewInfoHandler(tokenService, host, subjectValues)

		req, err := http.NewRequest(http.MethodPost, url, nil)
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		req = mux.SetURLVars(req, map[string]string{"identifier": identifier})

		// when
		infoHandler.GetInfo(rr, req)

		// then
		responseBody, err := ioutil.ReadAll(rr.Body)
		require.NoError(t, err)

		var errorResponse httperrors.ErrorResponse
		err = json.Unmarshal(responseBody, &errorResponse)
		require.NoError(t, err)

		assert.Equal(t, http.StatusForbidden, errorResponse.Code)
		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("should return 403 when token not found", func(t *testing.T) {
		// given
		url := fmt.Sprintf("/v1/applications/%s/client-cert?token=%s", identifier, token)

		tokenService := &tokenMocks.ClusterService{}
		tokenService.On("GetClusterToken", identifier).Return("", false)

		infoHandler := NewInfoHandler(tokenService, host, subjectValues)

		req, err := http.NewRequest(http.MethodPost, url, nil)
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		req = mux.SetURLVars(req, map[string]string{"identifier": identifier})

		// when
		infoHandler.GetInfo(rr, req)

		// then
		responseBody, err := ioutil.ReadAll(rr.Body)
		require.NoError(t, err)

		var errorResponse httperrors.ErrorResponse
		err = json.Unmarshal(responseBody, &errorResponse)
		require.NoError(t, err)

		assert.Equal(t, http.StatusForbidden, errorResponse.Code)
		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("should return 403 when wrong token provided", func(t *testing.T) {
		// given
		url := fmt.Sprintf("/v1/applications/%s/client-cert?token=%s", identifier, token)

		tokenService := &tokenMocks.ClusterService{}
		tokenService.On("GetClusterToken", identifier).Return("invalid token", true)

		infoHandler := NewInfoHandler(tokenService, host, subjectValues)

		req, err := http.NewRequest(http.MethodPost, url, nil)
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		req = mux.SetURLVars(req, map[string]string{"identifier": identifier})

		// when
		infoHandler.GetInfo(rr, req)

		// then
		responseBody, err := ioutil.ReadAll(rr.Body)
		require.NoError(t, err)

		var errorResponse httperrors.ErrorResponse
		err = json.Unmarshal(responseBody, &errorResponse)
		require.NoError(t, err)

		assert.Equal(t, http.StatusForbidden, errorResponse.Code)
		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("should return 500 when failed to generate new token", func(t *testing.T) {
		// given
		url := fmt.Sprintf("/v1/applications/%s/client-cert?token=%s", identifier, token)

		tokenService := &tokenMocks.ClusterService{}
		tokenService.On("GetClusterToken", identifier).Return(token, true)
		tokenService.On("CreateClusterToken", identifier).Return("", apperrors.Internal("error"))

		infoHandler := NewInfoHandler(tokenService, host, subjectValues)

		req, err := http.NewRequest(http.MethodPost, url, nil)
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		req = mux.SetURLVars(req, map[string]string{"identifier": identifier})

		// when
		infoHandler.GetInfo(rr, req)

		// then
		responseBody, err := ioutil.ReadAll(rr.Body)
		require.NoError(t, err)

		var errorResponse httperrors.ErrorResponse
		err = json.Unmarshal(responseBody, &errorResponse)
		require.NoError(t, err)

		assert.Equal(t, http.StatusInternalServerError, errorResponse.Code)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}
