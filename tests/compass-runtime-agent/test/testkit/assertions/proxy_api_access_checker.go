package assertions

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/mock"
	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit"

	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/util"

	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/applications"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

// TODO: we should consider enhancing test with sending events (also use Mock Service)

const (
	defaultCheckInterval           = 2 * time.Second
	accessServiceConnectionTimeout = 90 * time.Second
	dnsWaitTime                    = 15 * time.Second
)

type ProxyAPIAccessChecker struct {
	nameResolver *applications.NameResolver
	client       *http.Client
}

func NewAPIAccessChecker(nameResolver *applications.NameResolver) *ProxyAPIAccessChecker {
	return &ProxyAPIAccessChecker{
		nameResolver: nameResolver,
		client:       &http.Client{},
	}
}

func (c *ProxyAPIAccessChecker) AssertAPIAccess(t *testing.T, applicationName string, apis ...*graphql.APIDefinition) {
	t.Log("Waiting for DNS in Istio Proxy...")
	// Wait for Istio Pilot to propagate DNS
	time.Sleep(dnsWaitTime)

	for _, api := range apis {
		c.accessAPI(t, applicationName, api)
	}
}

func (c *ProxyAPIAccessChecker) accessAPI(t *testing.T, applicationName string, api *graphql.APIDefinition) {
	path := c.GetPathBasedOnAuth(t, api.DefaultAuth)
	response := c.CallAccessService(t, applicationName, api.ID, path)
	defer response.Body.Close()
	util.RequireStatus(t, http.StatusOK, response)
}

func (c *ProxyAPIAccessChecker) GetPathBasedOnAuth(t *testing.T, auth *graphql.Auth) string {
	if auth == nil {
		return mock.StatusOk.String()
	}

	switch cred := auth.Credential.(type) {
	case *graphql.BasicCredentialData:
		return fmt.Sprintf("%s/%s/%s", mock.BasicAuth, cred.Username, cred.Password)
	case *graphql.OAuthCredentialData:
		return fmt.Sprintf("%s/%s/%s", mock.OAuth, cred.ClientID, cred.ClientSecret)
	default:
		t.Fatalf("Failed to get path based on authentication: unknow credentials type")
	}

	return ""
}

func (c *ProxyAPIAccessChecker) CallAccessService(t *testing.T, applicationId, apiId, path string) *http.Response {
	gatewayURL := c.nameResolver.GetGatewayUrl(applicationId, apiId)
	url := fmt.Sprintf("%s%s", gatewayURL, path)

	var resp *http.Response

	err := testkit.WaitForFunction(defaultCheckInterval, accessServiceConnectionTimeout, func() bool {
		t.Logf("Accessing proxy at: %s", url)
		var err error

		resp, err = http.Get(url)
		if err != nil {
			t.Logf("Access service not ready: %s", err.Error())
			return false
		}

		if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusServiceUnavailable {
			defer resp.Body.Close()
			t.Logf("Invalid response from access service, status: %d.", resp.StatusCode)
			bytes, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)
			t.Log(string(bytes))
			t.Logf("Access service is not ready. Retrying.")
			return false
		}

		return true
	})
	require.NoError(t, err)

	return resp
}
