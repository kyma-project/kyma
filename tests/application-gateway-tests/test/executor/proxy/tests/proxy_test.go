package tests

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-project/kyma/tests/application-gateway-tests/test/executor/testkit/registry"
	"github.com/kyma-project/kyma/tests/application-gateway-tests/test/executor/testkit/util"

	"github.com/kyma-project/kyma/tests/application-gateway-tests/test/executor/proxy"
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

	t.Run("basic auth spec url test", func(t *testing.T) {
		userName := "myUser"
		password := "mySecret"

		mockServiceURL := testSuit.GetMockServiceURL()
		specUrl := fmt.Sprintf("%s/spec/auth/basic/%s/%s", mockServiceURL, userName, password)

		apiID := client.CreateAPIWithBasicAuthSecuredSpec(t, mockServiceURL, specUrl, userName, password)
		t.Logf("Created service with apiID: %s", apiID)
		defer func() {
			t.Logf("Cleaning up service %s", apiID)
			client.CleanupService(t, apiID)
		}()
		assert.NotEmpty(t, apiID)
	})

}
