package externalapi

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-project/kyma/components/connector-service/pkg/apis/applicationconnector/v1alpha1"

	"github.com/kyma-project/kyma/components/connector-service/internal/tokens"

	applicationMocks "github.com/kyma-project/kyma/components/connector-service/internal/applications/mocks"
	kymaGroupMocks "github.com/kyma-project/kyma/components/connector-service/internal/kymagroup/mocks"

	"github.com/kyma-project/kyma/components/connector-service/internal/api"

	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/certificates"
	certMock "github.com/kyma-project/kyma/components/connector-service/internal/certificates/mocks"
	"github.com/kyma-project/kyma/components/connector-service/internal/httperrors"
	tokensMock "github.com/kyma-project/kyma/components/connector-service/internal/tokens/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	identifier = "identifier"
	token      = "token"

	host               = "host"
	domain             = "domain"
	country            = "country"
	organization       = "organization"
	organizationalUnit = "organizationalUnit"
	locality           = "locality"
	province           = "province"
)

var (
	tokenRequestRaw = compact([]byte("{\"csr\":\"CSR\"}"))
	caCrtEncoded    = []byte("caCrtEncoded")
	caKeyEncoded    = []byte("caKeyEncoded")

	tokenRequest = CertificateRequest{CSR: "CSR"}
	caCrt        = &x509.Certificate{}
	caKey        = &rsa.PrivateKey{}
	csr          = &x509.CertificateRequest{}
	crtBase64    = "crtBase64"
)

func TestSignatureHandler_SignCSR(t *testing.T) {

	tokenData := &tokens.TokenData{
		Token:  token,
		Group:  group,
		Tenant: tenant,
	}

	kymaGroup := v1alpha1.KymaGroup{
		Spec: v1alpha1.KymaGroupSpec{
			Cluster: v1alpha1.Cluster{},
		},
	}

	appGroupEntry := v1alpha1.Application{ID: identifier}

	t.Run("should create certificate", func(t *testing.T) {
		// given
		url := fmt.Sprintf("/v1/applications/%s/client-cert?token=%s", identifier, token)

		tokenService := &tokensMock.ApplicationService{}
		tokenService.On("GetToken", identifier).Return(tokenData, true)
		tokenService.On("DeleteToken", identifier).Return()

		certService := &certMock.Service{}
		certService.On("SignCSR", tokenRequest.CSR, identifier).Return(crtBase64, nil)

		groupRepository := &kymaGroupMocks.Repository{}
		groupRepository.On("Get", group).Return(kymaGroup, nil)
		groupRepository.On("AddApplication", group, appGroupEntry).Return(kymaGroup, nil)

		appRepository := &applicationMocks.Repository{}
		appRepository.On("Get", identifier).Return(nil, apperrors.NotFound("error"))
		appRepository.On("Create", identifier).Return(nil, apperrors.NotFound("error"))

		signatureHandler := NewSignatureHandler(tokenService, certService, host, domain, groupRepository, appRepository)

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(tokenRequestRaw))
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
		url := fmt.Sprintf("/v1/applications/%s/client-cert", identifier)

		tokenService := &tokensMock.ApplicationService{}
		certService := &certMock.Service{}

		groupRepository := &kymaGroupMocks.Repository{}
		appRepository := &applicationMocks.Repository{}

		registrationHandler := NewSignatureHandler(tokenService, certService, host, domain, groupRepository, appRepository)

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(tokenRequestRaw))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		// when
		registrationHandler.SignCSR(rr, req)

		// then
		responseBody, err := ioutil.ReadAll(rr.Body)
		require.NoError(t, err)

		var errorResponse httperrors.ErrorResponse
		err = json.Unmarshal(responseBody, &errorResponse)
		require.NoError(t, err)

		assert.Equal(t, http.StatusForbidden, errorResponse.Code)
		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("should return 400 when token not found", func(t *testing.T) {
		// given
		url := fmt.Sprintf("/v1/applications/%s/client-cert?token=%s", identifier, token)

		tokenService := &tokensMock.ApplicationService{}
		tokenService.On("GetToken", identifier).Return("", false)

		certService := &certMock.Service{}

		groupRepository := &kymaGroupMocks.Repository{}
		appRepository := &applicationMocks.Repository{}

		registrationHandler := NewSignatureHandler(tokenService, certService, host, domain, groupRepository, appRepository)

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(tokenRequestRaw))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		req = mux.SetURLVars(req, map[string]string{"identifier": identifier})

		// when
		registrationHandler.SignCSR(rr, req)

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

		tokenService := &tokensMock.ApplicationService{}
		tokenService.On("GetToken", identifier).Return("differentToken", true)

		certService := &certMock.Service{}

		groupRepository := &kymaGroupMocks.Repository{}
		appRepository := &applicationMocks.Repository{}

		signatureHandler := NewSignatureHandler(tokenService, certService, host, domain, groupRepository, appRepository)

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(tokenRequestRaw))
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
		url := fmt.Sprintf("/v1/applications/%s/client-cert?token=%s", identifier, token)

		tokenService := &tokensMock.ApplicationService{}
		tokenService.On("GetToken", identifier).Return(token, true)

		certService := &certMock.Service{}

		groupRepository := &kymaGroupMocks.Repository{}
		appRepository := &applicationMocks.Repository{}

		signatureHandler := NewSignatureHandler(tokenService, certService, host, domain, groupRepository, appRepository)

		incorrectBody := []byte("incorrectBody")
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(incorrectBody))
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

	t.Run("should return 404 when secret not found", func(t *testing.T) {
		// given
		url := fmt.Sprintf("/v1/applications/%s/client-cert?token=%s", identifier, token)

		tokenService := &tokensMock.ApplicationService{}
		tokenService.On("GetToken", identifier).Return(token, true)
		tokenService.On("DeleteToken", identifier).Return()

		certService := &certMock.Service{}
		certService.On("LoadCSR", mock.Anything).Return(nil, nil)
		certService.On("CheckCSRValues", mock.Anything, mock.Anything).Return(nil, nil)

		groupRepository := &kymaGroupMocks.Repository{}
		appRepository := &applicationMocks.Repository{}

		signatureHandler := NewSignatureHandler(tokenService, certService, host, domain, groupRepository, appRepository)

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(tokenRequestRaw))
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

		assert.Equal(t, http.StatusNotFound, errorResponse.Code)
		assert.Equal(t, "error", errorResponse.Error)
	})

	t.Run("should return 500 when couldn't load cert", func(t *testing.T) {
		// given
		url := fmt.Sprintf("/v1/applications/%s/client-cert?token=%s", identifier, token)

		tokenService := &tokensMock.ApplicationService{}
		tokenService.On("GetToken", identifier).Return(token, true)

		certService := &certMock.Service{}
		certService.On("LoadCSR", mock.Anything).Return(nil, nil)
		certService.On("CheckCSRValues", mock.Anything, mock.Anything).Return(nil, nil)
		certService.On("LoadCert", caCrtEncoded).Return(nil, apperrors.Internal("error"))

		groupRepository := &kymaGroupMocks.Repository{}
		appRepository := &applicationMocks.Repository{}

		signatureHandler := NewSignatureHandler(tokenService, certService, host, domain, groupRepository, appRepository)

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(tokenRequestRaw))
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

	t.Run("should return 500 when couldn't load key", func(t *testing.T) {
		// given
		url := fmt.Sprintf("/v1/applications/%s/client-cert?token=%s", identifier, token)

		tokenService := &tokensMock.ApplicationService{}
		tokenService.On("GetToken", identifier).Return(token, true)

		certService := &certMock.Service{}
		certService.On("LoadCSR", mock.Anything).Return(nil, nil)
		certService.On("CheckCSRValues", mock.Anything, mock.Anything).Return(nil, nil)
		certService.On("LoadCert", caCrtEncoded).Return(caCrt, nil)
		certService.On("LoadKey", caKeyEncoded).Return(nil, apperrors.Internal("error"))

		groupRepository := &kymaGroupMocks.Repository{}
		appRepository := &applicationMocks.Repository{}

		signatureHandler := NewSignatureHandler(tokenService, certService, host, domain, groupRepository, appRepository)

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(tokenRequestRaw))
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

	t.Run("should return 500 when couldn't load CSR", func(t *testing.T) {
		// given
		url := fmt.Sprintf("/v1/applications/%s/client-cert?token=%s", identifier, token)

		tokenService := &tokensMock.ApplicationService{}
		tokenService.On("GetToken", identifier).Return(token, true)

		certService := &certMock.Service{}
		certService.On("LoadCert", caCrtEncoded).Return(caCrt, nil)
		certService.On("LoadKey", caKeyEncoded).Return(caKey, nil)
		certService.On("LoadCSR", tokenRequest.CSR).Return(nil, apperrors.Internal("error"))

		groupRepository := &kymaGroupMocks.Repository{}
		appRepository := &applicationMocks.Repository{}

		signatureHandler := NewSignatureHandler(tokenService, certService, host, domain, groupRepository, appRepository)

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(tokenRequestRaw))
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

	t.Run("should return 500 when failed to check CSR values", func(t *testing.T) {
		// given
		url := fmt.Sprintf("/v1/applications/%s/client-cert?token=%s", identifier, token)

		tokenService := &tokensMock.ApplicationService{}
		tokenService.On("GetToken", identifier).Return(token, true)

		certService := &certMock.Service{}
		certService.On("LoadCert", caCrtEncoded).Return(caCrt, nil)
		certService.On("LoadKey", caKeyEncoded).Return(caKey, nil)
		certService.On("LoadCSR", tokenRequest.CSR).Return(csr, nil)

		subjectValues := certificates.CSRSubject{
			CName:              identifier,
			Country:            country,
			Organization:       organization,
			OrganizationalUnit: organizationalUnit,
			Locality:           locality,
			Province:           province,
		}
		certService.On("CheckCSRValues", csr, subjectValues).Return(apperrors.Forbidden("error"))

		groupRepository := &kymaGroupMocks.Repository{}
		appRepository := &applicationMocks.Repository{}

		signatureHandler := NewSignatureHandler(tokenService, certService, host, domain, groupRepository, appRepository)

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(tokenRequestRaw))
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
		assert.Equal(t, "error", errorResponse.Error)
		assert.Equal(t, http.StatusForbidden, rr.Code)
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
