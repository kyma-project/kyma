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

	"github.com/kyma-project/kyma/components/connector-service/internal/certificates"

	"github.com/sirupsen/logrus"

	"github.com/kyma-project/kyma/components/connector-service/internal/httperrors"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"
	tokenMocks "github.com/kyma-project/kyma/components/connector-service/internal/tokens/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	subject = certificates.CSRSubject{
		Organization:       "Org",
		OrganizationalUnit: "OrgUnit",
		Country:            "PL",
		Province:           "Province",
		Locality:           "Gliwice",
		CommonName:         "CommonName",
	}

	strSubject = "O=Org,OU=OrgUnit,L=Gliwice,ST=Province,C=PL,CN=CommonName"
)

type dummyClientContextService struct{}

func (dc dummyClientContextService) ToJSON() ([]byte, error) {
	return []byte("test"), nil
}

func (dc dummyClientContextService) GetRuntimeUrls() *clientcontext.RuntimeURLs {
	return nil
}

func (dc dummyClientContextService) GetLogger() *logrus.Entry {
	return logrus.WithFields(logrus.Fields{})
}

type dummyClientContextServiceWithEmptyURLs struct {
	dummyClientContextService
}

func (dc dummyClientContextServiceWithEmptyURLs) GetRuntimeUrls() *clientcontext.RuntimeURLs {
	return &clientcontext.RuntimeURLs{
		MetadataURL: "",
		EventsURL:   "",
	}
}

type dummyClientCertCtx struct {
	clientcontext.ClientContextService
}

func (cc dummyClientCertCtx) GetSubject() certificates.CSRSubject {
	return subject
}

func (cc dummyClientCertCtx) ClientContext() clientcontext.ClientContextService {
	return cc.ClientContextService
}

func TestCSRInfoHandler_GetCSRInfo(t *testing.T) {

	baseURL := "https://connector-service.kyma.cx/v1/applications"
	infoURL := "https://connector-service.test.cluster.kyma.cx/v1/applications/management/info"
	newToken := "newToken"

	urlApps := fmt.Sprintf("/v1/applications/signingRequests/info?token=%s", token)

	dummyClientContextService := dummyClientContextService{}
	clientContextService := func(ctx context.Context) (clientcontext.ClientCertContextService, apperrors.AppError) {
		return dummyClientCertCtx{dummyClientContextService}, nil
	}

	expectedSignUrl := fmt.Sprintf("%s/certificates?token=%s", baseURL, newToken)

	expectedCertInfo := certInfo{
		Subject:      strSubject,
		Extensions:   "",
		KeyAlgorithm: "rsa2048",
	}

	t.Run("should successfully get csr info", func(t *testing.T) {
		// given
		expectedAPI := api{
			CertificatesURL: baseURL + CertsEndpoint,
			InfoURL:         infoURL,
		}

		tokenCreator := &tokenMocks.Creator{}
		tokenCreator.On("Save", dummyClientContextService).Return(newToken, nil)

		infoHandler := NewCSRInfoHandler(tokenCreator, clientContextService, infoURL, baseURL)

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

	t.Run("should get csr info with empty URLs", func(t *testing.T) {
		// given
		expectedAPI := api{
			CertificatesURL: baseURL + CertsEndpoint,
			InfoURL:         infoURL,
			RuntimeURLs: &clientcontext.RuntimeURLs{
				MetadataURL: "",
				EventsURL:   "",
			},
		}

		dummyClientContextServiceWithEmptyURLs := &dummyClientContextServiceWithEmptyURLs{dummyClientContextService: dummyClientContextService}
		clientContextService := func(ctx context.Context) (clientcontext.ClientCertContextService, apperrors.AppError) {
			return dummyClientCertCtx{dummyClientContextServiceWithEmptyURLs}, nil
		}

		tokenCreator := &tokenMocks.Creator{}
		tokenCreator.On("Save", dummyClientContextServiceWithEmptyURLs).Return(newToken, nil)

		infoHandler := NewCSRInfoHandler(tokenCreator, clientContextService, infoURL, baseURL)

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
		predefinedGetInfoURL := "https://predefined.test.cluster.kyma.cx/v1/applications/management/info"

		expectedAPI := api{
			InfoURL: predefinedGetInfoURL,
		}

		tokenCreator := &tokenMocks.Creator{}
		tokenCreator.On("Save", dummyClientContextService).Return(newToken, nil)

		infoHandler := NewCSRInfoHandler(tokenCreator, clientContextService, predefinedGetInfoURL, baseURL)

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
		tokenCreator := &tokenMocks.Creator{}

		errorExtractor := func(ctx context.Context) (clientcontext.ClientCertContextService, apperrors.AppError) {
			return nil, apperrors.Internal("error")
		}

		infoHandler := NewCSRInfoHandler(tokenCreator, errorExtractor, infoURL, baseURL)

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

	t.Run("should return 500 when failed to save token", func(t *testing.T) {
		// given
		tokenCreator := &tokenMocks.Creator{}
		tokenCreator.On("Save", dummyClientContextService).Return("", apperrors.Internal("error"))

		infoHandler := NewCSRInfoHandler(tokenCreator, clientContextService, infoURL, baseURL)

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

		clientContextService := func(ctx context.Context) (clientcontext.ClientCertContextService, apperrors.AppError) {
			return dummyClientCertCtx{extendedCtx}, nil
		}

		tokenCreator := &tokenMocks.Creator{}
		tokenCreator.On("Save", extendedCtx).Return(newToken, nil)

		req, err := http.NewRequest(http.MethodGet, urlApps, nil)
		require.NoError(t, err)

		infoHandler := NewCSRInfoHandler(tokenCreator, clientContextService, infoURL, baseURL)

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
