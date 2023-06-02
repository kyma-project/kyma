package api_gateway

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
	"github.com/kyma-project/kyma/tests/components/api-gateway/gateway-tests/pkg/helpers"
)

type oauth2Scenario struct {
	*Scenario
}

func InitializeScenarioOAuth2Endpoint(ctx *godog.ScenarioContext) {
	mainScenario, err := CreateScenario(oauth2SecuredEndpointApiruleFile, "oauth2-secured")
	if err != nil {
		t.Fatalf("could not initialize unsecure endpoint scenario err=%s", err)
	}

	scenario := oauth2Scenario{mainScenario}

	ctx.Step(`^OAuth2: There is an endpoint /headers secured with OAuth2 introspection and requiring scope read$`, scenario.thereIsAnOauth2Endpoint)
	ctx.Step(`^OAuth2: There is an endpoint /ip secured with OAuth2 introspection and requiring scope special-scope$`, func() {})
	ctx.Step(`^OAuth2: Calling the "([^"]*)" endpoint without a token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutTokenShouldResultInStatusBeetween)
	ctx.Step(`^OAuth2: Calling the "([^"]*)" endpoint with a invalid token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithInvalidTokenShouldResultInStatusBeetween)
	ctx.Step(`^OAuth2: Calling the "([^"]*)" endpoint with a valid token with scope claim read should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithValidTokenShouldResultInStatusBeetween)
}

func (o *oauth2Scenario) thereIsAnOauth2Endpoint() error {
	return helper.APIRuleWithRetries(batch.CreateResources, batch.UpdateResources, k8sClient, o.apiResource)
}

func (o *oauth2Scenario) callingTheEndpointWithValidTokenShouldResultInStatusBeetween(endpoint string, lower, higher int) error {
	token, err := getOAUTHToken(*oauth2Cfg)
	if err != nil {
		return err
	}

	headerVal := fmt.Sprintf("Bearer %s", token.AccessToken)
	url := getUrlWithEndpoint(o.url, endpoint)
	return helper.CallEndpointWithHeadersWithRetries(headerVal, authorizationHeaderName, url, &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
}

func (o *oauth2Scenario) callingTheEndpointWithInvalidTokenShouldResultInStatusBeetween(endpoint string, lower, higher int) error {
	url := getUrlWithEndpoint(o.url, endpoint)
	return helper.CallEndpointWithHeadersWithRetries(anyToken, authorizationHeaderName, url, &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
}

func (o *oauth2Scenario) callingTheEndpointWithoutTokenShouldResultInStatusBeetween(endpoint string, lower, higher int) error {
	url := getUrlWithEndpoint(o.url, endpoint)
	return helper.CallEndpointWithRetries(url, &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
}

func getUrlWithEndpoint(url, endpoint string) string {
	return fmt.Sprintf("%s/%s", url, strings.TrimLeft(endpoint, "/"))
}
