package externalapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-project/kyma/components/connector-service/internal/httperrors"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/certificates"
	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"
	tokenMocks "github.com/kyma-project/kyma/components/connector-service/internal/tokens/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	commonName  = "commonName"
	application = "application"
)

type dummyClientContext struct{}

func (dc dummyClientContext) ToJSON() ([]byte, error) {
	return []byte("test"), nil
}

func (dc dummyClientContext) GetApplication() string {
	return application
}

func (dc dummyClientContext) GetCommonName() string {
	return commonName
}

func (dc dummyClientContext) GetRuntimeUrls() *clientcontext.RuntimeURLs {
	return &clientcontext.RuntimeURLs{}
}

func TestCSRInfoHandler_GetCSRInfo(t *testing.T) {

	host := "connector-service.kyma.cx"
	infoURL := "https://connector-service.test.cluster.kyma.cx/v1/applications/management/info"

	urlApps := fmt.Sprintf("/v1/applications/signingRequests/info?token=%s", token)

	subjectValues := certificates.CSRSubject{
		Country:            country,
		Organization:       organization,
		OrganizationalUnit: organizationalUnit,
		Locality:           locality,
		Province:           province,
	}

	dummyClientContext := dummyClientContext{}
	connectorClientExtractor := func(ctx context.Context) (clientcontext.ConnectorClientContext, apperrors.AppError) {
		return dummyClientContext, nil
	}

	t.Run("should successfully get csr info", func(t *testing.T) {
		// given
		newToken := "newToken"
		expectedSignUrl := fmt.Sprintf("https://%s/v1/applications/certificates?token=%s", host, newToken)

		expectedCertInfo := certInfo{
			Subject:      fmt.Sprintf("OU=%s,O=%s,L=%s,ST=%s,C=%s,CN=%s", organizationalUnit, organization, locality, province, country, commonName),
			Extensions:   "",
			KeyAlgorithm: "rsa2048",
		}

		expectedAPI := api{
			CertificatesURL: fmt.Sprintf(AppURLFormat, host, CertsEndpoint),
			InfoURL:         infoURL,
		}

		tokenManager := &tokenMocks.Manager{}
		tokenManager.On("Replace", token, dummyClientContext).Return(newToken, nil)

		infoHandler := NewCSRInfoHandler(tokenManager, connectorClientExtractor, infoURL, host, subjectValues, AppURLFormat)

		req, err := http.NewRequest(http.MethodPost, urlApps, bytes.NewReader(tokenRequestRaw))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		// when
		infoHandler.GetCSRInfo(rr, req)

		// then
		responseBody, err := ioutil.ReadAll(rr.Body)
		require.NoError(t, err)

		var infoResponse csrInfoResponse
		err = json.Unmarshal(responseBody, &infoResponse)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, expectedSignUrl, infoResponse.CsrURL)
		assert.EqualValues(t, expectedAPI, infoResponse.API)
		assert.EqualValues(t, expectedCertInfo, infoResponse.CertificateInfo)
	})

	t.Run("should use predefined getInfoURL", func(t *testing.T) {
		// given
		newToken := "newToken"
		predefinedGetInfoURL := "https://predefined.test.cluster.kyma.cx/v1/applications/management/info"

		expectedAPI := api{
			InfoURL: predefinedGetInfoURL,
		}

		tokenManager := &tokenMocks.Manager{}
		tokenManager.On("Replace", token, dummyClientContext).Return(newToken, nil)

		infoHandler := NewCSRInfoHandler(tokenManager, connectorClientExtractor, predefinedGetInfoURL, host, subjectValues, AppURLFormat)

		req, err := http.NewRequest(http.MethodPost, urlApps, bytes.NewReader(tokenRequestRaw))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		// when
		infoHandler.GetCSRInfo(rr, req)

		// then
		responseBody, err := ioutil.ReadAll(rr.Body)
		require.NoError(t, err)

		var infoResponse csrInfoResponse
		err = json.Unmarshal(responseBody, &infoResponse)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.EqualValues(t, expectedAPI.InfoURL, infoResponse.API.InfoURL)
	})

	t.Run("should return 500 when failed to extract context", func(t *testing.T) {
		// given
		tokenManager := &tokenMocks.Manager{}

		errorExtractor := func(ctx context.Context) (clientcontext.ConnectorClientContext, apperrors.AppError) {
			return nil, apperrors.Internal("error")
		}

		infoHandler := NewCSRInfoHandler(tokenManager, errorExtractor, infoURL, host, subjectValues, AppURLFormat)

		req, err := http.NewRequest(http.MethodPost, urlApps, bytes.NewReader(tokenRequestRaw))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		// when
		infoHandler.GetCSRInfo(rr, req)

		// then
		responseBody, err := ioutil.ReadAll(rr.Body)
		require.NoError(t, err)

		var errorResponse httperrors.ErrorResponse
		err = json.Unmarshal(responseBody, &errorResponse)
		require.NoError(t, err)

		assert.Equal(t, http.StatusInternalServerError, errorResponse.Code)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("should return 500 when failed to replace token", func(t *testing.T) {
		// given
		tokenManager := &tokenMocks.Manager{}
		tokenManager.On("Replace", token, dummyClientContext).Return("", apperrors.Internal("error"))

		infoHandler := NewCSRInfoHandler(tokenManager, connectorClientExtractor, infoURL, host, subjectValues, AppURLFormat)

		req, err := http.NewRequest(http.MethodPost, urlApps, bytes.NewReader(tokenRequestRaw))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		// when
		infoHandler.GetCSRInfo(rr, req)

		// then
		responseBody, err := ioutil.ReadAll(rr.Body)
		require.NoError(t, err)

		var errorResponse httperrors.ErrorResponse
		err = json.Unmarshal(responseBody, &errorResponse)
		require.NoError(t, err)

		assert.Equal(t, http.StatusInternalServerError, errorResponse.Code)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("should retrieve metadata and events urls from context", func(t *testing.T) {
		//given
		expectedMetadataUrl := "https://metadata.base.path/application/v1/metadata/services"
		expectedEventsUrl := "https://events.base.path/application/v1/events"

		extendedCtx := &clientcontext.ExtendedApplicationContext{
			ApplicationContext: clientcontext.ApplicationContext{},
			RuntimeURLs: clientcontext.RuntimeURLs{
				MetadataURL: "https://metadata.base.path/application/v1/metadata/services",
				EventsURL:   "https://events.base.path/application/v1/events",
			},
		}

		connectorClientExtractor := func(ctx context.Context) (clientcontext.ConnectorClientContext, apperrors.AppError) {
			return *extendedCtx, nil
		}

		newToken := "newToken"
		tokenManager := &tokenMocks.Manager{}
		tokenManager.On("Replace", token, *extendedCtx).Return(newToken, nil)

		req, err := http.NewRequest(http.MethodGet, urlApps, nil)
		require.NoError(t, err)

		infoHandler := NewCSRInfoHandler(tokenManager, connectorClientExtractor, infoURL, host, subjectValues, AppURLFormat)

		rr := httptest.NewRecorder()

		//when
		infoHandler.GetCSRInfo(rr, req)

		//then
		responseBody, err := ioutil.ReadAll(rr.Body)
		require.NoError(t, err)

		var infoResponse csrInfoResponse
		err = json.Unmarshal(responseBody, &infoResponse)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, rr.Code)
		api := infoResponse.API
		assert.Equal(t, expectedMetadataUrl, api.MetadataURL)
		assert.Equal(t, expectedEventsUrl, api.EventsURL)
	})
}
