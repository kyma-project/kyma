package api_gateway

import (
	_ "embed"
	"errors"
	"fmt"

	"github.com/cucumber/godog"
	"github.com/kyma-project/kyma/tests/components/api-gateway/gateway-tests/pkg/helpers"
	"github.com/kyma-project/kyma/tests/components/api-gateway/gateway-tests/pkg/jwt"
)

type oauthJWTTwoPathsScenario struct {
	*Scenario
}

func InitializeScenarioOAuth2JWTTwoPaths(ctx *godog.ScenarioContext) {
	mainScenario, err := CreateScenario(jwtAndOauthStrategyApiruleFile, "oauth2-jwt-two-paths")
	if err != nil {
		t.Fatalf("could not initialize unsecure endpoint scenario err=%s", err)
	}

	scenario := oauthJWTTwoPathsScenario{mainScenario}

	ctx.Step(`^OAuth2JWTTwoPaths: There is a deployment secured with OAuth2 on path /headers and JWT on path /image$`, scenario.thereIsAnOauth2Endpoint)
	ctx.Step(`^OAuth2JWTTwoPaths: Calling the "([^"]*)" endpoint without a token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutTokenShouldResultInStatusBeetween)
	ctx.Step(`^OAuth2JWTTwoPaths: Calling the "([^"]*)" endpoint with an invalid token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithInvalidTokenShouldResultInStatusBeetween)
	ctx.Step(`^OAuth2JWTTwoPaths: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithValidTokenShouldResultInStatusBeetween)
}

func (o *oauthJWTTwoPathsScenario) thereIsAnOauth2Endpoint() error {
	return helper.APIRuleWithRetries(batch.CreateResources, batch.UpdateResources, k8sClient, o.apiResource)
}

func (o *oauthJWTTwoPathsScenario) callingTheEndpointWithValidTokenShouldResultInStatusBeetween(path, tokenType string, lower, higher int) error {
	switch tokenType {
	case "JWT":
		tokenJWT, err := jwt.Authenticate(oauth2Cfg.ClientID, jwtConfig.OidcHydraConfig)
		if err != nil {
			return fmt.Errorf("failed to fetch an id_token: %s", err.Error())
		}
		headerVal := fmt.Sprintf("Bearer %s", tokenJWT)

		return helper.CallEndpointWithHeadersWithRetries(headerVal, authorizationHeaderName, fmt.Sprintf("%s%s", o.url, path), &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
	case "OAuth2":
		token, err := getOAUTHToken(*oauth2Cfg)
		if err != nil {
			return err
		}
		headerVal := fmt.Sprintf("Bearer %s", token.AccessToken)

		return helper.CallEndpointWithHeadersWithRetries(headerVal, authorizationHeaderName, fmt.Sprintf("%s%s", o.url, path), &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
	}
	return errors.New("should not happen")
}

func (o *oauthJWTTwoPathsScenario) callingTheEndpointWithInvalidTokenShouldResultInStatusBeetween(path string, lower, higher int) error {
	return helper.CallEndpointWithHeadersWithRetries(anyToken, authorizationHeaderName, fmt.Sprintf("%s%s", o.url, path), &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
}

func (o *oauthJWTTwoPathsScenario) callingTheEndpointWithoutTokenShouldResultInStatusBeetween(path string, lower, higher int) error {
	return helper.CallEndpointWithRetries(fmt.Sprintf("%s%s", o.url, path), &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
}
