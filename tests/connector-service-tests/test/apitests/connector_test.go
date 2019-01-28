package apitests

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1"

	"io/ioutil"

	"gopkg.in/yaml.v2"

	"github.com/kyma-project/kyma/tests/connector-service-tests/test/testkit"
	"github.com/stretchr/testify/require"
)

const (
	retryWaitTimeSeconds = 5 * time.Second
	retryCount           = 20
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
		appTokenRequest := createApplicationTokenRequest(t, config, "test")
		CertificateGenerationSuite(t, appTokenRequest, config.SkipSslVerify)
	})

	t.Run("Connector Service flow for Runtime", func(t *testing.T) {
		runtimeTokenRequest := createRuntimeTokenRequest(t, config)
		CertificateGenerationSuite(t, runtimeTokenRequest, config.SkipSslVerify)
	})
}

func TestApiSpec(t *testing.T) {

	apiSpecPath := "/v1/api.yaml"

	config, err := testkit.ReadConfig()
	require.NoError(t, err)

	hc := testkit.NewHttpClient(config.SkipSslVerify)

	t.Run("should receive api spec", func(t *testing.T) {
		// when
		response, err := hc.Get(config.ExternalAPIUrl + apiSpecPath)

		// then
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, response.StatusCode)

		// when
		body, err := ioutil.ReadAll(response.Body)

		// then
		require.NoError(t, err)

		var apiSpec struct{}
		err = yaml.Unmarshal(body, &apiSpec)
		require.NoError(t, err)
	})

	t.Run("should receive 301 when accessing base path", func(t *testing.T) {
		// given
		hc.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			require.Equal(t, apiSpecPath, req.URL.Path)
			return http.ErrUseLastResponse
		}

		// when
		response, err := hc.Get(config.ExternalAPIUrl + "/v1")

		// then
		require.NoError(t, err)
		require.Equal(t, http.StatusMovedPermanently, response.StatusCode)
	})
}

func TestCertificateValidation(t *testing.T) {

	config, err := testkit.ReadConfig()
	require.NoError(t, err)

	gatewayUrlFormat := config.GatewayUrl + "/%s/v1/metadata/services"

	k8sResourcesClient, err := testkit.NewK8sResourcesClient()
	require.NoError(t, err)
	testApp, err := k8sResourcesClient.CreateDummyApplication("app-connector-test-1", "", false)
	require.NoError(t, err)
	defer func() {
		k8sResourcesClient.DeleteApplication(testApp.Name, &v1.DeleteOptions{})
	}()

	t.Run("should access application", func(t *testing.T) {
		// given
		tokenRequest := createApplicationTokenRequest(t, config, testApp.Name)
		connectorClient := testkit.NewConnectorClient(tokenRequest, config.SkipSslVerify)
		tlsClient := createTLSClientWithCert(t, connectorClient, config.SkipSslVerify)

		// when
		response, err := repeatUntilIngressIsCreated(tlsClient, gatewayUrlFormat, testApp.Name)

		// then
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, response.StatusCode)
	})

	t.Run("should receive 403 when accessing Application with invalid CN", func(t *testing.T) {
		// given
		tokenRequest := createApplicationTokenRequest(t, config, "another-application")
		connectorClient := testkit.NewConnectorClient(tokenRequest, config.SkipSslVerify)
		tlsClient := createTLSClientWithCert(t, connectorClient, config.SkipSslVerify)

		// when
		response, err := repeatUntilIngressIsCreated(tlsClient, gatewayUrlFormat, testApp.Name)

		// then
		require.NoError(t, err)
		require.Equal(t, http.StatusForbidden, response.StatusCode)
	})

}

func repeatUntilIngressIsCreated(tlsClient *http.Client, gatewayUrlFormat string, appName string) (*http.Response, error) {
	var response *http.Response
	var err error
	for i := 0; (shouldRetry(response, err)) && i < retryCount; i++ {
		response, err = tlsClient.Get(fmt.Sprintf(gatewayUrlFormat, appName))
		time.Sleep(retryWaitTimeSeconds)
	}
	return response, err
}

func shouldRetry(response *http.Response, err error) bool {
	return response == nil || http.StatusNotFound == response.StatusCode || err != nil
}

func createTLSClientWithCert(t *testing.T, client testkit.ConnectorClient, skipVerify bool) *http.Client {
	key := testkit.CreateKey(t)

	crtResponse, _ := createCertificateChain(t, client, key)
	require.NotEmpty(t, crtResponse.CRTChain)
	clientCertBytes := testkit.EncodedCertToPemBytes(t, crtResponse.ClientCRT)

	tlsCert := tls.Certificate{
		Certificate: [][]byte{clientCertBytes},
		PrivateKey:  key,
	}

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{tlsCert},
		ClientAuth:         tls.RequireAndVerifyClientCert,
		InsecureSkipVerify: skipVerify,
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	return &http.Client{
		Transport: transport,
	}
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
