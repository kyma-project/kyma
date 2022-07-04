package api_gateway

import (
	_ "embed"
	"fmt"

	"github.com/cucumber/godog"
	"github.com/kyma-project/kyma/tests/components/api-gateway/gateway-tests/pkg/helpers"
	"github.com/kyma-project/kyma/tests/components/api-gateway/gateway-tests/pkg/manifestprocessor"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func InitializeScenarioUnsecuredEndpoint(ctx *godog.ScenarioContext) {
	scenario := unsecuredScenario{namespace: namespace}
	ctx.Step(`^There is an unsecured endpoint$`, scenario.thereIsAnUnsecuredEndpoint)
	ctx.Step(`^Calling the endpoint with any token should result in status beetween (\d+) and (\d+)$`, scenario.callingTheEndpointShouldResultInStatusBeetween)
	ctx.Step(`^Calling the endpoint without a token should result in status beetween (\d+) and (\d+)$`, scenario.callingTheEndpointShouldResultInStatusBeetween)
}

type unsecuredScenario struct {
	namespace   string
	url         string
	apiResource []unstructured.Unstructured
}

func (u *unsecuredScenario) thereIsAnUnsecuredEndpoint() error {
	return batch.CreateResources(k8sClient, u.apiResource...)
}

func (u *unsecuredScenario) callingTheEndpointShouldResultInStatusBeetween(arg1, arg2 int) error {
	return helper.CallEndpointsWithRetries(u.url, &helpers.StatusPredicate{LowerStatusBound: arg1, UpperStatusBound: arg2})
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
	batch.CreateResources(k8sClient, commonResources...)

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
