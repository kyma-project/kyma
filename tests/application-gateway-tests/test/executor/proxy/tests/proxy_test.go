package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/kyma-project/kyma/tests/application-gateway-tests/test/executor/testkit/util"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-project/kyma/tests/application-gateway-tests/test/executor/proxy"
	"github.com/kyma-project/kyma/tests/application-gateway-tests/test/executor/testkit/registry"
)

func TestProxyService(t *testing.T) {

	testSuit := proxy.NewTestSuite(t)
	defer testSuit.Cleanup(t)
	testSuit.Setup(t)

	client := registry.NewAppRegistryClient("http://application-registry-external-api:8081", testSuit.ApplicationName())

	t.Run("no-auth api test", func(t *testing.T) {
		apiID := client.CreateNotSecuredAPI(t, testSuit.GetMockServiceURL())
		t.Logf("Created service with apiID: %s", apiID)
		defer func() {
			t.Logf("Cleaning up service %s", apiID)
			client.CleanupService(t, apiID)
		}()

		t.Log("Labeling tests pod with denier label")
		testSuit.AddDenierLabel(t, apiID)

		t.Log("Calling Access Service")
		resp := testSuit.CallAccessService(t, apiID, "status/ok")
		util.RequireStatus(t, http.StatusOK, resp)

		t.Log("Successfully accessed application")
	})

	t.Run("basic auth api test", func(t *testing.T) {
		userName := "myUser"
		password := "mySecret"

		apiID := client.CreateBasicAuthSecuredAPI(t, testSuit.GetMockServiceURL(), userName, password)
		t.Logf("Created service with apiID: %s", apiID)
		defer func() {
			t.Logf("Cleaning up service %s", apiID)
			client.CleanupService(t, apiID)
		}()

		t.Log("Labeling tests pod with denier label")
		testSuit.AddDenierLabel(t, apiID)

		t.Log("Calling Access Service")
		resp := testSuit.CallAccessService(t, apiID, fmt.Sprintf("auth/basic/%s/%s", userName, password))
		util.RequireStatus(t, http.StatusOK, resp)

		t.Log("Successfully accessed application")
	})

	t.Run("oauth api test", func(t *testing.T) {
		clientId := "myUser"
		clientSecret := "mySecret"

		mockServiceURL := testSuit.GetMockServiceURL()
		oauthUrl := fmt.Sprintf("%s/auth/oauth/token/%s/%s", mockServiceURL, clientId, clientSecret)
		apiID := client.CreateOAuthSecuredAPI(t, mockServiceURL, oauthUrl, clientId, clientSecret)
		t.Logf("Created service with apiID: %s", apiID)
		defer func() {
			t.Logf("Cleaning up service %s", apiID)
			client.CleanupService(t, apiID)
		}()

		t.Log("Labeling tests pod with denier label")
		testSuit.AddDenierLabel(t, apiID)

		t.Log("Calling Access Service")
		resp := testSuit.CallAccessService(t, apiID, "/target/auth/oauth")
		util.RequireStatus(t, http.StatusOK, resp)

		t.Log("Successfully accessed application")
	})

	t.Run("additional header test", func(t *testing.T) {
		headerName := "Custom"
		headerValue := "CustomValue"

		headers := map[string][]string{
			headerName: []string{headerValue},
		}

		apiID := client.CreateNotSecuredAPICustomHeaders(t, testSuit.GetMockServiceURL(), headers)
		t.Logf("Created service with apiID: %s", apiID)
		defer func() {
			t.Logf("Cleaning up service %s", apiID)
			client.CleanupService(t, apiID)
		}()

		t.Log("Labeling tests pod with denier label")
		testSuit.AddDenierLabel(t, apiID)

		t.Log("Calling Access Service")
		resp := testSuit.CallAccessService(t, apiID, fmt.Sprintf("headers/%s/%s", headerName, headerValue))
		util.RequireStatus(t, http.StatusOK, resp)

		t.Log("Successfully accessed application")
	})

	t.Run("additional query parameters test", func(t *testing.T) {
		paramName := "customParam"
		paramValue := "customValue"

		queryParams := map[string][]string{
			paramName: []string{paramValue},
		}

		apiID := client.CreateNotSecuredAPICustomQueryParams(t, testSuit.GetMockServiceURL(), queryParams)
		t.Logf("Created service with apiID: %s", apiID)
		defer func() {
			t.Logf("Cleaning up service %s", apiID)
			client.CleanupService(t, apiID)
		}()

		t.Log("Labeling tests pod with denier label")
		testSuit.AddDenierLabel(t, apiID)

		t.Log("Calling Access Service")
		resp := testSuit.CallAccessService(t, apiID, fmt.Sprintf("queryparams/%s/%s", paramName, paramValue))
		util.RequireStatus(t, http.StatusOK, resp)

		t.Log("Successfully accessed application")
	})

	t.Run("denier test - should receive with 403 when not labeled", func(t *testing.T) {
		apiID := client.CreateNotSecuredAPI(t, testSuit.GetMockServiceURL())
		t.Logf("Created service with apiID: %s", apiID)
		defer func() {
			t.Logf("Cleaning up service %s", apiID)
			client.CleanupService(t, apiID)
		}()

		t.Log("Skip labeling tests pod with denier label")

		t.Log("Calling Access Service")
		resp := testSuit.CallAccessService(t, apiID, "status/ok")
		util.RequireStatus(t, http.StatusForbidden, resp)

		t.Log("Denier successfully restricted access")
	})

	t.Run("retry with new CSRF token for basic auth test", func(t *testing.T) {
		username := "username"
		password := "password"
		csrfURL := fmt.Sprintf("%s%s", testSuit.GetMockServiceURL(), "/csrftoken")

		apiID := client.CreateCSRFAndBasicSecuredAPI(t, testSuit.GetMockServiceURL(), username, password, csrfURL)
		t.Logf("Created service with apiID: %s", apiID)
		defer func() {
			t.Logf("Cleaning up service %s", apiID)
			client.CleanupService(t, apiID)
		}()

		t.Log("Labeling tests pod with denier label")
		testSuit.AddDenierLabel(t, apiID)

		t.Log("Calling Access Service with correct CSRF token")
		resp := testSuit.CallAccessService(t, apiID, "target")
		util.RequireStatus(t, http.StatusOK, resp)

		t.Log("Calling Access Service with incorrect CSRF token and retry with the correct one")
		resp = testSuit.CallAccessService(t, apiID, "target")
		util.RequireStatus(t, http.StatusOK, resp)

		t.Log("Successfully accessed application")
	})

	t.Run("oauth additional query parameters api test", func(t *testing.T) {
		clientId := "myUser"
		clientSecret := "mySecret"

		paramName := "customParam"
		paramValue := "customValue"

		queryParams := map[string][]string{
			paramName: []string{paramValue},
		}

		mockServiceURL := testSuit.GetMockServiceURL()
		oauthUrl := fmt.Sprintf("%s/auth/oauthqueryparams/token/%s/%s/%s/%s", mockServiceURL, clientId, clientSecret, paramName, paramValue)
		apiID := client.CreateOAuthWithCustomQueryParamsSecuredAPI(t, mockServiceURL, oauthUrl, clientId, clientSecret, queryParams)
		t.Logf("Created service with apiID: %s", apiID)
		defer func() {
			t.Logf("Cleaning up service %s", apiID)
			client.CleanupService(t, apiID)
		}()

		t.Log("Labeling tests pod with denier label")
		testSuit.AddDenierLabel(t, apiID)

		t.Log("Calling Access Service")
		resp := testSuit.CallAccessService(t, apiID, "/target/auth/oauth")
		util.RequireStatus(t, http.StatusOK, resp)

		t.Log("Successfully accessed application")
	})

	t.Run("oauth additional headers api test", func(t *testing.T) {
		clientId := "myUser"
		clientSecret := "mySecret"

		headerName := "Custom"
		headerValue := "CustomValue"

		headers := map[string][]string{
			headerName: []string{headerValue},
		}

		mockServiceURL := testSuit.GetMockServiceURL()
		oauthUrl := fmt.Sprintf("%s/auth/oauthheaders/token/%s/%s/%s/%s", mockServiceURL, clientId, clientSecret, headerName, headerValue)
		apiID := client.CreateOAuthWithCustomHeadersSecuredAPI(t, mockServiceURL, oauthUrl, clientId, clientSecret, headers)
		t.Logf("Created service with apiID: %s", apiID)
		defer func() {
			t.Logf("Cleaning up service %s", apiID)
			client.CleanupService(t, apiID)
		}()

		t.Log("Labeling tests pod with denier label")
		testSuit.AddDenierLabel(t, apiID)

		t.Log("Calling Access Service")
		resp := testSuit.CallAccessService(t, apiID, "/target/auth/oauth")
		util.RequireStatus(t, http.StatusOK, resp)

		t.Log("Successfully accessed application")
	})

	//Protected spec fetching tests
	t.Run("basic auth spec url test", func(t *testing.T) {
		userName := "myUser"
		password := "mySecret"

		mockServiceURL := testSuit.GetMockServiceURL()
		specUrl := fmt.Sprintf("%s/spec/auth/basic/%s/%s", mockServiceURL, userName, password)

		apiID := client.CreateAPIWithBasicAuthSecuredSpec(t, mockServiceURL, specUrl, userName, password)
		require.NotEmpty(t, apiID)
		t.Logf("Created service with apiID: %s", apiID)
		defer func() {
			t.Logf("Cleaning up service %s", apiID)
			client.CleanupService(t, apiID)
		}()

		spec, err := client.GetApiSpecWithRetries(t, apiID)
		require.NoError(t, err)

		assertAPISpec(t, spec)

		t.Log("Successfully fetched api spec")
	})

	t.Run("oauth spec url test", func(t *testing.T) {
		clientId := "myUser"
		clientSecret := "mySecret"

		mockServiceURL := testSuit.GetMockServiceURL()
		specUrl := fmt.Sprintf("%s/spec/auth/oauth", mockServiceURL)
		oauthUrl := fmt.Sprintf("%s/auth/oauth/token/%s/%s", mockServiceURL, clientId, clientSecret)

		apiID := client.CreateAPIWithOAuthSecuredSpec(t, mockServiceURL, specUrl, oauthUrl, clientId, clientSecret)

		require.NotEmpty(t, apiID)
		t.Logf("Created service with apiID: %s", apiID)
		defer func() {
			t.Logf("Cleaning up service %s", apiID)
			client.CleanupService(t, apiID)
		}()

		spec, err := client.GetApiSpecWithRetries(t, apiID)
		require.NoError(t, err)

		assertAPISpec(t, spec)

		assert.NotEmpty(t, spec)
		t.Log("Successfully fetched api spec")
	})

	t.Run("oauth with additional headers spec url test", func(t *testing.T) {
		clientId := "myUser"
		clientSecret := "mySecret"

		headerName := "Custom"
		headerValue := "CustomValue"

		headers := map[string][]string{
			headerName: []string{headerValue},
		}

		mockServiceURL := testSuit.GetMockServiceURL()
		specUrl := fmt.Sprintf("%s/spec/auth/oauth", mockServiceURL)
		oauthUrl := fmt.Sprintf("%s/auth/oauthheaders/token/%s/%s/%s/%s", mockServiceURL, clientId, clientSecret, headerName, headerValue)

		apiID := client.CreateAPIWithOAuthWithCustomHeadersSecuredSpec(t, mockServiceURL, specUrl, oauthUrl, clientId, clientSecret, headers)
		require.NotEmpty(t, apiID)
		t.Logf("Created service with apiID: %s", apiID)
		defer func() {
			t.Logf("Cleaning up service %s", apiID)
			client.CleanupService(t, apiID)
		}()

		spec, err := client.GetApiSpecWithRetries(t, apiID)
		require.NoError(t, err)

		assertAPISpec(t, spec)

		assert.NotEmpty(t, spec)
		t.Log("Successfully fetched api spec")
	})

	t.Run("oauth with additional query parameters spec url test", func(t *testing.T) {
		clientId := "myUser"
		clientSecret := "mySecret"

		paramName := "customParam"
		paramValue := "customValue"

		queryParams := map[string][]string{
			paramName: []string{paramValue},
		}

		mockServiceURL := testSuit.GetMockServiceURL()
		specUrl := fmt.Sprintf("%s/spec/auth/oauth", mockServiceURL)
		oauthUrl := fmt.Sprintf("%s/auth/oauthqueryparams/token/%s/%s/%s/%s", mockServiceURL, clientId, clientSecret, paramName, paramValue)

		apiID := client.CreateAPIWithOAuthWithCustomHeadersSecuredSpec(t, mockServiceURL, specUrl, oauthUrl, clientId, clientSecret, queryParams)
		require.NotEmpty(t, apiID)
		t.Logf("Created service with apiID: %s", apiID)
		defer func() {
			t.Logf("Cleaning up service %s", apiID)
			client.CleanupService(t, apiID)
		}()

		spec, err := client.GetApiSpecWithRetries(t, apiID)
		require.NoError(t, err)

		assertAPISpec(t, spec)

		assert.NotEmpty(t, spec)
		t.Log("Successfully fetched api spec")
	})

	t.Run("additional query params in spec test", func(t *testing.T) {
		paramName := "customParam"
		paramValue := "customValue"

		queryParams := map[string][]string{
			paramName: []string{paramValue},
		}

		mockServiceURL := testSuit.GetMockServiceURL()
		specUrl := fmt.Sprintf("%s/spec/queryparams/%s/%s", mockServiceURL, paramName, paramValue)

		apiID := client.CreateAPIWithCustomQueryParamsSpec(t, mockServiceURL, specUrl, queryParams)

		require.NotEmpty(t, apiID)
		t.Logf("Created service with apiID: %s", apiID)
		defer func() {
			t.Logf("Cleaning up service %s", apiID)
			client.CleanupService(t, apiID)
		}()

		spec, err := client.GetApiSpecWithRetries(t, apiID)
		require.NoError(t, err)

		assertAPISpec(t, spec)

		assert.NotEmpty(t, spec)
		t.Log("Successfully fetched api spec")
	})

	t.Run("custom headers in spec test", func(t *testing.T) {
		headerName := "Custom"
		headerValue := "CustomValue"

		headers := map[string][]string{
			headerName: []string{headerValue},
		}

		mockServiceURL := testSuit.GetMockServiceURL()
		specUrl := fmt.Sprintf("%s/spec/headers/%s/%s", mockServiceURL, headerName, headerValue)

		apiID := client.CreateAPIWithCustomHeadersSpec(t, mockServiceURL, specUrl, headers)

		require.NotEmpty(t, apiID)
		t.Logf("Created service with apiID: %s", apiID)
		defer func() {
			t.Logf("Cleaning up service %s", apiID)
			client.CleanupService(t, apiID)
		}()

		spec, err := client.GetApiSpecWithRetries(t, apiID)
		require.NoError(t, err)

		assertAPISpec(t, spec)

		assert.NotEmpty(t, spec)
		t.Log("Successfully fetched api spec")
	})
}

func assertAPISpec(t *testing.T, spec json.RawMessage) {
	var swaggerAPISpec map[string]interface{}
	err := json.Unmarshal(spec, &swaggerAPISpec)
	require.NoError(t, err)

	require.NotEmpty(t, swaggerAPISpec)
	assert.Equal(t, "2.0", swaggerAPISpec["swagger"])
}
