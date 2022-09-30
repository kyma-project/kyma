package api_gateway

import (
	_ "embed"

	"github.com/cucumber/godog"
	"github.com/kyma-project/kyma/tests/components/api-gateway/gateway-tests/pkg/helpers"
)

type twoServiceScenario struct {
	*Scenario
}

func InitializeScenarioTwoServices(ctx *godog.ScenarioContext) {
	mainScenario, err := CreateScenario(twoServicesApiruleFile, "unsecured", twoServicesDeploymentFile)
	if err != nil {
		t.Fatalf("could not initialize unsecure endpoint scenario err=%s", err)
	}

	scenario := twoServiceScenario{mainScenario}

	ctx.Step(`^Service per path: There are two endpoints exposed with different services$`, scenario.thereAreTwoEndpointsExposedWithDifferentServices)
	ctx.Step(`^Service per path: Calling the endpoint "([^"]*)" and "([^"]*)" with any token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointsWithAnyTokenShouldResultInStatusbetween)
	ctx.Step(`^Service per path: Calling the endpoint "([^"]*)" and "([^"]*)" without a token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointsWithoutTokenShouldResultInStatusbetween)
}

func (u *twoServiceScenario) thereAreTwoEndpointsExposedWithDifferentServices() error {
	return helper.APIRuleWithRetries(batch.CreateResources, batch.UpdateResources, k8sClient, u.apiResource)
}

func (u *twoServiceScenario) callingTheEndpointsWithAnyTokenShouldResultInStatusbetween(path1, path2 string, arg1, arg2 int) error {
	return helper.CallEndpointWithHeadersWithRetries(anyToken, authorizationHeaderName, u.url, &helpers.StatusPredicate{LowerStatusBound: arg1, UpperStatusBound: arg2})
}

func (u *twoServiceScenario) callingTheEndpointsWithoutTokenShouldResultInStatusbetween(path1, path2 string, arg1, arg2 int) error {
	return helper.CallEndpointWithRetries(u.url, &helpers.StatusPredicate{LowerStatusBound: arg1, UpperStatusBound: arg2})
}
