package externalapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-project/kyma/components/connector-service/pkg/apis/applicationconnector/v1alpha1"

	"github.com/kyma-project/kyma/components/connector-service/internal/tokens"

	kymaGroupMocks "github.com/kyma-project/kyma/components/connector-service/internal/kymagroup/mocks"

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
	group      = "group"
	tenant     = "tenant"
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
	subjectValues = certificates.CSRSubject{
		CName:              identifier,
		Country:            country,
		Organization:       organization,
		OrganizationalUnit: organizationalUnit,
		Locality:           locality,
		Province:           province,
	}

	tokenData = &tokens.TokenData{
		Group:  group,
		Tenant: tenant,
		Token:  token,
	}
)

func TestInfoHandler_GetInfo(t *testing.T) {

	getInfoUrl := fmt.Sprintf("/v1/applications/%s/client-cert?token=%s", identifier, token)

	metadataUrl := fmt.Sprintf("https://gateway.%s/%s/v1/metadata/services", domain, identifier)
	eventsUrl := fmt.Sprintf("https://gateway.%s/%s/v1/events", domain, identifier)

	newToken := "newToken"

	kymaGroup := &v1alpha1.KymaGroup{
		Spec: v1alpha1.KymaGroupSpec{
			Cluster: v1alpha1.Cluster{
				AppRegistryUrl: "https://gateway.domain/%s/v1/metadata/services",
				EventsUrl:      "https://gateway.domain/%s/v1/events",
			},
		},
	}

	t.Run("should get info", func(t *testing.T) {
		// given
		expectedSignUrl := fmt.Sprintf("https://%s/v1/applications/%s/client-certs?token=%s", host, identifier, newToken)

		expectedApi := Api{
			MetadataURL:     metadataUrl,
			EventsURL:       eventsUrl,
			CertificatesUrl: fmt.Sprintf("https://%s/v1/applications/%s", host, identifier),
		}

		tokenService := &tokenMocks.ApplicationService{}
		tokenService.On("GetAppToken", identifier).Return(tokenData, true)
		tokenService.On("CreateAppToken", identifier, tokenData).Return(newToken, nil)

		groupRepository := &kymaGroupMocks.Repository{}
		groupRepository.On("Get", group).Return(kymaGroup, nil)

		infoHandler := NewInfoHandler(tokenService, host, subjectValues, groupRepository)

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
		assert.EqualValues(t, expectedApi, infoResponse.Api)
		assert.EqualValues(t, getExpectedCertInfo(), infoResponse.CertificateInfo)
	})

	t.Run("should return empty urls if Kyma Group not found", func(t *testing.T) {
		// given
		expectedSignUrl := fmt.Sprintf("https://%s/v1/applications/%s/client-certs?token=%s", host, identifier, newToken)

		expectedApi := Api{
			MetadataURL:     "",
			EventsURL:       "",
			CertificatesUrl: fmt.Sprintf("https://%s/v1/applications/%s", host, identifier),
		}

		tokenService := &tokenMocks.ApplicationService{}
		tokenService.On("GetAppToken", identifier).Return(tokenData, true)
		tokenService.On("CreateAppToken", identifier, tokenData).Return(newToken, nil)

		groupRepository := &kymaGroupMocks.Repository{}
		groupRepository.On("Get", group).Return(nil, apperrors.NotFound("error"))

		infoHandler := NewInfoHandler(tokenService, host, subjectValues, groupRepository)

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
		assert.EqualValues(t, expectedApi, infoResponse.Api)
		assert.EqualValues(t, getExpectedCertInfo(), infoResponse.CertificateInfo)
	})

	t.Run("should return 403 when token not provided", func(t *testing.T) {
		// given
		url := fmt.Sprintf("/v1/applications/%s/client-cert", identifier)

		tokenService := &tokenMocks.ApplicationService{}
		groupRepository := &kymaGroupMocks.Repository{}

		infoHandler := NewInfoHandler(tokenService, host, subjectValues, groupRepository)

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

		tokenService := &tokenMocks.ApplicationService{}
		tokenService.On("GetAppToken", identifier).Return(nil, false)

		groupRepository := &kymaGroupMocks.Repository{}

		infoHandler := NewInfoHandler(tokenService, host, subjectValues, groupRepository)

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

		invalidTokenData := &tokens.TokenData{
			Group:  group,
			Tenant: tenant,
			Token:  "invalid token",
		}

		tokenService := &tokenMocks.ApplicationService{}
		tokenService.On("GetAppToken", identifier).Return(invalidTokenData, true)

		groupRepository := &kymaGroupMocks.Repository{}

		infoHandler := NewInfoHandler(tokenService, host, subjectValues, groupRepository)

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

		tokenService := &tokenMocks.ApplicationService{}
		tokenService.On("GetAppToken", identifier).Return(tokenData, true)
		tokenService.On("CreateAppToken", identifier, tokenData).Return("", apperrors.Internal("error"))

		groupRepository := &kymaGroupMocks.Repository{}
		groupRepository.On("Get", group).Return(kymaGroup, nil)

		infoHandler := NewInfoHandler(tokenService, host, subjectValues, groupRepository)

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

	t.Run("should return 500 when failed to read group", func(t *testing.T) {
		// given
		url := fmt.Sprintf("/v1/applications/%s/client-cert?token=%s", identifier, token)

		tokenService := &tokenMocks.ApplicationService{}
		tokenService.On("GetAppToken", identifier).Return(tokenData, true)

		groupRepository := &kymaGroupMocks.Repository{}
		groupRepository.On("Get", group).Return(nil, apperrors.Internal("error"))

		infoHandler := NewInfoHandler(tokenService, host, subjectValues, groupRepository)

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

func getExpectedCertInfo() api.CertificateInfo {
	return api.CertificateInfo{
		Subject:      fmt.Sprintf("OU=%s,O=%s,L=%s,ST=%s,C=%s,CN=%s", organizationalUnit, organization, locality, province, country, identifier),
		Extensions:   "",
		KeyAlgorithm: "rsa2048",
	}
}
