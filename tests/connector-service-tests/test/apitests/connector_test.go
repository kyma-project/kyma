package apitests

import (
	"crypto/rsa"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/kyma-project/kyma/tests/connector-service-tests/test/testkit"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"gopkg.in/yaml.v2"
)

func TestConnector(t *testing.T) {

	config, err := testkit.ReadConfig()
	require.NoError(t, err)

	remoteEnvName := "ec-default"

	client := testkit.NewConnectorClient(remoteEnvName, config.InternalAPIUrl, config.ExternalAPIUrl)

	clientKey := testkit.CreateKey(t)

	t.Run("should create client certificate", func(t *testing.T) {
		// when
		crtResponse, infoResponse := createClientCertificate(t, client, clientKey)

		//then
		require.NotEmpty(t, crtResponse.Crt)

		// when
		certificate := testkit.DecodeAndParseCert(t, crtResponse)

		// then
		testkit.CheckIfSubjectEquals(t, infoResponse.Certificate.Subject, certificate)
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
		_, err := client.CreateClientCert(t, csrBase64, infoResponse.CertUrl)

		// then
		require.NotNil(t, err)
		require.Equal(t, http.StatusForbidden, err.StatusCode)
		require.Equal(t, http.StatusForbidden, err.ErrorResponse.Code)
		require.Equal(t, "CSR: Invalid CName provided.", err.ErrorResponse.Error)
	})

	t.Run("should accept only one token per remote environment", func(t *testing.T) {
		// when
		tokenResponse := client.CreateToken(t)

		// then
		require.NotEmpty(t, tokenResponse.Token)
		require.Contains(t, tokenResponse.URL, "token="+tokenResponse.Token)

		// when
		tokenResponse2 := client.CreateToken(t)

		// then
		require.NotEmpty(t, tokenResponse2.Token)
		require.Contains(t, tokenResponse2.URL, "token="+tokenResponse2.Token)

		// when
		infoResponse, errorResponse := client.GetInfo(t, tokenResponse.URL)

		// then
		require.Nil(t, infoResponse)
		require.Equal(t, http.StatusForbidden, errorResponse.StatusCode)

		// when
		infoResponse2, errorResponse2 := client.GetInfo(t, tokenResponse2.URL)

		// then
		require.Nil(t, errorResponse2)
		require.NotEmpty(t, infoResponse2.CertUrl)
		require.Equal(t, "rsa2048", infoResponse2.Certificate.KeyAlgorithm)
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
		_, err := client.CreateClientCert(t, "csr", wrongUrl)

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
		infoResponse, errorResponse := client.GetInfo(t, tokenResponse.URL)

		// then
		require.Nil(t, errorResponse)
		require.NotEmpty(t, infoResponse.CertUrl)
		require.Equal(t, "rsa2048", infoResponse.Certificate.KeyAlgorithm)

		// when
		_, err := client.CreateClientCert(t, "wrong-csr", infoResponse.CertUrl)

		// then
		// TODO after fixing the service, set status code to 400.
		require.NotNil(t, err)
		require.Equal(t, http.StatusInternalServerError, err.StatusCode)
		require.Equal(t, http.StatusInternalServerError, err.ErrorResponse.Code)
		require.Equal(t, "Error while decoding base64 string: illegal base64 data at input byte 5", err.ErrorResponse.Error)
	})
}

func TestApiSpec(t *testing.T) {

	apiSpecPath := "/v1/api.yaml"

	config, err := testkit.ReadConfig()
	require.NoError(t, err)

	t.Run("should receive api spec", func(t *testing.T) {
		// when
		response, err := http.Get(config.ExternalAPIUrl + apiSpecPath)

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
		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				require.Equal(t, apiSpecPath, req.URL.Path)
				return http.ErrUseLastResponse
			},
		}

		// when
		response, err := client.Get(config.ExternalAPIUrl + "/v1")

		// then
		require.NoError(t, err)
		require.Equal(t, http.StatusMovedPermanently, response.StatusCode)
	})
}

func TestCertificateValidation(t *testing.T) {

	config, err := testkit.ReadConfig()
	require.NoError(t, err)

	gatewayUrlFormat := config.GatewayUrl + "/%s/v1/metadata/services"

	remoteEnvName := "ec-default"
	forbiddenRemoteEnvName := "hmc-default"

	client := testkit.NewConnectorClient(remoteEnvName, config.InternalAPIUrl, config.ExternalAPIUrl)

	clientKey := testkit.CreateKey(t)

	t.Run("should access remote environment", func(t *testing.T) {
		// given
		crtResponse, _ := createClientCertificate(t, client, clientKey)
		require.NotEmpty(t, crtResponse.Crt)
		pemBytes := testkit.CrtResponseToPemBytes(t, crtResponse)
		tlsClient := createTLSClient(pemBytes, clientKey)

		// when
		response, err := tlsClient.Get(fmt.Sprintf(gatewayUrlFormat, remoteEnvName))

		// then
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, response.StatusCode)
	})

	t.Run("should receive 403 when accessing RE with invalid CN", func(t *testing.T) {
		// given
		crtResponse, _ := createClientCertificate(t, client, clientKey)
		require.NotEmpty(t, crtResponse.Crt)
		pemBytes := testkit.CrtResponseToPemBytes(t, crtResponse)
		tlsClient := createTLSClient(pemBytes, clientKey)

		// when
		response, err := tlsClient.Get(fmt.Sprintf(gatewayUrlFormat, forbiddenRemoteEnvName))

		// then
		require.NoError(t, err)
		require.Equal(t, http.StatusForbidden, response.StatusCode)
	})
}

func createClientCertificate(t *testing.T, connectorClient testkit.ConnectorClient, key *rsa.PrivateKey) (*testkit.CrtResponse, *testkit.InfoResponse) {
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
	crtResponse, errorResponse := connectorClient.CreateClientCert(t, csrBase64, infoResponse.CertUrl)

	// then
	require.Nil(t, errorResponse)

	return crtResponse, infoResponse
}

func createTLSClient(certificate []byte, key *rsa.PrivateKey) *http.Client {
	tlsCert := tls.Certificate{
		Certificate: [][]byte{certificate},
		PrivateKey:  key,
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	return &http.Client{
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
