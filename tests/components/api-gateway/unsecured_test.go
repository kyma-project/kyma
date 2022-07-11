package api_gateway

import (
	_ "embed"

	"github.com/cucumber/godog"
	"github.com/kyma-project/kyma/tests/components/api-gateway/gateway-tests/pkg/helpers"
)

type unsecuredScenario struct {
	*Scenario
}

func InitializeScenarioUnsecuredEndpoint(ctx *godog.ScenarioContext) {
	mainScenario, err := CreateScenario(noAccessStrategyApiruleFile, "unsecured")
	if err != nil {
		t.Fatalf("could not initialize unsecure endpoint scenario err=%s", err)
	}

	scenario := unsecuredScenario{mainScenario}

	ctx.Step(`^Unsecured: There is an unsecured endpoint$`, scenario.thereIsAnUnsecuredEndpoint)
	ctx.Step(`^Unsecured: Calling the endpoint with any token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithAnyTokenShouldResultInStatusbetween)
	ctx.Step(`^Unsecured: Calling the endpoint without a token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutTokenShouldResultInStatusbetween)
}

func (u *unsecuredScenario) thereIsAnUnsecuredEndpoint() error {
	return helper.APIRuleWithRetries(batch.CreateResources, k8sClient, u.apiResource)
}

func (u *unsecuredScenario) callingTheEndpointWithAnyTokenShouldResultInStatusbetween(arg1, arg2 int) error {
	return helper.CallEndpointWithHeadersWithRetries(anyToken, authorizationHeaderName, u.url, &helpers.StatusPredicate{LowerStatusBound: arg1, UpperStatusBound: arg2})
}

func (u *unsecuredScenario) callingTheEndpointWithoutTokenShouldResultInStatusbetween(arg1, arg2 int) error {
	return helper.CallEndpointWithRetries(u.url, &helpers.StatusPredicate{LowerStatusBound: arg1, UpperStatusBound: arg2})
}
