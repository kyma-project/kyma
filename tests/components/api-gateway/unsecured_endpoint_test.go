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

type unsecuredScenario struct {
	namespace   string
	url         string
	apiResource []unstructured.Unstructured
}

func InitializeScenarioUnsecuredEndpoint(ctx *godog.ScenarioContext) {
	scenario, err := CreateUnsecuredScenario()
	if err != nil {
		t.Fatalf("could not initialize unsecure endpoint scenario err=%s", err)
	}
	ctx.Step(`^There is an unsecured endpoint$`, scenario.thereIsAnUnsecuredEndpoint)
	ctx.Step(`^Calling the endpoint with any token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithAnyTokenShouldResultInStatusbetween)
	ctx.Step(`^Calling the endpoint without a token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutTokenShouldResultInStatusbetween)
}

func (u *unsecuredScenario) thereIsAnUnsecuredEndpoint() error {
	return batch.CreateResources(k8sClient, u.apiResource...)
}

func (u *unsecuredScenario) callingTheEndpointWithAnyTokenShouldResultInStatusbetween(arg1, arg2 int) error {
	return helper.CallEndpointWithHeadersWithRetries(anyToken, authorizationHeaderName, u.url, &helpers.StatusPredicate{LowerStatusBound: arg1, UpperStatusBound: arg2})
}

func (u *unsecuredScenario) callingTheEndpointWithoutTokenShouldResultInStatusbetween(arg1, arg2 int) error {
	return helper.CallEndpointWithRetries(u.url, &helpers.StatusPredicate{LowerStatusBound: arg1, UpperStatusBound: arg2})
}

func CreateUnsecuredScenario() (*unsecuredScenario, error) {
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
	noAccessStrategyApiruleResource, err := manifestprocessor.ParseFromFileWithTemplate(noAccessStrategyApiruleFile, manifestsDirectory, resourceSeparator, struct {
		Namespace        string
		NamePrefix       string
		TestID           string
		Domain           string
		GatewayName      string
		GatewayNamespace string
	}{Namespace: namespace, NamePrefix: "unsecured", TestID: testID, Domain: conf.Domain, GatewayName: conf.GatewayName,
		GatewayNamespace: conf.GatewayNamespace})
	if err != nil {
		return nil, fmt.Errorf("failed to process resource manifest files, details %s", err.Error())
	}

	return &unsecuredScenario{namespace: namespace, url: fmt.Sprintf("https://httpbin-%s.%s", testID, conf.Domain), apiResource: noAccessStrategyApiruleResource}, nil
}


//Funnni step bug
func TestApiEndpointUnsecured(t *testing.T) {
	t.Parallel()
	
	apiGatewayOpts := goDogOpts
	apiGatewayOpts.Paths = []string{"features/unsecured_endpoint.feature"}
	apiGatewayOpts.Concurrency = conf.TestConcurency

	unsecuredSuite := godog.TestSuite{
		Name:                "API-Gateway-Unsecured",
		ScenarioInitializer: InitializeScenarioUnsecuredEndpoint,
		Options:             &apiGatewayOpts,
	}

	if unsecuredSuite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
	if os.Getenv(exportResultVar) == "true" {
		generateHTMLReport()
	}
}
