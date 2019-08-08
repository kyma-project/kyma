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
// TODO: should the mock service pod be created separetly?

const (
	defaultCheckInterval           = 2 * time.Second
	accessServiceConnectionTimeout = 60 * time.Second
	dnsWaitTime                    = 15 * time.Second

	// TODO - move to mock service?
	oauthValidToken = "valid-token"
)

type APIAccessChecker struct {
	nameResolver *applications.NameResolver
	client       *http.Client
}

func NewAPIAccessChecker(nameResolver *applications.NameResolver) *APIAccessChecker {
	return &APIAccessChecker{
		nameResolver: nameResolver,
		client:       &http.Client{},
	}
}

func (c *APIAccessChecker) AssertAPIAccess(t *testing.T, apis ...*graphql.APIDefinition) {
	t.Log("Waiting for DNS in Istio Proxy...")
	// Wait for Istio Pilot to propagate DNS
	time.Sleep(dnsWaitTime)

	for _, api := range apis {
		path := c.getPathBasedOnAuth(t, api.DefaultAuth)
		response := c.CallAccessService(t, api.ApplicationID, api.ID, path)
		util.RequireStatus(t, http.StatusOK, response)
	}
}

func (c *APIAccessChecker) getPathBasedOnAuth(t *testing.T, auth *graphql.Auth) string {
	if auth == nil {
		return mock.StatusOk.String()
	}

	switch cred := auth.Credential.(type) {
	case *graphql.BasicCredentialData:
		return fmt.Sprintf("%s/%s/%s", mock.BasicAuth, cred.Username, cred.Password)
	case *graphql.OAuthCredentialData:
		return fmt.Sprintf("%s/%s/%s", mock.OAuth, cred.ClientID, cred.ClientSecret)
	default:
		t.Fatalf("Failed to get path based on authentication: unkonw credentials type")
	}

	return ""
}

func (c *APIAccessChecker) CallAccessService(t *testing.T, applicationId, apiId, path string) *http.Response {
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
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusServiceUnavailable {
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
