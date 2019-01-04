package externalapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1"

	v1alpha12 "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"

	"github.com/kyma-project/kyma/components/connector-service/pkg/apis/applicationconnector/v1alpha1"

	"github.com/kyma-project/kyma/components/connector-service/internal/tokens"

	applicationMocks "github.com/kyma-project/kyma/components/connector-service/internal/applications/mocks"
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

func TestSignatureHandler_SignCSR(t *testing.T) {

	certRequestRaw := compact([]byte("{\"csr\":\"CSR\"}"))
	certRequest := CertificateRequest{CSR: "CSR"}
	crtBase64 := "crtBase64"

	kymaGroup := &v1alpha1.KymaGroup{
		ObjectMeta: v1.ObjectMeta{
			Name: group,
		},
		Spec: v1alpha1.KymaGroupSpec{
			Cluster: v1alpha1.Cluster{},
		},
	}

	appGroupEntry := &v1alpha1.Application{ID: identifier}

	application := &v1alpha12.Application{
		ObjectMeta: v1.ObjectMeta{
			Name: identifier,
		},
		Spec: v1alpha12.ApplicationSpec{
			Description: "Description",
			Services:    []v1alpha12.Service{},
		},
	}

	signatureHandlerUrl := fmt.Sprintf("/v1/applications/%s/client-cert?token=%s", identifier, token)

	t.Run("should create certificate when Group and App exist", func(t *testing.T) {
		// given
		tokenService := &tokensMock.ApplicationService{}
		tokenService.On("GetAppToken", identifier).Return(tokenData, true)
		tokenService.On("DeleteAppToken", identifier)

		certService := &certMock.Service{}
		certService.On("SignCSR", certRequest.CSR, identifier).Return(crtBase64, nil)

		groupRepository := &kymaGroupMocks.Repository{}
		groupRepository.On("Get", group).Return(kymaGroup, nil)
		groupRepository.On("AddApplication", group, appGroupEntry).Return(nil)

		appRepository := &applicationMocks.Repository{}
		appRepository.On("Get", identifier).Return(application, nil)

		signatureHandler := NewSignatureHandler(tokenService, certService, groupRepository, appRepository)

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

	t.Run("should create certificate together with Group and App", func(t *testing.T) {
		// given
		rawCertRequest := compact([]byte("{\"csr\":\"CSR\", \"application\":{\"description\":\"description\"}}"))

		createdKymaGroup := &v1alpha1.KymaGroup{
			TypeMeta:   v1.TypeMeta{Kind: "KymaGroup", APIVersion: v1alpha1.SchemeGroupVersion.String()},
			ObjectMeta: v1.ObjectMeta{Name: group},
			Spec: v1alpha1.KymaGroupSpec{
				Applications: []v1alpha1.Application{*appGroupEntry},
				Cluster:      v1alpha1.Cluster{},
			},
		}

		createdApplication := &v1alpha12.Application{
			TypeMeta:   v1.TypeMeta{Kind: "Application", APIVersion: v1alpha12.SchemeGroupVersion.String()},
			ObjectMeta: v1.ObjectMeta{Name: identifier},
			Spec: v1alpha12.ApplicationSpec{
				Description: "description",
				Services:    []v1alpha12.Service{},
			},
		}

		tokenService := &tokensMock.ApplicationService{}
		tokenService.On("GetAppToken", identifier).Return(tokenData, true)
		tokenService.On("DeleteAppToken", identifier)

		certService := &certMock.Service{}
		certService.On("SignCSR", certRequest.CSR, identifier).Return(crtBase64, nil)

		groupRepository := &kymaGroupMocks.Repository{}
		groupRepository.On("Get", group).Return(nil, apperrors.NotFound("error"))
		groupRepository.On("Create", createdKymaGroup).Return(nil)

		appRepository := &applicationMocks.Repository{}
		appRepository.On("Get", identifier).Return(nil, apperrors.NotFound("error"))
		appRepository.On("Create", createdApplication).Return(nil)

		signatureHandler := NewSignatureHandler(tokenService, certService, groupRepository, appRepository)

		req, err := http.NewRequest(http.MethodPost, signatureHandlerUrl, bytes.NewReader(rawCertRequest))
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

		tokenService := &tokensMock.ApplicationService{}
		certService := &certMock.Service{}

		groupRepository := &kymaGroupMocks.Repository{}
		appRepository := &applicationMocks.Repository{}

		signatureHandler := NewSignatureHandler(tokenService, certService, groupRepository, appRepository)

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
		tokenService := &tokensMock.ApplicationService{}
		tokenService.On("GetAppToken", identifier).Return(nil, false)

		certService := &certMock.Service{}

		groupRepository := &kymaGroupMocks.Repository{}
		appRepository := &applicationMocks.Repository{}

		signatureHandler := NewSignatureHandler(tokenService, certService, groupRepository, appRepository)

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
		invalidTokenData := &tokens.TokenData{
			Token: "invalid token",
		}

		tokenService := &tokensMock.ApplicationService{}
		tokenService.On("GetAppToken", identifier).Return(invalidTokenData, true)

		certService := &certMock.Service{}

		groupRepository := &kymaGroupMocks.Repository{}
		appRepository := &applicationMocks.Repository{}

		signatureHandler := NewSignatureHandler(tokenService, certService, groupRepository, appRepository)

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
		tokenService := &tokensMock.ApplicationService{}
		tokenService.On("GetAppToken", identifier).Return(tokenData, true)

		certService := &certMock.Service{}

		groupRepository := &kymaGroupMocks.Repository{}
		appRepository := &applicationMocks.Repository{}

		signatureHandler := NewSignatureHandler(tokenService, certService, groupRepository, appRepository)

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

		tokenService := &tokensMock.ApplicationService{}
		tokenService.On("GetAppToken", identifier).Return(tokenData, true)

		certService := &certMock.Service{}
		certService.On("SignCSR", certRequest.CSR, identifier).Return("", apperrors.Internal("error"))

		groupRepository := &kymaGroupMocks.Repository{}
		appRepository := &applicationMocks.Repository{}

		signatureHandler := NewSignatureHandler(tokenService, certService, groupRepository, appRepository)

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

	t.Run("should return 500 when failed to read Kyma Group", func(t *testing.T) {
		// given
		url := fmt.Sprintf("/v1/applications/%s/client-cert?token=%s", identifier, token)

		tokenService := &tokensMock.ApplicationService{}
		tokenService.On("GetAppToken", identifier).Return(tokenData, true)

		certService := &certMock.Service{}
		certService.On("SignCSR", certRequest.CSR, identifier).Return(crtBase64, nil)

		groupRepository := &kymaGroupMocks.Repository{}
		groupRepository.On("Get", group).Return(nil, apperrors.Internal("error"))

		appRepository := &applicationMocks.Repository{}
		appRepository.On("Get", identifier).Return(application, nil)

		signatureHandler := NewSignatureHandler(tokenService, certService, groupRepository, appRepository)

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

	t.Run("should return 500 when failed to read Application", func(t *testing.T) {
		// given
		url := fmt.Sprintf("/v1/applications/%s/client-cert?token=%s", identifier, token)

		tokenService := &tokensMock.ApplicationService{}
		tokenService.On("GetAppToken", identifier).Return(tokenData, true)

		certService := &certMock.Service{}
		certService.On("SignCSR", certRequest.CSR, identifier).Return(crtBase64, nil)

		groupRepository := &kymaGroupMocks.Repository{}

		appRepository := &applicationMocks.Repository{}
		appRepository.On("Get", identifier).Return(nil, apperrors.Internal("error"))

		signatureHandler := NewSignatureHandler(tokenService, certService, groupRepository, appRepository)

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
