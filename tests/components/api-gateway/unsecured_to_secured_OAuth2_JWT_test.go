package api_gateway

import (
	_ "embed"
	"errors"
	"fmt"

	"github.com/cucumber/godog"
	"github.com/kyma-project/kyma/tests/components/api-gateway/gateway-tests/pkg/helpers"
	"github.com/kyma-project/kyma/tests/components/api-gateway/gateway-tests/pkg/jwt"
)

type unsecureToSecureScenarioJWT struct {
	*TwoStepScenario
}

func InitializeScenarioUnsecuredToSecuredEndpointJWT(ctx *godog.ScenarioContext) {
	mainScenario, err := CreateTwoStepScenario(noAccessStrategyApiruleFile, jwtAndOauthStrategyApiruleFile, "u2s-jwt-two-paths")
	if err != nil {
		t.Fatalf("could not initialize unsecure endpoint scenario err=%s", err)
	}

	scenario := unsecureToSecureScenarioJWT{mainScenario}

	ctx.Step(`^UnsecureToSecureOAuth2JWT: There is an unsecured API with all paths available without authorization$`, scenario.thereIsAnUnsecuredAPI)
	ctx.Step(`^UnsecureToSecureOAuth2JWT: The endpoint is reachable$`, scenario.theEndpointIsReachable)
	ctx.Step(`^UnsecureToSecureOAuth2JWT: API is secured with OAuth2 on path \/headers and JWT on path \/image$`, scenario.secureWithOAuth2JWT)
	ctx.Step(`^UnsecureToSecureOAuth2JWT: Calling the "([^"]*)" endpoint with an invalid token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithAInvalidTokenShouldResultInStatusBeetween)
	ctx.Step(`^UnsecureToSecureOAuth2JWT: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithAValidTokenShouldResultInStatusBeetween)
	ctx.Step(`^UnsecureToSecureOAuth2JWT: Calling the "([^"]*)" endpoint without a token should result in status between (\d+) and (\d+)$$`, scenario.callingTheEndpointWithoutATokenShouldResultInStatusBeetween)
}

func (u *unsecureToSecureScenarioJWT) thereIsAnUnsecuredAPI() error {
	return helper.APIRuleWithRetries(batch.CreateResources, batch.UpdateResources, k8sClient, u.apiResourceOne)
}

func (u *unsecureToSecureScenarioJWT) theEndpointIsReachable() error {
	return helper.CallEndpointWithRetries(u.url, &helpers.StatusPredicate{LowerStatusBound: 200, UpperStatusBound: 299})
}

func (u *unsecureToSecureScenarioJWT) secureWithOAuth2JWT() error {
	return helper.APIRuleWithRetries(batch.UpdateResources, batch.UpdateResources, k8sClient, u.apiResourceTwo)
}

func (u *unsecureToSecureScenarioJWT) callingTheEndpointWithAInvalidTokenShouldResultInStatusBeetween(path string, lower int, higher int) error {
	return helper.CallEndpointWithHeadersWithRetries(anyToken, authorizationHeaderName, fmt.Sprintf("%s%s", u.url, path), &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
}

func (u *unsecureToSecureScenarioJWT) callingTheEndpointWithAValidTokenShouldResultInStatusBeetween(path, tokenType string, lower int, higher int) error {
	switch tokenType {
	case "JWT":
		tokenJWT, err := jwt.Authenticate(oauth2Cfg.ClientID, jwtConfig.OidcHydraConfig)
		if err != nil {
			return fmt.Errorf("failed to fetch and id_token. %s", err.Error())
		}
		fmt.Printf("-->vladimir, OAuth2_JWT JWT Bearer token: %s", tokenJWT)
		headerVal := fmt.Sprintf("Bearer %s", tokenJWT)

		return helper.CallEndpointWithHeadersWithRetries(headerVal, authorizationHeaderName, fmt.Sprintf("%s%s", u.url, path), &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
	case "OAuth2":
		token, err := getOAUTHToken(*oauth2Cfg)
		if err != nil {
			return err
		}
		fmt.Printf("-->vladimir, OAuth2_JWT OAuth2 Bearer token: %s", token.AccessToken)
		headerVal := fmt.Sprintf("Bearer %s", token.AccessToken)

		return helper.CallEndpointWithHeadersWithRetries(headerVal, authorizationHeaderName, fmt.Sprintf("%s%s", u.url, path), &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
	}
	return errors.New("should not happen")
}

func (u *unsecureToSecureScenarioJWT) callingTheEndpointWithoutATokenShouldResultInStatusBeetween(path string, lower int, higher int) error {
	return helper.CallEndpointWithRetries(fmt.Sprintf("%s%s", u.url, path), &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
}
