package api_gateway

import (
	_ "embed"
	"fmt"
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/kyma-project/kyma/tests/components/api-gateway/gateway-tests/pkg/helpers"
	"github.com/kyma-project/kyma/tests/components/api-gateway/gateway-tests/pkg/manifestprocessor"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func InitializeScenarioOAuth2Endpoint(ctx *godog.ScenarioContext) {
	scenario, err := CreateOAuth2Scenario()
	if err != nil {
		t.Fatalf("could not initialize unsecure endpoint scenario err=%s", err)
	}
	ctx.Step(`^There is an endpoint secured with OAuth2 introspection$`, scenario.thereIsAnOauth2Endpoint)
	ctx.Step(`^Calling the endpoint without a token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutTokenShouldResultInStatusBeetween)
	ctx.Step(`^Calling the endpoint with a invalid token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithInvalidTokenShouldResultInStatusBeetween)
	ctx.Step(`^Calling the endpoint with a valid token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithValidTokenShouldResultInStatusBeetween)
}

type oauth2Scenario struct {
	namespace   string
	url         string
	apiResource []unstructured.Unstructured
}

func (o *oauth2Scenario) thereIsAnOauth2Endpoint() error {
	return batch.CreateResources(k8sClient, o.apiResource...)
}

func (o *oauth2Scenario) callingTheEndpointWithValidTokenShouldResultInStatusBeetween(arg1, arg2 int) error {
	token, err := getOAUTHToken(*oauth2Cfg)
	if err != nil {
		return err
	}

	headerVal := fmt.Sprintf("Bearer %s", token.AccessToken)

	return helper.CallEndpointWithHeadersWithRetries(headerVal, authorizationHeaderName, o.url, &helpers.StatusPredicate{LowerStatusBound: arg1, UpperStatusBound: arg2})
}

func (o *oauth2Scenario) callingTheEndpointWithInvalidTokenShouldResultInStatusBeetween(arg1, arg2 int) error {
	return helper.CallEndpointWithHeadersWithRetries(anyToken, authorizationHeaderName, o.url, &helpers.StatusPredicate{LowerStatusBound: arg1, UpperStatusBound: arg2})
}

func (o *oauth2Scenario) callingTheEndpointWithoutTokenShouldResultInStatusBeetween(arg1, arg2 int) error {
	return helper.CallEndpointWithRetries(o.url, &helpers.StatusPredicate{LowerStatusBound: arg1, UpperStatusBound: arg2})
}

func CreateOAuth2Scenario() (*oauth2Scenario, error) {
	testID := generateRandomString(testIDLength)

	// create common resources from files
	commonResources, err := manifestprocessor.ParseFromFileWithTemplate(testingAppFile, manifestsDirectory, resourceSeparator, struct {
		Namespace string
		TestID    string
	}{
		Namespace: namespace,
		TestID:    testID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to process common manifest files, details %s", err.Error())
	}
	err = batch.CreateResources(k8sClient, commonResources...)

	if err != nil {
		return nil, err
	}

	// create api-rule from file
	oauth2AccessStrategyApiruleResource, err := manifestprocessor.ParseFromFileWithTemplate(oauthStrategyApiruleFile, manifestsDirectory, resourceSeparator, struct {
		Namespace        string
		NamePrefix       string
		TestID           string
		Domain           string
		GatewayName      string
		GatewayNamespace string
	}{Namespace: namespace, NamePrefix: "oauth2", TestID: testID, Domain: conf.Domain, GatewayName: conf.GatewayName,
		GatewayNamespace: conf.GatewayNamespace})
	if err != nil {
		return nil, fmt.Errorf("failed to process resource manifest files, details %s", err.Error())
	}

	return &oauth2Scenario{namespace: namespace, url: fmt.Sprintf("https://httpbin-%s.%s", testID, conf.Domain), apiResource: oauth2AccessStrategyApiruleResource}, nil
}

func TestApiEndpointOAuth2(t *testing.T) {
	t.Parallel()
	
	apiGatewayOpts := goDogOpts
	apiGatewayOpts.Paths = []string{"features/oauth2_secured_endpoint.feature"}
	apiGatewayOpts.Concurrency = conf.TestConcurency

	unsecuredSuite := godog.TestSuite{
		Name:                "API-Gateway-OAuth2",
		ScenarioInitializer: InitializeScenarioOAuth2Endpoint,
		Options:             &apiGatewayOpts,
	}

	if unsecuredSuite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
	if os.Getenv(exportResultVar) == "true" {
		generateHTMLReport()
	}
}
