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
	commonName     = "commonName"
	application    = "application"
	certificateURL = "https://connector-service.kyma.cx/v1/applications/certificates"
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

func TestInfoHandler_GetCSRInfo(t *testing.T) {

	host := "connector-service.kyma.cx"
	infoURL := "https://connector-service.test.cluster.kyma.cx/v1/applications/management/info"

	url := fmt.Sprintf("/v1/applications/signingRequests/info?token=%s", token)

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
			InfoURL:         fmt.Sprintf(AppURLFormat, host, ManagementInfoEndpoint),
		}

		tokenCreator := &tokenMocks.Creator{}
		tokenCreator.On("Replace", token, dummyClientContext).Return(newToken, nil)

		infoHandler := NewCSRInfoHandler(tokenCreator, connectorClientExtractor, certificateURL, "", host, subjectValues, AppURLFormat)

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(tokenRequestRaw))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		// when
		infoHandler.GetCSRInfo(rr, req)

		// then
		responseBody, err := ioutil.ReadAll(rr.Body)
		require.NoError(t, err)

		var infoResponse infoResponse
		infoResponse.API = &api{}
		err = json.Unmarshal(responseBody, &infoResponse)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, expectedSignUrl, infoResponse.CsrURL)
		assert.EqualValues(t, expectedAPI.CertificatesURL, infoResponse.API.(*api).CertificatesURL)
		assert.EqualValues(t, expectedAPI.InfoURL, infoResponse.API.(*api).InfoURL)
		assert.EqualValues(t, expectedCertInfo, infoResponse.CertificateInfo)
	})

	t.Run("should use predefined getInfoURL", func(t *testing.T) {
		// given
		newToken := "newToken"
		predefinedGetInfoURL := "https://predefined.test.cluster.kyma.cx/v1/applications/management/info"

		expectedAPI := api{
			InfoURL: predefinedGetInfoURL,
		}

		tokenCreator := &tokenMocks.Creator{}
		tokenCreator.On("Replace", token, dummyClientContext).Return(newToken, nil)

		infoHandler := NewCSRInfoHandler(tokenCreator, connectorClientExtractor, certificateURL, predefinedGetInfoURL, host, subjectValues, AppURLFormat)

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(tokenRequestRaw))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		// when
		infoHandler.GetCSRInfo(rr, req)

		// then
		responseBody, err := ioutil.ReadAll(rr.Body)
		require.NoError(t, err)

		var infoResponse infoResponse
		infoResponse.API = &api{}
		err = json.Unmarshal(responseBody, &infoResponse)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.EqualValues(t, expectedAPI.InfoURL, infoResponse.API.(*api).InfoURL)
	})

	t.Run("should return 500 when failed to extract context", func(t *testing.T) {
		// given
		tokenCreator := &tokenMocks.Creator{}

		errorExtractor := func(ctx context.Context) (clientcontext.ConnectorClientContext, apperrors.AppError) {
			return nil, apperrors.Internal("error")
		}

		infoHandler := NewCSRInfoHandler(tokenCreator, errorExtractor, certificateURL, infoURL, host, subjectValues, AppURLFormat)

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(tokenRequestRaw))
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
		tokenCreator := &tokenMocks.Creator{}
		tokenCreator.On("Replace", token, dummyClientContext).Return("", apperrors.Internal("error"))

		infoHandler := NewCSRInfoHandler(tokenCreator, connectorClientExtractor, certificateURL, infoURL, host, subjectValues, AppURLFormat)

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(tokenRequestRaw))
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

	t.Run("should use headers if present", func(t *testing.T) {
		//given
		newToken := "newToken"
		tokenCreator := &tokenMocks.Creator{}
		tokenCreator.On("Replace", token, dummyClientContext).Return(newToken, nil)

		req, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err)
		req.Header.Set(BaseEventsPathHeader, "events.base.path")
		req.Header.Set(BaseMetadataPathHeader, "metadata.base.path")
		expectedMetadataUrl := "https://metadata.base.path/application/v1/metadata/services"
		expectedEventsUrl := "https://events.base.path/application/v1/events"

		infoHandler := NewCSRInfoHandler(tokenCreator, connectorClientExtractor, certificateURL, infoURL, host, subjectValues, AppURLFormat)

		rr := httptest.NewRecorder()

		//when
		infoHandler.GetCSRInfo(rr, req)

		//then
		responseBody, err := ioutil.ReadAll(rr.Body)
		require.NoError(t, err)

		var infoResponse infoResponse
		infoResponse.API = &api{}
		err = json.Unmarshal(responseBody, &infoResponse)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, rr.Code)
		api := infoResponse.API.(*api)
		assert.Equal(t, expectedMetadataUrl, api.MetadataURL)
		assert.Equal(t, expectedEventsUrl, api.EventsURL)
	})

	t.Run("should use default value for empty header", func(t *testing.T) {
		//given
		newToken := "newToken"
		tokenCreator := &tokenMocks.Creator{}
		tokenCreator.On("Replace", token, dummyClientContext).Return(newToken, nil)

		req, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err)
		req.Header.Add(BaseEventsPathHeader, "events.base.path")
		req.Header.Add(BaseMetadataPathHeader, "")
		expectedMetadataUrl := "https://metadata.default.path/application/v1/metadata/services"
		expectedEventsUrl := "https://events.base.path/application/v1/events"

		infoHandler := NewCSRInfoHandler(tokenCreator, connectorClientExtractor, certificateURL, infoURL, host, subjectValues, AppURLFormat)

		rr := httptest.NewRecorder()

		//when
		infoHandler.GetCSRInfo(rr, req)

		//then
		responseBody, err := ioutil.ReadAll(rr.Body)
		require.NoError(t, err)

		var infoResponse infoResponse
		infoResponse.API = &api{}
		err = json.Unmarshal(responseBody, &infoResponse)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, rr.Code)
		api := infoResponse.API.(*api)
		assert.Equal(t, expectedMetadataUrl, api.MetadataURL)
		assert.Equal(t, expectedEventsUrl, api.EventsURL)
	})

	t.Run("should use default value for non-existing header", func(t *testing.T) {
		//given
		newToken := "newToken"
		tokenCreator := &tokenMocks.Creator{}
		tokenCreator.On("Replace", token, dummyClientContext).Return(newToken, nil)

		req, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err)
		req.Header.Add(BaseEventsPathHeader, "events.base.path")
		expectedMetadataUrl := "https://metadata.default.path/application/v1/metadata/services"
		expectedEventsUrl := "https://events.base.path/application/v1/events"

		infoHandler := NewCSRInfoHandler(tokenCreator, connectorClientExtractor, certificateURL, infoURL, host, subjectValues, AppURLFormat)

		rr := httptest.NewRecorder()

		//when
		infoHandler.GetCSRInfo(rr, req)

		//then
		responseBody, err := ioutil.ReadAll(rr.Body)
		require.NoError(t, err)

		var infoResponse infoResponse
		infoResponse.API = &api{}
		err = json.Unmarshal(responseBody, &infoResponse)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, rr.Code)
		api := infoResponse.API.(*api)
		assert.Equal(t, expectedMetadataUrl, api.MetadataURL)
		assert.Equal(t, expectedEventsUrl, api.EventsURL)
	})

	t.Run("should return empty url when both header and default value are empty", func(t *testing.T) {
		//given
		newToken := "newToken"
		tokenCreator := &tokenMocks.Creator{}
		tokenCreator.On("Replace", token, dummyClientContext).Return(newToken, nil)

		req, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err)
		req.Header.Add(BaseEventsPathHeader, "events.base.path")
		req.Header.Add(BaseMetadataPathHeader, "")
		expectedMetadataUrl := ""
		expectedEventsUrl := "https://events.base.path/application/v1/events"

		infoHandler := NewCSRInfoHandler(tokenCreator, connectorClientExtractor, certificateURL, infoURL, host, subjectValues, AppURLFormat)

		rr := httptest.NewRecorder()

		//when
		infoHandler.GetCSRInfo(rr, req)

		//then
		responseBody, err := ioutil.ReadAll(rr.Body)
		require.NoError(t, err)

		var infoResponse infoResponse
		infoResponse.API = &api{}
		err = json.Unmarshal(responseBody, &infoResponse)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, rr.Code)
		api := infoResponse.API.(*api)
		assert.Equal(t, expectedMetadataUrl, api.MetadataURL)
		assert.Equal(t, expectedEventsUrl, api.EventsURL)
	})
}
