package api_gateway

import (
	_ "embed"
	"fmt"

	"github.com/cucumber/godog"
	"github.com/kyma-project/kyma/tests/components/api-gateway/gateway-tests/pkg/helpers"
)

type secureToUnsecureScenario struct {
	*TwoStepScenario
}

func InitializeScenarioSecuredToUnsecuredEndpoint(ctx *godog.ScenarioContext) {
	mainScenario, err := CreateTwoStepScenario(oauthStrategyApiruleFile, noAccessStrategyApiruleFile, "secured-to-unsecured")
	if err != nil {
		t.Fatalf("could not initialize unsecure endpoint scenario err=%s", err)
	}

	scenario := secureToUnsecureScenario{mainScenario}

	ctx.Step(`^SecureToUnsecure: There is an endpoint secured with OAuth2$`, scenario.thereIsASecuredOAuth2Endpoint)
	ctx.Step(`^SecureToUnsecure: The endpoint is reachable$`, scenario.theEndpointIsReachable)
	ctx.Step(`^SecureToUnsecure: Endpoint is exposed with noop strategy$`, scenario.unsecureTheEndpoint)
	ctx.Step(`^SecureToUnsecure: Calling the endpoint with any token should result in status beetween (\d+) and (\d+)$`, scenario.callingTheEndpointWithAnyTokenShouldResultInStatusBeetween)
	ctx.Step(`^SecureToUnsecure: Calling the endpoint without a token should result in status beetween (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutATokenShouldResultInStatusBeetween)
}

func (u *secureToUnsecureScenario) thereIsASecuredOAuth2Endpoint() error {
	return helper.APIRuleWithRetries(batch.CreateResources, batch.UpdateResources, k8sClient, u.apiResourceOne)
}

func (u *secureToUnsecureScenario) theEndpointIsReachable() error {
	token, err := getOAUTHToken(*oauth2Cfg)
	if err != nil {
		return err
	}

	headerVal := fmt.Sprintf("Bearer %s", token.AccessToken)

	return helper.CallEndpointWithHeadersWithRetries(headerVal, authorizationHeaderName, u.url, &helpers.StatusPredicate{LowerStatusBound: 200, UpperStatusBound: 299})
}

func (u *secureToUnsecureScenario) unsecureTheEndpoint() error {
	return helper.APIRuleWithRetries(batch.UpdateResources, batch.UpdateResources, k8sClient, u.apiResourceTwo)
}

func (u *secureToUnsecureScenario) callingTheEndpointWithAnyTokenShouldResultInStatusBeetween(lower int, higher int) error {
	return helper.CallEndpointWithHeadersWithRetries(anyToken, authorizationHeaderName, u.url, &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
}

func (u *secureToUnsecureScenario) callingTheEndpointWithoutATokenShouldResultInStatusBeetween(lower int, higher int) error {
	return helper.CallEndpointWithRetries(u.url, &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
}
