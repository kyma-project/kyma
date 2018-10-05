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

	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/certificates"
	certMock "github.com/kyma-project/kyma/components/connector-service/internal/certificates/mocks"
	"github.com/kyma-project/kyma/components/connector-service/internal/httperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/secrets"
	secrectsMock "github.com/kyma-project/kyma/components/connector-service/internal/secrets/mocks"
	"github.com/kyma-project/kyma/components/connector-service/internal/tokens/tokencache"
	tokensMock "github.com/kyma-project/kyma/components/connector-service/internal/tokens/tokencache/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	authSecretName = "nginx-auth-ca"
	reName         = "reName"
	token          = "token"

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

	tokenRequest = certRequest{CSR: "CSR"}
	caCrt        = &x509.Certificate{}
	caKey        = &rsa.PrivateKey{}
	csr          = &x509.CertificateRequest{}
	crtBase64    = "crtBase64"
)

func TestSignatureHandler_SignCSR(t *testing.T) {

	t.Run("should create certificate", func(t *testing.T) {
		// given
		url := fmt.Sprintf("/v1/remoteenvironments/%s/client-cert?token=%s", reName, token)

		tokenCache := &tokensMock.TokenCache{}
		tokenCache.On("Get", reName).Return(token, true)
		tokenCache.On("Delete", reName).Return()

		secretsRepository := &secrectsMock.Repository{}
		secretsRepository.On("Get", authSecretName).Return(caCrtEncoded, caKeyEncoded, nil)

		certUtils := &certMock.CertificateUtility{}
		certUtils.On("LoadCert", caCrtEncoded).Return(caCrt, nil)
		certUtils.On("LoadKey", caKeyEncoded).Return(caKey, nil)
		certUtils.On("LoadCSR", tokenRequest.CSR).Return(csr, nil)

		subjectValues := certificates.CSRSubject{
			CName:              reName,
			Country:            country,
			Organization:       organization,
			OrganizationalUnit: organizationalUnit,
			Locality:           locality,
			Province:           province,
		}
		certUtils.On("CheckCSRValues", csr, subjectValues).Return(nil)
		certUtils.On("CreateCrtChain", caCrt, csr, caKey).Return(crtBase64, nil)

		registrationHandler := createSignatureHandler(tokenCache, certUtils, secretsRepository)

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(tokenRequestRaw))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		req = mux.SetURLVars(req, map[string]string{"reName": reName})

		// when
		registrationHandler.SignCSR(rr, req)

		// then
		responseBody, err := ioutil.ReadAll(rr.Body)
		require.NoError(t, err)

		var certResponse certResponse
		err = json.Unmarshal(responseBody, &certResponse)
		require.NoError(t, err)

		assert.Equal(t, crtBase64, certResponse.CRT)
		assert.Equal(t, http.StatusCreated, rr.Code)
	})

	t.Run("should return 403 when token not provided", func(t *testing.T) {
		// given
		url := fmt.Sprintf("/v1/remoteenvironments/%s/client-cert", reName)

		tokenCache := &tokensMock.TokenCache{}
		secretsRepository := &secrectsMock.Repository{}
		certUtils := &certMock.CertificateUtility{}

		registrationHandler := createSignatureHandler(tokenCache, certUtils, secretsRepository)

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
		url := fmt.Sprintf("/v1/remoteenvironments/%s/client-cert?token=%s", reName, token)

		tokenCache := &tokensMock.TokenCache{}
		tokenCache.On("Get", reName).Return("", false)

		secretsRepository := &secrectsMock.Repository{}
		certUtils := &certMock.CertificateUtility{}

		registrationHandler := createSignatureHandler(tokenCache, certUtils, secretsRepository)

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(tokenRequestRaw))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		req = mux.SetURLVars(req, map[string]string{"reName": reName})

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
		url := fmt.Sprintf("/v1/remoteenvironments/%s/client-cert?token=%s", reName, token)

		tokenCache := &tokensMock.TokenCache{}
		tokenCache.On("Get", reName).Return("differentToken", true)

		secretsRepository := &secrectsMock.Repository{}
		certUtils := &certMock.CertificateUtility{}

		registrationHandler := createSignatureHandler(tokenCache, certUtils, secretsRepository)

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(tokenRequestRaw))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		req = mux.SetURLVars(req, map[string]string{"reName": reName})

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

	t.Run("should return 500 when couldn't unmarshal request body", func(t *testing.T) {
		// given
		url := fmt.Sprintf("/v1/remoteenvironments/%s/client-cert?token=%s", reName, token)

		tokenCache := &tokensMock.TokenCache{}
		tokenCache.On("Get", reName).Return(token, true)

		secretsRepository := &secrectsMock.Repository{}
		certUtils := &certMock.CertificateUtility{}

		registrationHandler := createSignatureHandler(tokenCache, certUtils, secretsRepository)

		incorrectBody := []byte("incorrectBody")
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(incorrectBody))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		req = mux.SetURLVars(req, map[string]string{"reName": reName})

		// when
		registrationHandler.SignCSR(rr, req)

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
		url := fmt.Sprintf("/v1/remoteenvironments/%s/client-cert?token=%s", reName, token)

		tokenCache := &tokensMock.TokenCache{}
		tokenCache.On("Get", reName).Return(token, true)
		tokenCache.On("Delete", reName).Return()

		secretNotFoundError := apperrors.NotFound("error")
		secretsRepository := &secrectsMock.Repository{}
		secretsRepository.On("Get", authSecretName).Return([]byte(""), []byte(""), secretNotFoundError)

		certUtils := &certMock.CertificateUtility{}
		certUtils.On("LoadCSR", mock.Anything).Return(nil, nil)
		certUtils.On("CheckCSRValues", mock.Anything, mock.Anything).Return(nil, nil)

		registrationHandler := createSignatureHandler(tokenCache, certUtils, secretsRepository)

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(tokenRequestRaw))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		req = mux.SetURLVars(req, map[string]string{"reName": reName})

		// when
		registrationHandler.SignCSR(rr, req)

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
		url := fmt.Sprintf("/v1/remoteenvironments/%s/client-cert?token=%s", reName, token)

		tokenCache := &tokensMock.TokenCache{}
		tokenCache.On("Get", reName).Return(token, true)

		secretsRepository := &secrectsMock.Repository{}
		secretsRepository.On("Get", authSecretName).Return(caCrtEncoded, caKeyEncoded, nil)

		certUtils := &certMock.CertificateUtility{}
		certUtils.On("LoadCSR", mock.Anything).Return(nil, nil)
		certUtils.On("CheckCSRValues", mock.Anything, mock.Anything).Return(nil, nil)
		certUtils.On("LoadCert", caCrtEncoded).Return(nil, apperrors.Internal("error"))

		registrationHandler := createSignatureHandler(tokenCache, certUtils, secretsRepository)

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(tokenRequestRaw))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		req = mux.SetURLVars(req, map[string]string{"reName": reName})

		// when
		registrationHandler.SignCSR(rr, req)

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
		url := fmt.Sprintf("/v1/remoteenvironments/%s/client-cert?token=%s", reName, token)

		tokenCache := &tokensMock.TokenCache{}
		tokenCache.On("Get", reName).Return(token, true)

		secretsRepository := &secrectsMock.Repository{}
		secretsRepository.On("Get", authSecretName).Return(caCrtEncoded, caKeyEncoded, nil)

		certUtils := &certMock.CertificateUtility{}
		certUtils.On("LoadCSR", mock.Anything).Return(nil, nil)
		certUtils.On("CheckCSRValues", mock.Anything, mock.Anything).Return(nil, nil)
		certUtils.On("LoadCert", caCrtEncoded).Return(caCrt, nil)
		certUtils.On("LoadKey", caKeyEncoded).Return(nil, apperrors.Internal("error"))

		registrationHandler := createSignatureHandler(tokenCache, certUtils, secretsRepository)

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(tokenRequestRaw))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		req = mux.SetURLVars(req, map[string]string{"reName": reName})

		// when
		registrationHandler.SignCSR(rr, req)

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
		url := fmt.Sprintf("/v1/remoteenvironments/%s/client-cert?token=%s", reName, token)

		tokenCache := &tokensMock.TokenCache{}
		tokenCache.On("Get", reName).Return(token, true)

		secretsRepository := &secrectsMock.Repository{}
		secretsRepository.On("Get", authSecretName).Return(caCrtEncoded, caKeyEncoded, nil)

		certUtils := &certMock.CertificateUtility{}
		certUtils.On("LoadCert", caCrtEncoded).Return(caCrt, nil)
		certUtils.On("LoadKey", caKeyEncoded).Return(caKey, nil)
		certUtils.On("LoadCSR", tokenRequest.CSR).Return(nil, apperrors.Internal("error"))

		registrationHandler := createSignatureHandler(tokenCache, certUtils, secretsRepository)

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(tokenRequestRaw))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		req = mux.SetURLVars(req, map[string]string{"reName": reName})

		// when
		registrationHandler.SignCSR(rr, req)

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
		url := fmt.Sprintf("/v1/remoteenvironments/%s/client-cert?token=%s", reName, token)

		tokenCache := &tokensMock.TokenCache{}
		tokenCache.On("Get", reName).Return(token, true)

		secretsRepository := &secrectsMock.Repository{}
		secretsRepository.On("Get", authSecretName).Return(caCrtEncoded, caKeyEncoded, nil)

		certUtils := &certMock.CertificateUtility{}
		certUtils.On("LoadCert", caCrtEncoded).Return(caCrt, nil)
		certUtils.On("LoadKey", caKeyEncoded).Return(caKey, nil)
		certUtils.On("LoadCSR", tokenRequest.CSR).Return(csr, nil)

		subjectValues := certificates.CSRSubject{
			CName:              reName,
			Country:            country,
			Organization:       organization,
			OrganizationalUnit: organizationalUnit,
			Locality:           locality,
			Province:           province,
		}
		certUtils.On("CheckCSRValues", csr, subjectValues).Return(apperrors.Forbidden("error"))

		registrationHandler := createSignatureHandler(tokenCache, certUtils, secretsRepository)

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(tokenRequestRaw))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		req = mux.SetURLVars(req, map[string]string{"reName": reName})

		// when
		registrationHandler.SignCSR(rr, req)

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

func createSignatureHandler(tokenCache tokencache.TokenCache, certUtils certificates.CertificateUtility, secretsRepository secrets.Repository) SignatureHandler {
	subjectValues := certificates.CSRSubject{
		CName:              reName,
		Country:            country,
		Organization:       organization,
		OrganizationalUnit: organizationalUnit,
		Locality:           locality,
		Province:           province,
	}

	return NewSignatureHandler(tokenCache, certUtils, secretsRepository, host, domain, subjectValues)
}

func compact(src []byte) []byte {
	buffer := new(bytes.Buffer)
	err := json.Compact(buffer, src)
	if err != nil {
		return src
	}
	return buffer.Bytes()
}
