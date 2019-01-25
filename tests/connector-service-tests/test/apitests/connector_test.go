package apitests

import (
	"crypto/rsa"
	"crypto/tls"
	"fmt"
	. "net/http"
	"net/url"
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
	app, e := k8sResourcesClient.CreateDummyApplication("app-connector-test-0", "", false)
	require.NoError(t, e)

	defer func() {
		k8sResourcesClient.DeleteApplication(app.Name, &v1.DeleteOptions{})
	}()

	client := testkit.NewConnectorClient(app.Name, config.InternalAPIUrl, config.ExternalAPIUrl, config.SkipSslVerify)

	clientKey := testkit.CreateKey(t)

	t.Run("should create client certificate", func(t *testing.T) {
		// when
		crtResponse, infoResponse := createCertificateChain(t, client, clientKey)

		//then
		require.NotEmpty(t, crtResponse.Crt)

		// when
		certificates := testkit.DecodeAndParseCert(t, crtResponse)

		// then
		clientsCrt := certificates[0]
		testkit.CheckIfSubjectEquals(t, infoResponse.Certificate.Subject, clientsCrt)
	})

	t.Run("should create two certificates in a chain", func(t *testing.T) {
		// when
		crtResponse, _ := createCertificateChain(t, client, clientKey)

		//then
		require.NotEmpty(t, crtResponse.Crt)

		// when
		certificates := testkit.DecodeAndParseCert(t, crtResponse)

		// then
		require.Equal(t, 2, len(certificates))
	})

	t.Run("client cert should be signed by server cert", func(t *testing.T) {
		//when
		crtResponse, _ := createCertificateChain(t, client, clientKey)

		//then
		require.NotEmpty(t, crtResponse.Crt)

		// when
		certificates := testkit.DecodeAndParseCert(t, crtResponse)

		//then
		testkit.CheckIfCertIsSigned(t, certificates)
	})

	t.Run("should validate CSR subject", func(t *testing.T) {
		// when
		tokenResponse := client.CreateToken(t)

		// then
		require.NotEmpty(t, tokenResponse.Token)
		require.Contains(t, tokenResponse.URL, "token="+tokenResponse.Token)

		// when
		infoResponse, errorResponse := client.GetInfo(t, tokenResponse.URL)

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
		require.Equal(t, StatusBadRequest, err.StatusCode)
		require.Equal(t, StatusBadRequest, err.ErrorResponse.Code)
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
		_, err := client.GetInfo(t, wrongUrl)

		// then
		require.NotNil(t, err)
		require.Equal(t, StatusForbidden, err.StatusCode)
		require.Equal(t, StatusForbidden, err.ErrorResponse.Code)
		require.Equal(t, "Invalid token.", err.ErrorResponse.Error)
	})

	t.Run("should return error for wrong token on client-certs", func(t *testing.T) {
		// when
		tokenResponse := client.CreateToken(t)

		// then
		require.NotEmpty(t, tokenResponse.Token)
		require.Contains(t, tokenResponse.URL, "token="+tokenResponse.Token)

		// when
		infoResponse, errorResponse := client.GetInfo(t, tokenResponse.URL)

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
		require.Equal(t, StatusForbidden, err.StatusCode)
		require.Equal(t, StatusForbidden, err.ErrorResponse.Code)
		require.Equal(t, "Invalid token.", err.ErrorResponse.Error)
	})

	t.Run("should return error on wrong CSR on client-certs", func(t *testing.T) {
		// when
		tokenResponse := client.CreateToken(t)

		// then
		require.NotEmpty(t, tokenResponse.Token)
		require.Contains(t, tokenResponse.URL, "token="+tokenResponse.Token)

		// when
		infoResponse, errorResponse := client.GetInfo(t, tokenResponse.URL)

		// then
		require.Nil(t, errorResponse)
		require.NotEmpty(t, infoResponse.CertUrl)
		require.Equal(t, "rsa2048", infoResponse.Certificate.KeyAlgorithm)

		// when
		_, err := client.CreateCertChain(t, "wrong-csr", infoResponse.CertUrl)

		// then
		require.NotNil(t, err)
		require.Equal(t, StatusBadRequest, err.StatusCode)
		require.Equal(t, StatusBadRequest, err.ErrorResponse.Code)
		require.Equal(t, "There was an error while parsing the base64 content. An incorrect value was provided.", err.ErrorResponse.Error)
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
		require.Equal(t, StatusOK, response.StatusCode)

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
		hc.CheckRedirect = func(req *Request, via []*Request) error {
			require.Equal(t, apiSpecPath, req.URL.Path)
			return ErrUseLastResponse
		}

		// when
		response, err := hc.Get(config.ExternalAPIUrl + "/v1")

		// then
		require.NoError(t, err)
		require.Equal(t, StatusMovedPermanently, response.StatusCode)
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

	forbiddenApp, err := k8sResourcesClient.CreateDummyApplication("app-connector-test-2", "", false)
	require.NoError(t, err)
	defer func() {
		k8sResourcesClient.DeleteApplication(forbiddenApp.Name, &v1.DeleteOptions{})
	}()

	client := testkit.NewConnectorClient(testApp.Name, config.InternalAPIUrl, config.ExternalAPIUrl, config.SkipSslVerify)

	clientKey := testkit.CreateKey(t)
	tlsClient := createTLSClientWithCert(t, client, clientKey, config.SkipSslVerify)

	t.Run("should access application", func(t *testing.T) {
		// when
		response, err := repeatUntilIngressIsCreated(tlsClient, gatewayUrlFormat, testApp.Name)

		// then
		require.NoError(t, err)
		require.Equal(t, StatusOK, response.StatusCode)
	})

	t.Run("should receive 403 when accessing RE with invalid CN", func(t *testing.T) {
		// when
		response, err := repeatUntilIngressIsCreated(tlsClient, gatewayUrlFormat, forbiddenApp.Name)

		// then
		require.NoError(t, err)
		require.Equal(t, StatusForbidden, response.StatusCode)
	})

}

func repeatUntilIngressIsCreated(tlsClient *Client, gatewayUrlFormat string, appName string) (*Response, error) {
	var response *Response
	var err error
	for i := 0; (shouldRetry(response, err)) && i < retryCount; i++ {
		response, err = tlsClient.Get(fmt.Sprintf(gatewayUrlFormat, appName))
		time.Sleep(retryWaitTimeSeconds)
	}
	return response, err
}

func shouldRetry(response *Response, err error) bool {
	return response == nil || StatusNotFound == response.StatusCode || err != nil
}

func createCertificateChain(t *testing.T, connectorClient testkit.ConnectorClient, key *rsa.PrivateKey) (*testkit.CrtResponse, *testkit.InfoResponse) {
	// when
	tokenResponse := connectorClient.CreateToken(t)

	// then
	require.NotEmpty(t, tokenResponse.Token)
	require.Contains(t, tokenResponse.URL, "token="+tokenResponse.Token)

	// when
	infoResponse, errorResponse := connectorClient.GetInfo(t, tokenResponse.URL)

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

func createTLSClientWithCert(t *testing.T, client testkit.ConnectorClient, key *rsa.PrivateKey, skipVerify bool) *Client {
	crtResponse, _ := createCertificateChain(t, client, key)
	require.NotEmpty(t, crtResponse.Crt)
	clientCertBytes, _ := testkit.CrtResponseToPemBytes(t, crtResponse)

	tlsCert := tls.Certificate{
		Certificate: [][]byte{clientCertBytes},
		PrivateKey:  key,
	}

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{tlsCert},
		ClientAuth:         tls.RequireAndVerifyClientCert,
		InsecureSkipVerify: skipVerify,
	}

	transport := &Transport{
		TLSClientConfig: tlsConfig,
	}

	return &Client{
		Transport: transport,
	}
}

func replaceToken(originalUrl string, newToken string) string {
	parsedUrl, _ := url.Parse(originalUrl)
	queryParams, _ := url.ParseQuery(parsedUrl.RawQuery)

	queryParams.Set("token", newToken)
	parsedUrl.RawQuery = queryParams.Encode()

	return parsedUrl.String()
}
