package tests

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/kyma-project/kyma/components/application-gateway/pkg/authorization"

	"github.com/kyma-project/kyma/tests/application-gateway-tests/test/gateway/testkit/proxyconfig"

	"github.com/kyma-project/kyma/tests/application-gateway-tests/test/gateway/testkit/util"

	"github.com/kyma-project/kyma/tests/application-gateway-tests/test/gateway"
)

func TestProxyService(t *testing.T) {

	testSuit := gateway.NewTestSuite(t)
	defer testSuit.Cleanup(t)
	testSuit.Setup(t)

	t.Run("no-auth api test", func(t *testing.T) {
		apiName := "no-auth-gateway-test"

		proxyConfig := proxyconfig.NewConfigBuilder(testSuit.GetMockServiceURL()).ToConfig()
		secretName := testSuit.CreateSecret(t, apiName, proxyConfig)
		defer func() {
			t.Logf("Cleaning up secret %s", secretName)
			testSuit.DeleteSecret(t, secretName)
		}()

		t.Log("Calling Gateway...")
		resp := testSuit.CallAPIThroughGateway(t, secretName, apiName, "status/ok")
		util.RequireStatus(t, http.StatusOK, resp)
		t.Log("Successfully accessed application")
	})

	t.Run("basic auth api test", func(t *testing.T) {
		apiName := "basic-auth-gateway-test"

		userName := "myUser"
		password := "mySecret"

		proxyConfig := proxyconfig.NewConfigBuilder(testSuit.GetMockServiceURL()).
			WithBasicAuth(userName, password).
			ToConfig()
		secretName := testSuit.CreateSecret(t, apiName, proxyConfig)
		defer func() {
			t.Logf("Cleaning up secret %s", secretName)
			testSuit.DeleteSecret(t, secretName)
		}()

		t.Log("Calling Gateway...")
		resp := testSuit.CallAPIThroughGateway(t, secretName, apiName, fmt.Sprintf("auth/basic/%s/%s", userName, password))
		util.RequireStatus(t, http.StatusOK, resp)
		t.Log("Successfully accessed application")
	})

	t.Run("oauth api test", func(t *testing.T) {
		apiName := "oauth-gateway-test"

		clientId := "myUser"
		clientSecret := "mySecret"
		mockServiceURL := testSuit.GetMockServiceURL()
		oauthUrl := fmt.Sprintf("/%s/auth/oauth/token/%s/%s", mockServiceURL, clientId, clientSecret)

		proxyConfig := proxyconfig.NewConfigBuilder(testSuit.GetMockServiceURL()).
			WithOAuth(clientId, clientSecret, oauthUrl, authorization.RequestParameters{}).
			ToConfig()
		secretName := testSuit.CreateSecret(t, apiName, proxyConfig)
		defer func() {
			t.Logf("Cleaning up secret %s", secretName)
			testSuit.DeleteSecret(t, secretName)
		}()

		t.Log("Calling Gateway...")
		resp := testSuit.CallAPIThroughGateway(t, secretName, apiName, "target/auth/oauth")
		util.RequireStatus(t, http.StatusOK, resp)
		t.Log("Successfully accessed application")
	})

	t.Run("additional header test", func(t *testing.T) {
		apiName := "additional-header-gateway-test"

		headerName := "Custom"
		headerValue := "CustomValue"

		headers := map[string][]string{
			headerName: []string{headerValue},
		}

		proxyConfig := proxyconfig.NewConfigBuilder(testSuit.GetMockServiceURL()).
			WithRequestParameters(authorization.RequestParameters{Headers: &headers}).
			ToConfig()
		secretName := testSuit.CreateSecret(t, apiName, proxyConfig)
		defer func() {
			t.Logf("Cleaning up secret %s", secretName)
			testSuit.DeleteSecret(t, secretName)
		}()

		t.Log("Calling Gateway...")
		resp := testSuit.CallAPIThroughGateway(t, secretName, apiName, fmt.Sprintf("headers/%s/%s", headerName, headerValue))
		util.RequireStatus(t, http.StatusOK, resp)
		t.Log("Successfully accessed application")
	})

	t.Run("additional query parameters  test", func(t *testing.T) {
		apiName := "additional-query-gateway-test"

		paramName := "customParam"
		paramValue := "customValue"

		queryParams := map[string][]string{
			paramName: []string{paramValue},
		}

		proxyConfig := proxyconfig.NewConfigBuilder(testSuit.GetMockServiceURL()).
			WithRequestParameters(authorization.RequestParameters{QueryParameters: &queryParams}).
			ToConfig()
		secretName := testSuit.CreateSecret(t, apiName, proxyConfig)
		defer func() {
			t.Logf("Cleaning up secret %s", secretName)
			testSuit.DeleteSecret(t, secretName)
		}()

		t.Log("Calling Gateway...")
		resp := testSuit.CallAPIThroughGateway(t, secretName, apiName, fmt.Sprintf("queryparams/%s/%s", paramName, paramValue))
		util.RequireStatus(t, http.StatusOK, resp)
		t.Log("Successfully accessed application")
	})

	t.Run("retry with new CSRF token for basic auth test", func(t *testing.T) {
		apiName := "csrf-retry-gateway-test"

		username := "username"
		password := "password"
		csrfURL := fmt.Sprintf("%s%s", testSuit.GetMockServiceURL(), "csrftoken")

		proxyConfig := proxyconfig.NewConfigBuilder(testSuit.GetMockServiceURL()).
			WithBasicAuth(username, password).
			WithCSRF(csrfURL).
			ToConfig()
		secretName := testSuit.CreateSecret(t, apiName, proxyConfig)
		defer func() {
			t.Logf("Cleaning up secret %s", secretName)
			testSuit.DeleteSecret(t, secretName)
		}()

		t.Log("Calling Gateway with correct token...")
		resp := testSuit.CallAPIThroughGateway(t, secretName, apiName, "target")
		util.RequireStatus(t, http.StatusOK, resp)

		t.Log("Calling Gateway second time with invalid token, with expected retry...")
		resp = testSuit.CallAPIThroughGateway(t, secretName, apiName, "target")
		util.RequireStatus(t, http.StatusOK, resp)
		t.Log("Successfully accessed application")
	})

	t.Run("oauth additional query parameters api test", func(t *testing.T) {
		apiName := "oauth-additional-query-gateway-test"

		clientId := "myUser"
		clientSecret := "mySecret"

		paramName := "customParam"
		paramValue := "customValue"

		oauthUrl := fmt.Sprintf("%s/auth/oauth/token/%s/%s/queryparams/%s/%s", testSuit.GetMockServiceURL(), clientId, clientSecret, paramName, paramValue)

		queryParams := map[string][]string{
			paramName: []string{paramValue},
		}

		proxyConfig := proxyconfig.NewConfigBuilder(testSuit.GetMockServiceURL()).
			WithOAuth(clientId, clientSecret, oauthUrl, authorization.RequestParameters{QueryParameters: &queryParams}).
			ToConfig()
		secretName := testSuit.CreateSecret(t, apiName, proxyConfig)
		defer func() {
			t.Logf("Cleaning up secret %s", secretName)
			testSuit.DeleteSecret(t, secretName)
		}()

		t.Log("Calling Gateway with correct token...")
		resp := testSuit.CallAPIThroughGateway(t, secretName, apiName, "target")
		util.RequireStatus(t, http.StatusOK, resp)

		t.Log("Calling Gateway second time with invalid token, with expected retry...")
		resp = testSuit.CallAPIThroughGateway(t, secretName, apiName, "target/auth/oauth")
		util.RequireStatus(t, http.StatusOK, resp)
		t.Log("Successfully accessed application")
	})

	t.Run("oauth additional headers api test", func(t *testing.T) {
		apiName := "oauth-additional-query-gateway-test"

		clientId := "myUser"
		clientSecret := "mySecret"

		headerName := "Custom"
		headerValue := "CustomValue"

		oauthUrl := fmt.Sprintf("%s/auth/oauth/token/%s/%s/headers/%s/%s", testSuit.GetMockServiceURL(), clientId, clientSecret, headerName, headerValue)

		headers := map[string][]string{
			headerName: []string{headerValue},
		}

		proxyConfig := proxyconfig.NewConfigBuilder(testSuit.GetMockServiceURL()).
			WithOAuth(clientId, clientSecret, oauthUrl, authorization.RequestParameters{Headers: &headers}).
			ToConfig()
		secretName := testSuit.CreateSecret(t, apiName, proxyConfig)
		defer func() {
			t.Logf("Cleaning up secret %s", secretName)
			testSuit.DeleteSecret(t, secretName)
		}()

		t.Log("Calling Gateway with correct token...")
		resp := testSuit.CallAPIThroughGateway(t, secretName, apiName, "target")
		util.RequireStatus(t, http.StatusOK, resp)

		t.Log("Calling Gateway second time with invalid token, with expected retry...")
		resp = testSuit.CallAPIThroughGateway(t, secretName, apiName, "target/auth/oauth")
		util.RequireStatus(t, http.StatusOK, resp)
		t.Log("Successfully accessed application")
	})

}
