package externalapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/certificates"
	"github.com/kyma-project/kyma/components/connector-service/internal/httperrors"
	tokenMocks "github.com/kyma-project/kyma/components/connector-service/internal/tokens/mocks"
	tokenCacheMocks "github.com/kyma-project/kyma/components/connector-service/internal/tokens/tokencache/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInfoHandler_GetInfo(t *testing.T) {

	t.Run("should get info", func(t *testing.T) {
		// given
		newToken := "newToken"
		url := fmt.Sprintf("/v1/remoteenvironments/%s/client-cert?token=%s", reName, token)

		expectedSignUrl := fmt.Sprintf("https://%s/v1/remoteenvironments/%s/client-certs?token=%s", host, reName, newToken)

		expectedApi := api{
			MetadataURL:     fmt.Sprintf("https://gateway.%s/%s/v1/metadata/services", domain, reName),
			EventsURL:       fmt.Sprintf("https://gateway.%s/%s/v1/events", domain, reName),
			CertificatesUrl: fmt.Sprintf("https://%s/v1/remoteenvironments/%s", host, reName),
		}

		expectedCertInfo := certInfo{
			Subject:      fmt.Sprintf("OU=%s,O=%s,L=%s,ST=%s,C=%s,CN=%s", organizationalUnit, organization, locality, province, country, reName),
			Extensions:   "",
			KeyAlgorithm: "rsa2048",
		}

		tokenCache := &tokenCacheMocks.TokenCache{}
		tokenCache.On("Get", reName).Return(token, true)

		tokenGenerator := &tokenMocks.TokenGenerator{}
		tokenGenerator.On("NewToken", reName).Return(newToken, nil)

		subjectValues := certificates.CSRSubject{
			CName:              reName,
			Country:            country,
			Organization:       organization,
			OrganizationalUnit: organizationalUnit,
			Locality:           locality,
			Province:           province,
		}
		infoHandler := NewInfoHandler(tokenCache, tokenGenerator, host, domain, subjectValues)

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(tokenRequestRaw))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		req = mux.SetURLVars(req, map[string]string{"reName": reName})

		// when
		infoHandler.GetInfo(rr, req)

		// then
		responseBody, err := ioutil.ReadAll(rr.Body)
		require.NoError(t, err)

		var infoResponse infoResponse
		err = json.Unmarshal(responseBody, &infoResponse)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, expectedSignUrl, infoResponse.SignUrl)
		assert.EqualValues(t, expectedApi, infoResponse.Api)
		assert.EqualValues(t, expectedCertInfo, infoResponse.CertificateInfo)
	})

	t.Run("should return 403 when token not provided", func(t *testing.T) {
		// given
		url := fmt.Sprintf("/v1/remoteenvironments/%s/client-cert", reName)

		tokenCache := &tokenCacheMocks.TokenCache{}
		tokenGenerator := &tokenMocks.TokenGenerator{}

		subjectValues := certificates.CSRSubject{
			CName:              reName,
			Country:            country,
			Organization:       organization,
			OrganizationalUnit: organizationalUnit,
			Locality:           locality,
			Province:           province,
		}
		infoHandler := NewInfoHandler(tokenCache, tokenGenerator, host, domain, subjectValues)

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(tokenRequestRaw))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		req = mux.SetURLVars(req, map[string]string{"reName": reName})

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
		url := fmt.Sprintf("/v1/remoteenvironments/%s/client-cert?token=%s", reName, token)

		tokenCache := &tokenCacheMocks.TokenCache{}
		tokenCache.On("Get", reName).Return("", false)

		tokenGenerator := &tokenMocks.TokenGenerator{}

		subjectValues := certificates.CSRSubject{
			CName:              reName,
			Country:            country,
			Organization:       organization,
			OrganizationalUnit: organizationalUnit,
			Locality:           locality,
			Province:           province,
		}
		infoHandler := NewInfoHandler(tokenCache, tokenGenerator, host, domain, subjectValues)

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(tokenRequestRaw))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		req = mux.SetURLVars(req, map[string]string{"reName": reName})

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
		url := fmt.Sprintf("/v1/remoteenvironments/%s/client-cert?token=%s", reName, token)

		tokenCache := &tokenCacheMocks.TokenCache{}
		tokenCache.On("Get", reName).Return("differentToken", true)

		tokenGenerator := &tokenMocks.TokenGenerator{}

		subjectValues := certificates.CSRSubject{
			CName:              reName,
			Country:            country,
			Organization:       organization,
			OrganizationalUnit: organizationalUnit,
			Locality:           locality,
			Province:           province,
		}
		infoHandler := NewInfoHandler(tokenCache, tokenGenerator, host, domain, subjectValues)

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(tokenRequestRaw))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		req = mux.SetURLVars(req, map[string]string{"reName": reName})

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
		url := fmt.Sprintf("/v1/remoteenvironments/%s/client-cert?token=%s", reName, token)

		tokenCache := &tokenCacheMocks.TokenCache{}
		tokenCache.On("Get", reName).Return(token, true)

		tokenGenerator := &tokenMocks.TokenGenerator{}
		tokenGenerator.On("NewToken", reName).Return("", apperrors.Internal("error"))

		subjectValues := certificates.CSRSubject{
			CName:              reName,
			Country:            country,
			Organization:       organization,
			OrganizationalUnit: organizationalUnit,
			Locality:           locality,
			Province:           province,
		}
		infoHandler := NewInfoHandler(tokenCache, tokenGenerator, host, domain, subjectValues)

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(tokenRequestRaw))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		req = mux.SetURLVars(req, map[string]string{"reName": reName})

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
