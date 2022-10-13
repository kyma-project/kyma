package api_gateway

import (
	_ "embed"
	"fmt"

	"github.com/cucumber/godog"
	"github.com/kyma-project/kyma/tests/components/api-gateway/gateway-tests/pkg/helpers"
)

type servicePerPathScenario struct {
	*Scenario
}

func InitializeScenarioServicePerPath(ctx *godog.ScenarioContext) {
	mainScenario, err := CreateScenario(twoServicesApiruleFile, "two-service-per-path", twoServicesDeploymentFile)
	if err != nil {
		t.Fatalf("could not initialize unsecure endpoint scenario err=%s", err)
	}

	scenario := servicePerPathScenario{mainScenario}

	ctx.Step(`^Service per path: There are two endpoints exposed with different services$`, scenario.thereAreTwoEndpointsExposedWithDifferentServices)
	ctx.Step(`^Service per path: Calling the endpoint "([^"]*)" and "([^"]*)" with any token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointsWithAnyTokenShouldResultInStatusbetween)
	ctx.Step(`^Service per path: Calling the endpoint "([^"]*)" and "([^"]*)" without a token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointsWithoutTokenShouldResultInStatusbetween)
}

func (u *servicePerPathScenario) thereAreTwoEndpointsExposedWithDifferentServices() error {
	return helper.APIRuleWithRetries(batch.CreateResources, batch.UpdateResources, k8sClient, u.apiResource)
}

func (u *servicePerPathScenario) callingTheEndpointsWithAnyTokenShouldResultInStatusbetween(path1, path2 string, arg1, arg2 int) error {
	err := helper.CallEndpointWithHeadersWithRetries(anyToken, authorizationHeaderName, fmt.Sprintf("%s%s", u.url, path1), &helpers.StatusPredicate{LowerStatusBound: arg1, UpperStatusBound: arg2})
	if err != nil {
		return err
	}
	return helper.CallEndpointWithHeadersWithRetries(anyToken, authorizationHeaderName, fmt.Sprintf("%s%s", u.url, path2), &helpers.StatusPredicate{LowerStatusBound: arg1, UpperStatusBound: arg2})
}

func (u *servicePerPathScenario) callingTheEndpointsWithoutTokenShouldResultInStatusbetween(path1, path2 string, arg1, arg2 int) error {
	err := helper.CallEndpointWithRetries(fmt.Sprintf("%s%s", u.url, path1), &helpers.StatusPredicate{LowerStatusBound: arg1, UpperStatusBound: arg2})
	if err != nil {
		return err
	}
	return helper.CallEndpointWithRetries(fmt.Sprintf("%s%s", u.url, path2), &helpers.StatusPredicate{LowerStatusBound: arg1, UpperStatusBound: arg2})
}
