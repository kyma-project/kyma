package apitests

import (
	"crypto/rsa"
	"net/http"
	"net/url"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/kyma/tests/connector-service-tests/test/testkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	retryWaitTimeSeconds = 5 * time.Second
	retryCount           = 20

	baseEventsHostHeader   = "EventsHost"
	baseMetadataHostHeader = "MetadataHost"
)

func TestConnector(t *testing.T) {

	config, err := testkit.ReadConfig()
	require.NoError(t, err)

	k8sResourcesClient, err := testkit.NewK8sResourcesClient()
	require.NoError(t, err)
	app, e := k8sResourcesClient.CreateDummyApplication("app-connector-test-0", "", true)
	require.NoError(t, e)

	defer func() {
		k8sResourcesClient.DeleteApplication(app.Name, &v1.DeleteOptions{})
	}()

	t.Run("Connector Service flow for Application", func(t *testing.T) {
		appName := "testCertGenApp"
		appTokenRequest := createApplicationTokenRequest(t, config, appName)
		certificateGenerationSuite(t, appTokenRequest, config.SkipSslVerify)

		appName = "testCSRInfoApp"
		appTokenRequest = createApplicationTokenRequest(t, config, appName)
		getAppCsrInfoEndpointSuite(t, appTokenRequest, config.SkipSslVerify, config.GatewayUrl, appName)

		appName = "testMgmInfoApp"
		appTokenRequest = createApplicationTokenRequest(t, config, appName)
		getAppMgmInfoEndpointSuite(t, appTokenRequest, config.SkipSslVerify, config.GatewayUrl, appName)
	})

	t.Run("Connector Service flow for Runtime", func(t *testing.T) {
		runtimeTokenRequest := createRuntimeTokenRequest(t, config)
		certificateGenerationSuite(t, runtimeTokenRequest, config.SkipSslVerify)

		runtimeTokenRequest = createRuntimeTokenRequest(t, config)
		getRuntimeCsrInfoEndpointSuite(t, runtimeTokenRequest, config.SkipSslVerify)

		runtimeTokenRequest = createRuntimeTokenRequest(t, config)
		getRuntimeMgmInfoEndpointSuite(t, runtimeTokenRequest, config.SkipSslVerify, config.GatewayUrl)
	})
}

func createApplicationTokenRequest(t *testing.T, config testkit.TestConfig, appName string) *http.Request {
	tokenURL := config.InternalAPIUrl + "/v1/applications/tokens"

	request := createTokenRequest(t, tokenURL, config)
	request.Header.Set(testkit.ApplicationHeader, appName)

	return request
}

func createRuntimeTokenRequest(t *testing.T, config testkit.TestConfig) *http.Request {
	tokenURL := config.InternalAPIUrl + "/v1/runtimes/tokens"

	request := createTokenRequest(t, tokenURL, config)

	return request
}

func createTokenRequest(t *testing.T, tokenURL string, config testkit.TestConfig) *http.Request {
	request, err := http.NewRequest(http.MethodPost, tokenURL, nil)
	require.NoError(t, err)

	if config.Group != "" {
		request.Header.Set(testkit.GroupHeader, config.Group)
	}

	if config.Tenant != "" {
		request.Header.Set(testkit.TenantHeader, config.Tenant)
	}

	return request
}

func certificateGenerationSuite(t *testing.T, tokenRequest *http.Request, skipVerify bool) {

	client := testkit.NewConnectorClient(tokenRequest, skipVerify)

	clientKey := testkit.CreateKey(t)

	t.Run("should create client certificate", func(t *testing.T) {
		// when
		crtResponse, infoResponse := createCertificateChain(t, client, clientKey)

		//then
		require.NotEmpty(t, crtResponse.CRTChain)

		// when
		certificates := testkit.DecodeAndParseCerts(t, crtResponse)

		// then
		clientsCrt := certificates.CRTChain[0]
		testkit.CheckIfSubjectEquals(t, infoResponse.Certificate.Subject, clientsCrt)
	})

	t.Run("should create two certificates in a chain", func(t *testing.T) {
		// when
		crtResponse, _ := createCertificateChain(t, client, clientKey)

		//then
		require.NotEmpty(t, crtResponse.CRTChain)

		// when
		certificates := testkit.DecodeAndParseCerts(t, crtResponse)

		// then
		require.Equal(t, 2, len(certificates.CRTChain))
	})

	t.Run("client cert should be signed by server cert", func(t *testing.T) {
		//when
		crtResponse, _ := createCertificateChain(t, client, clientKey)

		//then
		require.NotEmpty(t, crtResponse.CRTChain)

		// when
		certificates := testkit.DecodeAndParseCerts(t, crtResponse)

		//then
		testkit.CheckIfCertIsSigned(t, certificates.CRTChain)
	})

	t.Run("should respond with client certificate together with CA crt", func(t *testing.T) {
		// when
		crtResponse, infoResponse := createCertificateChain(t, client, clientKey)

		//then
		require.NotEmpty(t, crtResponse.CRTChain)

		// when
		certificates := testkit.DecodeAndParseCerts(t, crtResponse)

		// then
		clientsCrt := certificates.CRTChain[0]
		testkit.CheckIfSubjectEquals(t, infoResponse.Certificate.Subject, clientsCrt)
		require.Equal(t, certificates.ClientCRT, clientsCrt)

		caCrt := certificates.CRTChain[1]
		require.Equal(t, certificates.CaCRT, caCrt)
	})

	t.Run("should validate CSR subject", func(t *testing.T) {
		// when
		tokenResponse := client.CreateToken(t)

		// then
		require.NotEmpty(t, tokenResponse.Token)
		require.Contains(t, tokenResponse.URL, "token="+tokenResponse.Token)

		// when
		infoResponse, errorResponse := client.GetInfo(t, tokenResponse.URL, nil)

		// then
		require.Nil(t, errorResponse)
		require.NotEmpty(t, infoResponse.CertUrl)
		require.Equal(t, "rsa2048", infoResponse.Certificate.KeyAlgorithm)

		// given
		infoResponse.Certificate.Subject = "subject=OU=Test,O=Test,L=Wrong,ST=Wrong,C=PL,CN=Wrong"
		csr := testkit.CreateCsr(t, infoResponse.Certificate, clientKey)
		csrBase64 := testkit.EncodeBase64(csr)

		// when
		_, err := client.CreateCertChain(t, csrBase64, infoResponse.CertUrl)

		// then
		require.NotNil(t, err)
		require.Equal(t, http.StatusBadRequest, err.StatusCode)
		require.Equal(t, http.StatusBadRequest, err.ErrorResponse.Code)
		require.Equal(t, "CSR: Invalid common name provided.", err.ErrorResponse.Error)
	})

	t.Run("should return error for wrong token on info endpoint", func(t *testing.T) {
		// when
		tokenResponse := client.CreateToken(t)

		// then
		require.NotEmpty(t, tokenResponse.Token)
		require.Contains(t, tokenResponse.URL, "token="+tokenResponse.Token)

		wrongUrl := replaceToken(tokenResponse.URL, "incorrect-token")

		// when
		_, err := client.GetInfo(t, wrongUrl, nil)

		// then
		require.NotNil(t, err)
		require.Equal(t, http.StatusForbidden, err.StatusCode)
		require.Equal(t, http.StatusForbidden, err.ErrorResponse.Code)
		require.Equal(t, "Invalid token.", err.ErrorResponse.Error)
	})

	t.Run("should return error for wrong token on client-certs", func(t *testing.T) {
		// when
		tokenResponse := client.CreateToken(t)

		// then
		require.NotEmpty(t, tokenResponse.Token)
		require.Contains(t, tokenResponse.URL, "token="+tokenResponse.Token)

		// when
		infoResponse, errorResponse := client.GetInfo(t, tokenResponse.URL, nil)

		// then
		require.Nil(t, errorResponse)
		require.NotEmpty(t, infoResponse.CertUrl)

		wrongUrl := replaceToken(infoResponse.CertUrl, "incorrect-token")

		// then
		require.Nil(t, errorResponse)
		require.NotEmpty(t, infoResponse.CertUrl)
		require.Equal(t, "rsa2048", infoResponse.Certificate.KeyAlgorithm)

		// when
		_, err := client.CreateCertChain(t, "csr", wrongUrl)

		// then
		require.NotNil(t, err)
		require.Equal(t, http.StatusForbidden, err.StatusCode)
		require.Equal(t, http.StatusForbidden, err.ErrorResponse.Code)
		require.Equal(t, "Invalid token.", err.ErrorResponse.Error)
	})

	t.Run("should return error on wrong CSR on client-certs", func(t *testing.T) {
		// when
		tokenResponse := client.CreateToken(t)

		// then
		require.NotEmpty(t, tokenResponse.Token)
		require.Contains(t, tokenResponse.URL, "token="+tokenResponse.Token)

		// when
		infoResponse, errorResponse := client.GetInfo(t, tokenResponse.URL, nil)

		// then
		require.Nil(t, errorResponse)
		require.NotEmpty(t, infoResponse.CertUrl)
		require.Equal(t, "rsa2048", infoResponse.Certificate.KeyAlgorithm)

		// when
		_, err := client.CreateCertChain(t, "wrong-csr", infoResponse.CertUrl)

		// then
		require.NotNil(t, err)
		require.Equal(t, http.StatusBadRequest, err.StatusCode)
		require.Equal(t, http.StatusBadRequest, err.ErrorResponse.Code)
		require.Equal(t, "There was an error while parsing the base64 content. An incorrect value was provided.", err.ErrorResponse.Error)
	})

}

func getAppCsrInfoEndpointSuite(t *testing.T, tokenRequest *http.Request, skipVerify bool, defaultGatewayUrl string, appName string) {

	client := testkit.NewConnectorClient(tokenRequest, skipVerify)

	t.Run("should use headers to build CSR info response", func(t *testing.T) {
		// given
		metadataHost := "metadata.kyma.test.cx"
		eventsHost := "events.kyma.test.cx"

		expectedMetadataURL := "https://metadata.kyma.test.cx/" + appName + "/v1/metadata/services"
		expectedEventsURL := "https://events.kyma.test.cx/" + appName + "/v1/events"

		// when
		tokenResponse := client.CreateToken(t)

		// then
		require.NotEmpty(t, tokenResponse.Token)
		require.Contains(t, tokenResponse.URL, "token="+tokenResponse.Token)

		// when
		infoResponse, errorResponse := client.GetInfo(t, tokenResponse.URL, createHostsHeaders(metadataHost, eventsHost))

		// then
		// TODO - should check ManagementInfoURL when its returned properly
		require.Nil(t, errorResponse)
		assert.Equal(t, expectedEventsURL, infoResponse.Api.EventsURL)
		assert.Equal(t, expectedMetadataURL, infoResponse.Api.MetadataURL)
	})

	t.Run("should use default values to build CSR info response when headers are not given", func(t *testing.T) {
		// given
		expectedMetadataURL := defaultGatewayUrl
		expectedEventsURL := defaultGatewayUrl

		if defaultGatewayUrl != "" {
			expectedMetadataURL += "/" + appName + "/v1/metadata/services"
			expectedEventsURL += "/" + appName + "/v1/events"
		}

		// when
		tokenResponse := client.CreateToken(t)

		// then
		require.NotEmpty(t, tokenResponse.Token)
		require.Contains(t, tokenResponse.URL, "token="+tokenResponse.Token)

		// when
		infoResponse, errorResponse := client.GetInfo(t, tokenResponse.URL, nil)

		// then
		require.Nil(t, errorResponse)
		assert.Equal(t, expectedEventsURL, infoResponse.Api.EventsURL)
		assert.Equal(t, expectedMetadataURL, infoResponse.Api.MetadataURL)
	})

	// TODO - test with empty values
}

func getRuntimeCsrInfoEndpointSuite(t *testing.T, tokenRequest *http.Request, skipVerify bool) {

	client := testkit.NewConnectorClient(tokenRequest, skipVerify)

	t.Run("should provide not empty CSR info response", func(t *testing.T) {
		// when
		tokenResponse := client.CreateToken(t)

		// then
		require.NotEmpty(t, tokenResponse.Token)
		require.Contains(t, tokenResponse.URL, "token="+tokenResponse.Token)

		// when
		infoResponse, errorResponse := client.GetInfo(t, tokenResponse.URL, nil)

		// then
		// TODO - should check ManagementInfoURL when its returned properly
		require.Nil(t, errorResponse)
		assert.NotEmpty(t, infoResponse.CertUrl)
		assert.NotEmpty(t, infoResponse.Api)
		assert.NotEmpty(t, infoResponse.Certificate)
	})
}

func getAppMgmInfoEndpointSuite(t *testing.T, tokenRequest *http.Request, skipVerify bool, defaultGatewayUrl string, appName string) {

	client := testkit.NewConnectorClient(tokenRequest, skipVerify)

	clientKey := testkit.CreateKey(t)

	t.Run("should use headers to build management info response", func(t *testing.T) {
		// given
		metadataHost := "metadata.kyma.test.cx"
		eventsHost := "events.kyma.test.cx"

		expectedMetadataURL := "https://metadata.kyma.test.cx/" + appName + "/v1/metadata/services"
		expectedEventsURL := "https://events.kyma.test.cx/" + appName + "/v1/events"

		// when
		crtResponse, infoResponse := createCertificateChain(t, client, clientKey)

		// then
		require.NotEmpty(t, crtResponse.CRTChain)
		require.NotEmpty(t, infoResponse.Api.GetInfoURL)

		certificates := testkit.DecodeAndParseCerts(t, crtResponse)
		client := testkit.NewSecuredConnectorClient(skipVerify, clientKey, certificates.ClientCRT.Raw)

		// when
		mgmInfoResponse, errorResponse := client.GetMgmInfo(t, infoResponse.Api.GetInfoURL, createHostsHeaders(metadataHost, eventsHost))
		require.Nil(t, errorResponse)

		// then
		assert.Equal(t, expectedMetadataURL, mgmInfoResponse.MetadataUrl)
		assert.Equal(t, expectedEventsURL, mgmInfoResponse.EventsUrl)
	})

	t.Run("should use default values to build management info response when headers are not given", func(t *testing.T) {
		// given
		expectedMetadataURL := defaultGatewayUrl
		expectedEventsURL := defaultGatewayUrl

		if defaultGatewayUrl != "" {
			expectedMetadataURL += "/" + appName + "/v1/metadata/services"
			expectedEventsURL += "/" + appName + "/v1/events"
		}

		// when
		crtResponse, infoResponse := createCertificateChain(t, client, clientKey)

		// then
		require.NotEmpty(t, crtResponse.CRTChain)
		require.NotEmpty(t, infoResponse.Api.GetInfoURL)

		certificates := testkit.DecodeAndParseCerts(t, crtResponse)
		client := testkit.NewSecuredConnectorClient(skipVerify, clientKey, certificates.ClientCRT.Raw)

		// when
		mgmInfoResponse, errorResponse := client.GetMgmInfo(t, infoResponse.Api.GetInfoURL, nil)
		require.Nil(t, errorResponse)

		// then
		assert.Equal(t, expectedMetadataURL, mgmInfoResponse.MetadataUrl)
		assert.Equal(t, expectedEventsURL, mgmInfoResponse.EventsUrl)
	})

	// TODO - test with empty headers when implemented
}

func getRuntimeMgmInfoEndpointSuite(t *testing.T, tokenRequest *http.Request, skipVerify bool, defaultGatewayUrl string) {

	client := testkit.NewConnectorClient(tokenRequest, skipVerify)

	clientKey := testkit.CreateKey(t)

	t.Run("should provide not empty management info response", func(t *testing.T) {
		// when
		crtResponse, infoResponse := createCertificateChain(t, client, clientKey)

		// then
		require.NotEmpty(t, crtResponse.CRTChain)
		require.NotEmpty(t, infoResponse.Api.GetInfoURL)

		certificates := testkit.DecodeAndParseCerts(t, crtResponse)
		client := testkit.NewSecuredConnectorClient(skipVerify, clientKey, certificates.ClientCRT.Raw)

		// when
		mgmInfoResponse, errorResponse := client.GetMgmInfo(t, infoResponse.Api.GetInfoURL, nil)
		require.Nil(t, errorResponse)

		// then
		assert.NotEmpty(t, mgmInfoResponse)
	})

}

func createCertificateChain(t *testing.T, connectorClient testkit.ConnectorClient, key *rsa.PrivateKey) (*testkit.CrtResponse, *testkit.InfoResponse) {
	// when
	tokenResponse := connectorClient.CreateToken(t)

	// then
	require.NotEmpty(t, tokenResponse.Token)
	require.Contains(t, tokenResponse.URL, "token="+tokenResponse.Token)

	// when
	infoResponse, errorResponse := connectorClient.GetInfo(t, tokenResponse.URL, nil)

	// then
	require.Nil(t, errorResponse)
	require.NotEmpty(t, infoResponse.CertUrl)
	require.Equal(t, "rsa2048", infoResponse.Certificate.KeyAlgorithm)

	// given
	csr := testkit.CreateCsr(t, infoResponse.Certificate, key)
	csrBase64 := testkit.EncodeBase64(csr)

	// when
	crtResponse, errorResponse := connectorClient.CreateCertChain(t, csrBase64, infoResponse.CertUrl)

	// then
	require.Nil(t, errorResponse)

	return crtResponse, infoResponse
}

func replaceToken(originalUrl string, newToken string) string {
	parsedUrl, _ := url.Parse(originalUrl)
	queryParams, _ := url.ParseQuery(parsedUrl.RawQuery)

	queryParams.Set("token", newToken)
	parsedUrl.RawQuery = queryParams.Encode()

	return parsedUrl.String()
}

func createHostsHeaders(metadataHost string, eventsHost string) map[string]string {
	return map[string]string{
		baseMetadataHostHeader: metadataHost,
		baseEventsHostHeader:   eventsHost,
	}
}
