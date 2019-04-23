package apitests

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/tests/connector-service-tests/test/testkit"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCertificateValidation(t *testing.T) {

	config, err := testkit.ReadConfig()
	require.NoError(t, err)

	gatewayUrlFormat := config.GatewayUrl + "/%s/v1/metadata/services"

	k8sResourcesClient, err := testkit.NewK8sResourcesClient()
	require.NoError(t, err)
	dummyApplicationSpec := getApplicationSpec(config.Central)
	testApp, err := k8sResourcesClient.CreateDummyApplication("app-connector-test-1", dummyApplicationSpec)
	require.NoError(t, err)
	defer func() {
		k8sResourcesClient.DeleteApplication(testApp.Name, &v1.DeleteOptions{})
	}()

	t.Run("should access application", func(t *testing.T) {
		// given
		tokenRequest := createApplicationTokenRequest(t, config, testApp.Name, config.Central)
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
		tokenRequest := createApplicationTokenRequest(t, config, "another-application", config.Central)
		connectorClient := testkit.NewConnectorClient(tokenRequest, config.SkipSslVerify)
		tlsClient := createTLSClientWithCert(t, connectorClient, config.SkipSslVerify)

		// when
		response, err := repeatUntilIngressIsCreated(tlsClient, gatewayUrlFormat, testApp.Name)

		// then
		require.NoError(t, err)
		require.Equal(t, http.StatusForbidden, response.StatusCode)
	})

}

func getApplicationSpec(central bool) v1alpha1.ApplicationSpec {
	spec := v1alpha1.ApplicationSpec{
		Services:         []v1alpha1.Service{},
		AccessLabel:      "",
		SkipInstallation: false,
	}

	if central {
		spec.Tenant = testkit.Tenant
		spec.Group = testkit.Group
	}

	return spec
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

	crtResponse, _ := createCertificateChain(t, client, key, createHostsHeaders("", ""))
	require.NotEmpty(t, crtResponse.CRTChain)
	clientCertBytes := testkit.EncodedCertToPemBytes(t, crtResponse.ClientCRT)

	return testkit.NewTLSClientWithCert(skipVerify, key, clientCertBytes)
}
