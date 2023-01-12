package api_gateway

import (
	_ "embed"
	"errors"
	"fmt"

	"github.com/cucumber/godog"
	"github.com/kyma-project/kyma/tests/components/api-gateway/gateway-tests/pkg/helpers"
	"github.com/kyma-project/kyma/tests/components/api-gateway/gateway-tests/pkg/jwt"
)

type istioJwtScenario struct {
	*Scenario
}

func InitializeScenarioIstioJWT(ctx *godog.ScenarioContext) {
	mainScenario, err := CreateScenario(istioJwtApiruleFile, "istio-jwt")
	if err != nil {
		t.Fatalf("could not initialize unsecure endpoint scenario err=%s", err)
	}

	scenario := istioJwtScenario{mainScenario}

	ctx.Step(`^IstioJWT: There is an deployment secured with JWT on path /image$`, scenario.thereIsAnEndpoint)
	ctx.Step(`^IstioJWT: Calling the "([^"]*)" endpoint without a token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutTokenShouldResultInStatusBeetween)
	ctx.Step(`^IstioJWT: Calling the "([^"]*)" endpoint with a invalid token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithInvalidTokenShouldResultInStatusBeetween)
	ctx.Step(`^IstioJWT: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithValidTokenShouldResultInStatusBeetween)
}

func (o *istioJwtScenario) thereIsAnEndpoint() error {
	print("test")
	return helper.APIRuleWithRetries(batch.CreateResources, batch.UpdateResources, k8sClient, o.apiResource)
}

func (o *istioJwtScenario) callingTheEndpointWithValidTokenShouldResultInStatusBeetween(path, tokenType string, lower, higher int) error {
	switch tokenType {
	case "JWT":
		tokenJWT, err := jwt.Authenticate(oauth2Cfg.ClientID, jwtConfig.OidcHydraConfig)
		if err != nil {
			return fmt.Errorf("failed to fetch an id_token: %s", err.Error())
		}
		headerVal := fmt.Sprintf("Bearer %s", tokenJWT)

		return helper.CallEndpointWithHeadersWithRetries(headerVal, authorizationHeaderName, fmt.Sprintf("%s%s", o.url, path), &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
	}
	return errors.New("should not happen")
}

func (o *istioJwtScenario) callingTheEndpointWithInvalidTokenShouldResultInStatusBeetween(path string, lower, higher int) error {
	return helper.CallEndpointWithHeadersWithRetries(anyToken, authorizationHeaderName, fmt.Sprintf("%s%s", o.url, path), &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
}

func (o *istioJwtScenario) callingTheEndpointWithoutTokenShouldResultInStatusBeetween(path string, lower, higher int) error {
	return helper.CallEndpointWithRetries(fmt.Sprintf("%s%s", o.url, path), &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
}
