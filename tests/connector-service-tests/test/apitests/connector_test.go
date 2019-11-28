package apitests

import (
	"crypto/rsa"
	"net/http"
	"net/url"
	"testing"

	"github.com/kyma-project/kyma/tests/connector-service-tests/test/testkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnector(t *testing.T) {

	config, err := testkit.ReadConfig()
	require.NoError(t, err)

	t.Run("Connector Service flow for Application", func(t *testing.T) {
		appName := "test-app"
		appTokenRequest := createApplicationTokenRequest(t, config, appName, config.Central)
		certificateGenerationSuite(t, appTokenRequest, config.SkipSslVerify)
		appMgmInfoEndpointSuite(t, appTokenRequest, config.SkipSslVerify, config.Central, config.GatewayUrl, appName)
		appCsrInfoEndpointSuite(t, appTokenRequest, config, appName)
		certificateRotationSuite(t, appTokenRequest, config.SkipSslVerify)

		appTokenRequestWithTenantAndGroup := createApplicationTokenRequest(t, config, appName, true)
		subjectGenerationSuite(t, appTokenRequestWithTenantAndGroup, config, appName)

		internalRevocationUrl := createApplicationRevocationUrl(config)
		certificateRevocationSuite(t, appTokenRequest, config.SkipSslVerify, internalRevocationUrl)
	})

	if config.Central {
		t.Run("Connector Service flow for Runtime", func(t *testing.T) {
			runtimeTokenRequest := createRuntimeTokenRequest(t, config)
			certificateGenerationSuite(t, runtimeTokenRequest, config.SkipSslVerify)
			runtimeCsrInfoEndpointForCentralSuite(t, runtimeTokenRequest, config.SkipSslVerify)
			runtimeMgmInfoEndpointForCentralSuite(t, runtimeTokenRequest, config.SkipSslVerify)

			certificateRotationSuite(t, runtimeTokenRequest, config.SkipSslVerify)

			internalRevocationUrl := createRuntimeRevocationUrl(config)
			certificateRevocationSuite(t, runtimeTokenRequest, config.SkipSslVerify, internalRevocationUrl)
		})
	}
}

func createApplicationTokenRequest(t *testing.T, config testkit.TestConfig, appName string, withGroupAndTenant bool) *http.Request {
	tokenURL := config.InternalAPIUrl + "/v1/applications/tokens"

	request := createTokenRequest(t, tokenURL, withGroupAndTenant)
	request.Header.Set(testkit.ApplicationHeader, appName)

	return request
}

func createRuntimeTokenRequest(t *testing.T, config testkit.TestConfig) *http.Request {
	tokenURL := config.InternalAPIUrl + "/v1/runtimes/tokens"

	request := createTokenRequest(t, tokenURL, true)

	return request
}

func createTokenRequest(t *testing.T, tokenURL string, withGroupAndTenant bool) *http.Request {
	request, err := http.NewRequest(http.MethodPost, tokenURL, nil)
	require.NoError(t, err)

	request.Close = true

	if withGroupAndTenant {
		request.Header.Set(testkit.GroupHeader, testkit.Group)
		request.Header.Set(testkit.TenantHeader, testkit.Tenant)
	}

	return request
}

func createApplicationRevocationUrl(config testkit.TestConfig) string {
	return config.InternalAPIUrl + "/v1/applications/certificates/revocations"
}

func createRuntimeRevocationUrl(config testkit.TestConfig) string {
	return config.InternalAPIUrl + "/v1/runtimes/certificates/revocations"
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
		infoResponse, errorResponse := client.GetInfo(t, tokenResponse.URL)

		// then
		require.Nil(t, errorResponse)
		require.NotEmpty(t, infoResponse.CertUrl)
		require.Equal(t, "rsa2048", infoResponse.Certificate.KeyAlgorithm)

		// given
		infoResponse.Certificate.Subject = "subject=OU=Test,O=Test,L=Wrong,ST=Wrong,C=PL,CN=Wrong"
		csr := testkit.CreateCsr(t, infoResponse.Certificate.Subject, clientKey)
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
		infoResponse, errorResponse := client.GetInfo(t, tokenResponse.URL)

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

func appCsrInfoEndpointSuite(t *testing.T, tokenRequest *http.Request, config testkit.TestConfig, appName string) {

	t.Run("should use default values to build CSR info response", func(t *testing.T) {
		// given
		client := testkit.NewConnectorClient(tokenRequest, config.SkipSslVerify)
		expectedMetadataURL := config.GatewayUrl
		expectedEventsURL := config.GatewayUrl

		if config.GatewayUrl != "" {
			expectedMetadataURL += "/" + appName + "/v1/metadata/services"
			expectedEventsURL += "/" + appName + "/v1/events"
		}

		// when
		tokenResponse := client.CreateToken(t)

		// then
		require.NotEmpty(t, tokenResponse.Token)
		require.Contains(t, tokenResponse.URL, "token="+tokenResponse.Token)

		// when
		infoResponse, errorResponse := client.GetInfo(t, tokenResponse.URL)

		// then
		require.Nil(t, errorResponse)
		assert.Equal(t, expectedEventsURL, infoResponse.Api.RuntimeURLs.EventsUrl)
		assert.Equal(t, expectedMetadataURL, infoResponse.Api.RuntimeURLs.MetadataUrl)
	})
}

func subjectGenerationSuite(t *testing.T, tokenRequest *http.Request, config testkit.TestConfig, appName string) {

	client := testkit.NewConnectorClient(tokenRequest, config.SkipSslVerify)

	// when
	tokenResponse := client.CreateToken(t)

	// then
	require.NotEmpty(t, tokenResponse.Token)
	require.Contains(t, tokenResponse.URL, "token="+tokenResponse.Token)

	// when
	infoResponse, errorResponse := client.GetInfo(t, tokenResponse.URL)

	// then
	require.Nil(t, errorResponse)
	subject := testkit.ParseSubject(infoResponse.Certificate.Subject)
	require.NotEmpty(t, subject.Organization[0])
	require.NotEmpty(t, subject.OrganizationalUnit[0])

	if !config.Central {
		t.Run("should ignore tenant and group if not central", func(t *testing.T) {
			assert.NotEqual(t, testkit.Tenant, subject.Organization[0])
			assert.NotEqual(t, testkit.Group, subject.OrganizationalUnit[0])
		})
	} else {
		t.Run("should set Organization as Tenant and OrganizationalUnit as Group", func(t *testing.T) {
			assert.Equal(t, testkit.Tenant, subject.Organization[0])
			assert.Equal(t, testkit.Group, subject.OrganizationalUnit[0])
		})
	}

}

func runtimeCsrInfoEndpointForCentralSuite(t *testing.T, tokenRequest *http.Request, skipVerify bool) {

	client := testkit.NewConnectorClient(tokenRequest, skipVerify)

	t.Run("should provide not empty CSR info response", func(t *testing.T) {
		// when
		tokenResponse := client.CreateToken(t)

		// then
		require.NotEmpty(t, tokenResponse.Token)
		require.Contains(t, tokenResponse.URL, "token="+tokenResponse.Token)

		// when
		infoResponse, errorResponse := client.GetInfo(t, tokenResponse.URL)

		// then
		require.Nil(t, errorResponse)
		assert.NotEmpty(t, infoResponse.CertUrl)
		assert.NotEmpty(t, infoResponse.Api)
		assert.NotEmpty(t, infoResponse.Certificate)
		assert.Nil(t, infoResponse.Api.RuntimeURLs)
	})
}

func appMgmInfoEndpointSuite(t *testing.T, tokenRequest *http.Request, skipVerify bool, central bool, defaultGatewayUrl string, appName string) {

	client := testkit.NewConnectorClient(tokenRequest, skipVerify)

	clientKey := testkit.CreateKey(t)

	t.Run("should use default values to build management info", func(t *testing.T) {
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
		require.NotEmpty(t, infoResponse.Api.ManagementInfoURL)

		certificates := testkit.DecodeAndParseCerts(t, crtResponse)
		client := testkit.NewSecuredConnectorClient(skipVerify, clientKey, certificates.ClientCRT.Raw)

		// when
		mgmInfoResponse, errorResponse := client.GetMgmInfo(t, infoResponse.Api.ManagementInfoURL)
		require.Nil(t, errorResponse)

		// then
		assert.Equal(t, expectedMetadataURL, mgmInfoResponse.URLs.MetadataUrl)
		assert.Equal(t, expectedEventsURL, mgmInfoResponse.URLs.EventsUrl)
		assert.Equal(t, appName, mgmInfoResponse.ClientIdentity.Application)
		assert.NotEmpty(t, mgmInfoResponse.Certificate.Subject)
		assert.Equal(t, testkit.Extensions, mgmInfoResponse.Certificate.Extensions)
		assert.Equal(t, testkit.KeyAlgorithm, mgmInfoResponse.Certificate.KeyAlgorithm)

		if central {
			assert.Equal(t, testkit.Group, mgmInfoResponse.ClientIdentity.Group)
			assert.Equal(t, testkit.Tenant, mgmInfoResponse.ClientIdentity.Tenant)
		} else {
			assert.Empty(t, mgmInfoResponse.ClientIdentity.Group)
			assert.Empty(t, mgmInfoResponse.ClientIdentity.Tenant)
		}
	})
}

func runtimeMgmInfoEndpointForCentralSuite(t *testing.T, tokenRequest *http.Request, skipVerify bool) {

	client := testkit.NewConnectorClient(tokenRequest, skipVerify)

	clientKey := testkit.CreateKey(t)

	t.Run("should provide not empty management info response", func(t *testing.T) {
		// when
		crtResponse, infoResponse := createCertificateChain(t, client, clientKey)

		// then
		require.NotEmpty(t, crtResponse.CRTChain)
		require.NotEmpty(t, infoResponse.Api.ManagementInfoURL)

		certificates := testkit.DecodeAndParseCerts(t, crtResponse)
		client := testkit.NewSecuredConnectorClient(skipVerify, clientKey, certificates.ClientCRT.Raw)

		// when
		mgmInfoResponse, errorResponse := client.GetMgmInfo(t, infoResponse.Api.ManagementInfoURL)
		require.Nil(t, errorResponse)

		// then
		assert.Nil(t, mgmInfoResponse.URLs.RuntimeURLs)
		assert.Equal(t, testkit.Group, mgmInfoResponse.ClientIdentity.Group)
		assert.Equal(t, testkit.Tenant, mgmInfoResponse.ClientIdentity.Tenant)
		assert.NotEmpty(t, mgmInfoResponse.Certificate.Subject)
		assert.Equal(t, testkit.Extensions, mgmInfoResponse.Certificate.Extensions)
		assert.Equal(t, testkit.KeyAlgorithm, mgmInfoResponse.Certificate.KeyAlgorithm)
	})

}

func certificateRotationSuite(t *testing.T, tokenRequest *http.Request, skipVerify bool) {
	client := testkit.NewConnectorClient(tokenRequest, skipVerify)

	clientKey := testkit.CreateKey(t)

	t.Run("should renew client certificate", func(t *testing.T) {
		// when
		crtResponse, infoResponse := createCertificateChain(t, client, clientKey)
		require.NotEmpty(t, crtResponse.CRTChain)
		require.NotEmpty(t, infoResponse.Api.ManagementInfoURL)
		require.NotEmpty(t, infoResponse.Certificate)

		certificates := testkit.DecodeAndParseCerts(t, crtResponse)
		client := testkit.NewSecuredConnectorClient(skipVerify, clientKey, certificates.ClientCRT.Raw)

		mgmInfoResponse, errorResponse := client.GetMgmInfo(t, infoResponse.Api.ManagementInfoURL)
		require.Nil(t, errorResponse)
		require.NotEmpty(t, mgmInfoResponse.URLs.RenewCertUrl)
		require.NotEmpty(t, mgmInfoResponse.Certificate)
		require.Equal(t, infoResponse.Certificate, mgmInfoResponse.Certificate)

		csr := testkit.CreateCsr(t, mgmInfoResponse.Certificate.Subject, clientKey)
		csrBase64 := testkit.EncodeBase64(csr)

		certificateResponse, errorResponse := client.RenewCertificate(t, mgmInfoResponse.URLs.RenewCertUrl, csrBase64)

		// then
		require.Nil(t, errorResponse)

		certificates = testkit.DecodeAndParseCerts(t, certificateResponse)
		clientWithRenewedCert := testkit.NewSecuredConnectorClient(skipVerify, clientKey, certificates.ClientCRT.Raw)

		mgmInfoResponse, errorResponse = clientWithRenewedCert.GetMgmInfo(t, infoResponse.Api.ManagementInfoURL)
		require.Nil(t, errorResponse)
	})
}

func certificateRevocationSuite(t *testing.T, tokenRequest *http.Request, skipVerify bool, internalRevocationUrl string) {
	client := testkit.NewConnectorClient(tokenRequest, skipVerify)

	clientKey := testkit.CreateKey(t)

	t.Run("should revoke client certificate with external API", func(t *testing.T) {
		// when
		crtResponse, infoResponse := createCertificateChain(t, client, clientKey)

		// then
		require.NotEmpty(t, crtResponse.CRTChain)
		require.NotEmpty(t, infoResponse.Api.ManagementInfoURL)

		// when
		certificates := testkit.DecodeAndParseCerts(t, crtResponse)
		client := testkit.NewSecuredConnectorClient(skipVerify, clientKey, certificates.ClientCRT.Raw)

		mgmInfoResponse, errorResponse := client.GetMgmInfo(t, infoResponse.Api.ManagementInfoURL)

		// then
		require.Nil(t, errorResponse)
		require.NotEmpty(t, mgmInfoResponse.URLs.RevokeCertURL)

		// when
		errorResponse = client.RevokeCertificate(t, mgmInfoResponse.URLs.RevokeCertURL)

		// then
		require.Nil(t, errorResponse)

		// when
		csr := testkit.CreateCsr(t, infoResponse.Certificate.Subject, clientKey)
		csrBase64 := testkit.EncodeBase64(csr)

		_, errorResponse = client.RenewCertificate(t, mgmInfoResponse.URLs.RenewCertUrl, csrBase64)

		// then
		require.NotNil(t, errorResponse)
		require.Equal(t, http.StatusForbidden, errorResponse.StatusCode)
	})

	t.Run("should revoke client certificate with internal API", func(t *testing.T) {
		// when
		crtResponse, infoResponse := createCertificateChain(t, client, clientKey)

		// then
		require.NotEmpty(t, crtResponse.CRTChain)
		require.NotEmpty(t, infoResponse.Api.ManagementInfoURL)

		// when
		certificates := testkit.DecodeAndParseCerts(t, crtResponse)
		securedClient := testkit.NewSecuredConnectorClient(skipVerify, clientKey, certificates.ClientCRT.Raw)

		mgmInfoResponse, errorResponse := securedClient.GetMgmInfo(t, infoResponse.Api.ManagementInfoURL)

		// then
		require.Nil(t, errorResponse)
		require.NotEmpty(t, mgmInfoResponse.URLs.RevokeCertURL)

		// when
		sha256Fingerprint := testkit.CertificateSHA256Fingerprint(t, certificates.ClientCRT)

		errorResponse = client.RevokeCertificate(t, internalRevocationUrl, sha256Fingerprint)

		// then
		require.Nil(t, errorResponse)

		// when
		csr := testkit.CreateCsr(t, infoResponse.Certificate.Subject, clientKey)
		csrBase64 := testkit.EncodeBase64(csr)

		_, errorResponse = securedClient.RenewCertificate(t, mgmInfoResponse.URLs.RenewCertUrl, csrBase64)

		// then
		require.NotNil(t, errorResponse)
		require.Equal(t, http.StatusForbidden, errorResponse.StatusCode)
	})

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
	csr := testkit.CreateCsr(t, infoResponse.Certificate.Subject, key)
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
