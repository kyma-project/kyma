package api_gateway

import (
	_ "embed"

	"github.com/cucumber/godog"
	"github.com/kyma-project/kyma/tests/components/api-gateway/gateway-tests/pkg/helpers"
)

type mtlsGatewayScenario struct {
	*Scenario
}

func InitializeScenariomTLSGateway(ctx *godog.ScenarioContext) {
	mainScenario, err := CreateScenario(noAccessStrategyApiruleFile, "mtlsgateway")
	if err != nil {
		t.Fatalf("could not initialize unsecure endpoint scenario err=%s", err)
	}

	scenario := mtlsGatewayScenario{mainScenario}

	ctx.Step(`^mTLSGateway: There is an unsecured endpoint on "([^"]*)" gateway$`, scenario.thereIsAnUnsecuredEndpoint)
	ctx.Step(`^mTLSGateway: Calling the endpoint with any token should result in status between (\d+) and (\d+) on "([^"]*)" gateway$`, scenario.callingTheEndpointWithAnyTokenShouldResultInStatusbetween)
	ctx.Step(`^mTLSGateway: Calling the endpoint without a token should result in status between (\d+) and (\d+) on "([^"]*)" gateway$`, scenario.callingTheEndpointWithoutTokenShouldResultInStatusbetween)
}

func (u *mtlsGatewayScenario) thereIsAnUnsecuredEndpoint(gateway string) error {
	return helper.APIRuleWithRetries(batch.CreateResources, batch.UpdateResources, k8sClient, u.apiResource)
}

func (u *mtlsGatewayScenario) callingTheEndpointWithAnyTokenShouldResultInStatusbetween(arg1, arg2 int, gateway string) error {
	return helper.CallEndpointWithHeadersWithRetries(anyToken, authorizationHeaderName, u.url, &helpers.StatusPredicate{LowerStatusBound: arg1, UpperStatusBound: arg2})
}

func (u *mtlsGatewayScenario) callingTheEndpointWithoutTokenShouldResultInStatusbetween(arg1, arg2 int, gateway string) error {
	return helper.CallEndpointWithRetries(u.url, &helpers.StatusPredicate{LowerStatusBound: arg1, UpperStatusBound: arg2})
}
