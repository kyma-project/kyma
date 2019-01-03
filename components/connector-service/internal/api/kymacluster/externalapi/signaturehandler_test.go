package externalapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-project/kyma/components/connector-service/pkg/apis/applicationconnector/v1alpha1"

	kymaGroupMocks "github.com/kyma-project/kyma/components/connector-service/internal/kymagroup/mocks"

	"github.com/kyma-project/kyma/components/connector-service/internal/api"

	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	certMock "github.com/kyma-project/kyma/components/connector-service/internal/certificates/mocks"
	"github.com/kyma-project/kyma/components/connector-service/internal/httperrors"
	tokensMock "github.com/kyma-project/kyma/components/connector-service/internal/tokens/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	registryUrl = "http://registry"
	eventsUrl   = "http://events"
)

func TestSignatureHandler_SignCSR(t *testing.T) {

	certRequestRaw := compact([]byte("{\"csr\":\"CSR\",\"kymaCluster\":{\"appRegistryUrl\":\"http://registry\", \"eventsUrl\":\"http://events\"}}"))
	certRequest := CertificateRequest{CSR: "CSR"}
	crtBase64 := "crtBase64"

	clusterData := &v1alpha1.Cluster{
		AppRegistryUrl: registryUrl,
		EventsUrl:      eventsUrl,
	}

	signatureHandlerUrl := fmt.Sprintf("/v1/applications/%s/client-cert?token=%s", identifier, token)

	t.Run("should create certificate when Group and App exist", func(t *testing.T) {
		// given
		tokenService := &tokensMock.ClusterService{}
		tokenService.On("GetClusterToken", identifier).Return(token, true)
		tokenService.On("DeleteClusterToken", identifier)

		certService := &certMock.Service{}
		certService.On("SignCSR", certRequest.CSR, identifier).Return(crtBase64, nil)

		groupRepository := &kymaGroupMocks.Repository{}
		groupRepository.On("UpdateClusterData", identifier, clusterData).Return(nil)

		signatureHandler := NewSignatureHandler(tokenService, certService, host, groupRepository)

		req, err := http.NewRequest(http.MethodPost, signatureHandlerUrl, bytes.NewReader(certRequestRaw))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		req = mux.SetURLVars(req, map[string]string{"identifier": identifier})

		// when
		signatureHandler.SignCSR(rr, req)

		// then
		responseBody, err := ioutil.ReadAll(rr.Body)
		require.NoError(t, err)

		var certResponse api.CertificateResponse
		err = json.Unmarshal(responseBody, &certResponse)
		require.NoError(t, err)

		assert.Equal(t, crtBase64, certResponse.CRT)
		assert.Equal(t, http.StatusCreated, rr.Code)
	})

	t.Run("should return 403 when token not provided", func(t *testing.T) {
		// given
		urlWithoutToken := fmt.Sprintf("/v1/applications/%s/client-cert", identifier)

		tokenService := &tokensMock.ClusterService{}
		certService := &certMock.Service{}

		groupRepository := &kymaGroupMocks.Repository{}

		signatureHandler := NewSignatureHandler(tokenService, certService, host, groupRepository)

		req, err := http.NewRequest(http.MethodPost, urlWithoutToken, bytes.NewReader(certRequestRaw))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		// when
		signatureHandler.SignCSR(rr, req)

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
		tokenService := &tokensMock.ClusterService{}
		tokenService.On("GetClusterToken", identifier).Return("", false)

		certService := &certMock.Service{}

		groupRepository := &kymaGroupMocks.Repository{}

		signatureHandler := NewSignatureHandler(tokenService, certService, host, groupRepository)

		req, err := http.NewRequest(http.MethodPost, signatureHandlerUrl, bytes.NewReader(certRequestRaw))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		req = mux.SetURLVars(req, map[string]string{"identifier": identifier})

		// when
		signatureHandler.SignCSR(rr, req)

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
		tokenService := &tokensMock.ClusterService{}
		tokenService.On("GetClusterToken", identifier).Return("invalid token", true)

		certService := &certMock.Service{}

		groupRepository := &kymaGroupMocks.Repository{}

		signatureHandler := NewSignatureHandler(tokenService, certService, host, groupRepository)

		req, err := http.NewRequest(http.MethodPost, signatureHandlerUrl, bytes.NewReader(certRequestRaw))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		req = mux.SetURLVars(req, map[string]string{"identifier": identifier})

		// when
		signatureHandler.SignCSR(rr, req)

		// then
		responseBody, err := ioutil.ReadAll(rr.Body)
		require.NoError(t, err)

		var errorResponse httperrors.ErrorResponse
		err = json.Unmarshal(responseBody, &errorResponse)
		require.NoError(t, err)

		assert.Equal(t, http.StatusForbidden, errorResponse.Code)
		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("should return 500 when couldn't unmarshal request body", func(t *testing.T) {
		// given
		tokenService := &tokensMock.ClusterService{}
		tokenService.On("GetClusterToken", identifier).Return(token, true)

		certService := &certMock.Service{}

		groupRepository := &kymaGroupMocks.Repository{}

		signatureHandler := NewSignatureHandler(tokenService, certService, host, groupRepository)

		incorrectBody := []byte("incorrectBody")
		req, err := http.NewRequest(http.MethodPost, signatureHandlerUrl, bytes.NewReader(incorrectBody))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		req = mux.SetURLVars(req, map[string]string{"identifier": identifier})

		// when
		signatureHandler.SignCSR(rr, req)

		// then
		responseBody, err := ioutil.ReadAll(rr.Body)
		require.NoError(t, err)

		var errorResponse httperrors.ErrorResponse
		err = json.Unmarshal(responseBody, &errorResponse)
		require.NoError(t, err)

		assert.Equal(t, http.StatusInternalServerError, errorResponse.Code)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("should return 500 when failed to sign CSR ", func(t *testing.T) {
		// given
		url := fmt.Sprintf("/v1/applications/%s/client-cert?token=%s", identifier, token)

		tokenService := &tokensMock.ClusterService{}
		tokenService.On("GetClusterToken", identifier).Return(token, true)

		certService := &certMock.Service{}
		certService.On("SignCSR", certRequest.CSR, identifier).Return("", apperrors.Internal("error"))

		groupRepository := &kymaGroupMocks.Repository{}

		signatureHandler := NewSignatureHandler(tokenService, certService, host, groupRepository)

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(certRequestRaw))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		req = mux.SetURLVars(req, map[string]string{"identifier": identifier})

		// when
		signatureHandler.SignCSR(rr, req)

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

	t.Run("should return 500 when failed to update Kyma Group cluster data", func(t *testing.T) {
		// given
		url := fmt.Sprintf("/v1/applications/%s/client-cert?token=%s", identifier, token)

		tokenService := &tokensMock.ClusterService{}
		tokenService.On("GetClusterToken", identifier).Return(token, true)

		certService := &certMock.Service{}
		certService.On("SignCSR", certRequest.CSR, identifier).Return(crtBase64, nil)

		groupRepository := &kymaGroupMocks.Repository{}
		groupRepository.On("UpdateClusterData", identifier, clusterData).Return(apperrors.Internal("error"))

		signatureHandler := NewSignatureHandler(tokenService, certService, host, groupRepository)

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(certRequestRaw))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		req = mux.SetURLVars(req, map[string]string{"identifier": identifier})

		// when
		signatureHandler.SignCSR(rr, req)

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

func compact(src []byte) []byte {
	buffer := new(bytes.Buffer)
	err := json.Compact(buffer, src)
	if err != nil {
		return src
	}
	return buffer.Bytes()
}
