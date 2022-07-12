package api_gateway

import (
	_ "embed"
	"fmt"

	"github.com/cucumber/godog"
	"github.com/kyma-project/kyma/tests/components/api-gateway/gateway-tests/pkg/helpers"
)

type oauth2Scenario struct {
	*Scenario
}

func InitializeScenarioOAuth2Endpoint(ctx *godog.ScenarioContext) {
	mainScenario, err := CreateScenario(oauthStrategyApiruleFile, "oauth2-secured")
	if err != nil {
		t.Fatalf("could not initialize unsecure endpoint scenario err=%s", err)
	}

	scenario := oauth2Scenario{mainScenario}

	ctx.Step(`^OAuth2: There is an endpoint secured with OAuth2 introspection$`, scenario.thereIsAnOauth2Endpoint)
	ctx.Step(`^OAuth2: Calling the endpoint without a token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutTokenShouldResultInStatusBeetween)
	ctx.Step(`^OAuth2: Calling the endpoint with a invalid token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithInvalidTokenShouldResultInStatusBeetween)
	ctx.Step(`^OAuth2: Calling the endpoint with a valid token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithValidTokenShouldResultInStatusBeetween)
}

func (o *oauth2Scenario) thereIsAnOauth2Endpoint() error {
	return helper.APIRuleWithRetries(batch.CreateResources, k8sClient, o.apiResource)
}

func (o *oauth2Scenario) callingTheEndpointWithValidTokenShouldResultInStatusBeetween(lower, higher int) error {
	token, err := getOAUTHToken(*oauth2Cfg)
	if err != nil {
		return err
	}

	headerVal := fmt.Sprintf("Bearer %s", token.AccessToken)

	return helper.CallEndpointWithHeadersWithRetries(headerVal, authorizationHeaderName, o.url, &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
}

func (o *oauth2Scenario) callingTheEndpointWithInvalidTokenShouldResultInStatusBeetween(lower, higher int) error {
	return helper.CallEndpointWithHeadersWithRetries(anyToken, authorizationHeaderName, o.url, &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
}

func (o *oauth2Scenario) callingTheEndpointWithoutTokenShouldResultInStatusBeetween(lower, higher int) error {
	return helper.CallEndpointWithRetries(o.url, &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
}
