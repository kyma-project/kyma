package assertions

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit"

	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/kymaconfig"

	"github.com/kyma-project/kyma/tests/application-gateway-tests/test/tools"

	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/mock"

	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/util"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

// TODO: we should consider enhancing test with sending events (also use Mock Service)

const (
	defaultCheckInterval = 2 * time.Second

	gatewayConnectionTimeout     = 30 * time.Second
	appGatewayHealthCheckTimeout = 30 * time.Second
)

type ProxyAPIAccessChecker struct {
	client           *http.Client
	namespace        string
	kymaConfigurator *kymaconfig.KymaConfigurator
}

func NewAPIAccessChecker(
	namespace string,
	kymaConfigurator *kymaconfig.KymaConfigurator,
) *ProxyAPIAccessChecker {
	return &ProxyAPIAccessChecker{
		namespace:        namespace,
		kymaConfigurator: kymaConfigurator,
		client:           &http.Client{},
	}
}

func (c *ProxyAPIAccessChecker) AssertAPIAccess(t *testing.T, log *testkit.Logger, applicationId string, packages []*graphql.PackageExt) {
	log.Log("Provisioning Service Instances and Service Bindings...")
	secretMapping := c.kymaConfigurator.ConfigureApplication(t, log, applicationId, packages)
	defer func() {
		log.Log("Cleaning up Service Instances and Service Bindings...")
		c.kymaConfigurator.CleanupConfiguration(t, packages...)
	}()

	// Wait for Gateway to be healthy
	c.checkApplicationGatewayHealth(t)

	for _, apiPackage := range packages {
		for _, api := range apiPackage.APIDefinitions.Data {
			c.accessAPI(t, secretMapping.PackagesSecrets[apiPackage.ID], api)
		}
	}
}

func (c *ProxyAPIAccessChecker) accessAPI(t *testing.T, secretName string, api *graphql.APIDefinitionExt) {
	path := c.GetPathBasedOnAuth(t, api.DefaultAuth)
	apiNamePath := createAPIName(api)
	response := c.CallAPIThroughGateway(t, secretName, apiNamePath, path)
	defer response.Body.Close()
	util.RequireStatus(t, http.StatusOK, response)
}

func (c *ProxyAPIAccessChecker) checkApplicationGatewayHealth(t *testing.T) {
	t.Log("Checking application gateway health...")

	healthURL := c.gatewayHealthURL()
	err := tools.WaitForFunction(defaultCheckInterval, appGatewayHealthCheckTimeout, func() bool {
		req, err := http.NewRequest(http.MethodGet, healthURL, nil)
		if err != nil {
			return false
		}

		res, err := c.client.Do(req)
		if err != nil {
			return false
		}

		if res.StatusCode != http.StatusOK {
			return false
		}

		return true
	})

	require.NoError(t, err, "Failed to check health of Application Gateway.")
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

func (c *ProxyAPIAccessChecker) gatewayURL() string {
	return fmt.Sprintf("http://%s-gateway.%s.svc.cluster.local:8080", c.namespace, c.namespace)
}

func (c *ProxyAPIAccessChecker) gatewayHealthURL() string {
	return fmt.Sprintf("http://%s-gateway.%s.svc.cluster.local:8081/v1/health", c.namespace, c.namespace)
}

func (c *ProxyAPIAccessChecker) CallAPIThroughGateway(t *testing.T, secretName, apiNamePath, path string) *http.Response {
	url := fmt.Sprintf("%s/secret/%s/api/%s%s", c.gatewayURL(), secretName, apiNamePath, path)

	var resp *http.Response

	err := tools.WaitForFunction(defaultCheckInterval, gatewayConnectionTimeout, func() bool {
		t.Logf("Accessing Gateway at: %s", url)
		var err error

		resp, err = http.Get(url)
		if err != nil {
			t.Logf("Failed to access Gateway: %s", err.Error())
			return false
		}
		defer resp.Body.Close()

		fmt.Println("GATEWAY CALL:")
		d, _ := httputil.DumpResponse(resp, true)
		fmt.Println(string(d))

		if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusServiceUnavailable {
			t.Logf("Invalid response from Gateway, status: %d.", resp.StatusCode)
			bytes, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)
			t.Log(string(bytes))
			t.Logf("Gateway is not ready. Retrying.")
			return false
		}

		return true
	})
	require.NoError(t, err)

	return resp
}

// prefix returns a valid environment variable name prefix which consist of alphabetic characters, digits, '_' and does not start with a digit
func createAPIName(api *graphql.APIDefinitionExt) string {
	sanitizedName := sanitizeName(api.Name)
	sanitizedID := sanitizeID(api.ID)

	return strings.ToUpper(fmt.Sprintf("%s_%s", sanitizedName, sanitizedID))
}

var (
	whitespaces                  = regexp.MustCompile(`\s+`)
	dash                         = regexp.MustCompile(`-+`)
	envNameAllowedChars          = regexp.MustCompile(`[a-zA-Z_]+[a-zA-Z0-9_\s]*`)
	envNameSubstringAllowedChars = regexp.MustCompile(`[a-zA-Z0-9_]*`)
)

func sanitizeName(in string) string {
	// remove not allowed characters like @,#,$ etc.
	in = strings.Join(envNameAllowedChars.FindAllString(in, -1), "")
	// remove leading and trailing white space
	in = strings.TrimSpace(in)
	// replace rest white space between words with underscore
	in = whitespaces.ReplaceAllString(in, "_")

	return in
}

func sanitizeID(in string) string {
	// replace dash in UUID with underscores
	in = dash.ReplaceAllString(in, "_")
	// ensure that not allowed characters are removed (just in case)
	in = strings.Join(envNameSubstringAllowedChars.FindAllString(in, -1), "")

	return in
}
