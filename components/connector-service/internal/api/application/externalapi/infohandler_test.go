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
	group  = "group"
	tenant = "tenant"
)

func TestInfoHandler_GetInfo(t *testing.T) {

	tokenData := &tokens.TokenData{
		Group:  group,
		Tenant: tenant,
		Token:  token,
	}

	t.Run("should get info", func(t *testing.T) {
		// given
		newToken := "newToken"
		url := fmt.Sprintf("/v1/applications/%s/client-cert?token=%s", identifier, token)

		expectedSignUrl := fmt.Sprintf("https://%s/v1/applications/%s/client-certs?token=%s", host, identifier, newToken)

		metadataUrl := fmt.Sprintf("https://gateway.%s/%s/v1/metadata/services", domain, identifier)
		eventsUrl := fmt.Sprintf("https://gateway.%s/%s/v1/events", domain, identifier)

		kymaGroup := v1alpha1.KymaGroup{
			Spec: v1alpha1.KymaGroupSpec{
				Cluster: v1alpha1.Cluster{
					AppRegistryUrl: metadataUrl,
					EventsUrl:      eventsUrl,
				},
			},
		}

		expectedApi := Api{
			MetadataURL:     metadataUrl,
			EventsURL:       eventsUrl,
			CertificatesUrl: fmt.Sprintf("https://%s/v1/applications/%s", host, identifier),
		}

		expectedCertInfo := api.CertificateInfo{
			Subject:      fmt.Sprintf("OU=%s,O=%s,L=%s,ST=%s,C=%s,CN=%s", organizationalUnit, organization, locality, province, country, identifier),
			Extensions:   "",
			KeyAlgorithm: "rsa2048",
		}

		tokenService := &tokenMocks.Service{}
		tokenService.On("GetToken", identifier).Return(tokenData, true)
		tokenService.On("CreateToken", identifier).Return(newToken, nil)

		groupRepository := &kymaGroupMocks.Repository{}
		groupRepository.On("Get", group).Return(kymaGroup, nil)

		subjectValues := certificates.CSRSubject{
			CName:              identifier,
			Country:            country,
			Organization:       organization,
			OrganizationalUnit: organizationalUnit,
			Locality:           locality,
			Province:           province,
		}

		infoHandler := NewInfoHandler(tokenService, host, domain, subjectValues, groupRepository)

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(tokenRequestRaw))
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
		assert.EqualValues(t, expectedCertInfo, infoResponse.CertificateInfo)
	})

	t.Run("should return 403 when token not provided", func(t *testing.T) {
		// given
		url := fmt.Sprintf("/v1/applications/%s/client-cert", identifier)

		tokenService := &tokenMocks.Service{}
		groupRepository := &kymaGroupMocks.Repository{}

		subjectValues := certificates.CSRSubject{
			CName:              identifier,
			Country:            country,
			Organization:       organization,
			OrganizationalUnit: organizationalUnit,
			Locality:           locality,
			Province:           province,
		}
		infoHandler := NewInfoHandler(tokenService, host, domain, subjectValues, groupRepository)

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(tokenRequestRaw))
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

		tokenService := &tokenMocks.Service{}
		tokenService.On("GetToken", identifier).Return("", false)

		groupRepository := &kymaGroupMocks.Repository{}

		subjectValues := certificates.CSRSubject{
			CName:              identifier,
			Country:            country,
			Organization:       organization,
			OrganizationalUnit: organizationalUnit,
			Locality:           locality,
			Province:           province,
		}
		infoHandler := NewInfoHandler(tokenService, host, domain, subjectValues, groupRepository)

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(tokenRequestRaw))
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

		tokenService := &tokenMocks.Service{}
		tokenService.On("GetToken", identifier).Return("differentToken", true)

		groupRepository := &kymaGroupMocks.Repository{}

		subjectValues := certificates.CSRSubject{
			CName:              identifier,
			Country:            country,
			Organization:       organization,
			OrganizationalUnit: organizationalUnit,
			Locality:           locality,
			Province:           province,
		}
		infoHandler := NewInfoHandler(tokenService, host, domain, subjectValues, groupRepository)

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(tokenRequestRaw))
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

		tokenService := &tokenMocks.Service{}
		tokenService.On("GetToken", identifier).Return(token, true)
		tokenService.On("CreateToken", identifier).Return("", apperrors.Internal("error"))

		groupRepository := &kymaGroupMocks.Repository{}

		subjectValues := certificates.CSRSubject{
			CName:              identifier,
			Country:            country,
			Organization:       organization,
			OrganizationalUnit: organizationalUnit,
			Locality:           locality,
			Province:           province,
		}
		infoHandler := NewInfoHandler(tokenService, host, domain, subjectValues, groupRepository)

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(tokenRequestRaw))
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
