package api_gateway

import (
	_ "embed"
	"fmt"

	"github.com/cucumber/godog"
	"github.com/kyma-project/kyma/tests/components/api-gateway/gateway-tests/pkg/helpers"
)

type unsecureToSecureScenario struct {
	*TwoStepScenario
}

func InitializeScenarioUnsecuredToSecuredEndpoint(ctx *godog.ScenarioContext) {
	mainScenario, err := CreateTwoStepScenario(noAccessStrategyApiruleFile, oauthStrategyApiruleFile, "unsecured-to-secured-oauth2")
	if err != nil {
		t.Fatalf("could not initialize unsecure endpoint scenario err=%s", err)
	}

	scenario := unsecureToSecureScenario{mainScenario}

	ctx.Step(`^UnsecureToSecureOAuth2: There is an unsecured API with all paths available without authorization$`, scenario.thereIsAnUnsecuredAPI)
	ctx.Step(`^UnsecureToSecureOAuth2: The endpoint is reachable$`, scenario.theEndpointIsReachable)
	ctx.Step(`^UnsecureToSecureOAuth2: Endpoint is secured with OAuth2$`, scenario.secureWithOAuth2)
	ctx.Step(`^UnsecureToSecureOAuth2: Calling the endpoint with a invalid token should result in status beetween (\d+) and (\d+)$`, scenario.callingTheEndpointWithAInvalidTokenShouldResultInStatusBeetween)
	ctx.Step(`^UnsecureToSecureOAuth2: Calling the endpoint with a valid token should result in status beetween (\d+) and (\d+)$`, scenario.callingTheEndpointWithAValidTokenShouldResultInStatusBeetween)
	ctx.Step(`^UnsecureToSecureOAuth2: Calling the endpoint without a token should result in status beetween (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutATokenShouldResultInStatusBeetween)
}

func (u *unsecureToSecureScenario) thereIsAnUnsecuredAPI() error {
	return helper.APIRuleWithRetries(batch.CreateResources, k8sClient, u.apiResourceOne)
}

func (u *unsecureToSecureScenario) theEndpointIsReachable() error {
	return helper.CallEndpointWithRetries(u.url, &helpers.StatusPredicate{LowerStatusBound: 200, UpperStatusBound: 299})
}

func (u *unsecureToSecureScenario) secureWithOAuth2() error {
	return helper.APIRuleWithRetries(batch.UpdateResources, k8sClient, u.apiResourceTwo)
}

func (u *unsecureToSecureScenario) callingTheEndpointWithAInvalidTokenShouldResultInStatusBeetween(lower int, higher int) error {
	return helper.CallEndpointWithHeadersWithRetries(anyToken, authorizationHeaderName, u.url, &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
}

func (u *unsecureToSecureScenario) callingTheEndpointWithAValidTokenShouldResultInStatusBeetween(lower int, higher int) error {
	token, err := getOAUTHToken(*oauth2Cfg)
	if err != nil {
		return err
	}

	headerVal := fmt.Sprintf("Bearer %s", token.AccessToken)

	return helper.CallEndpointWithHeadersWithRetries(headerVal, authorizationHeaderName, u.url, &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
}

func (u *unsecureToSecureScenario) callingTheEndpointWithoutATokenShouldResultInStatusBeetween(lower int, higher int) error {
	return helper.CallEndpointWithRetries(u.url, &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
}
