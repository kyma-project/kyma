package api_gateway

import (
	_ "embed"
	"errors"
	"fmt"

	"github.com/cucumber/godog"
	"github.com/kyma-project/kyma/tests/components/api-gateway/gateway-tests/pkg/helpers"
	"github.com/kyma-project/kyma/tests/components/api-gateway/gateway-tests/pkg/jwt"
)

type oauthJWTOnePathScenario struct {
	*Scenario
}

func InitializeScenarioOAuth2JWTOnePath(ctx *godog.ScenarioContext) {
	mainScenario, err := CreateScenario(jwtAndOauthOnePathApiruleFile, "oauth2-jwt-one-path")
	if err != nil {
		t.Fatalf("could not initialize unsecure endpoint scenario err=%s", err)
	}

	scenario := oauthJWTOnePathScenario{mainScenario}

	ctx.Step(`^OAuth2JWT1Path: There is an deployment secured with both JWT and OAuth2 introspection on path /image$`, scenario.thereIsAnOauth2Endpoint)
	ctx.Step(`^OAuth2JWT1Path: Calling the "([^"]*)" endpoint without a token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutTokenShouldResultInStatusBeetween)
	ctx.Step(`^OAuth2JWT1Path: Calling the "([^"]*)" endpoint with a invalid token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithInvalidTokenShouldResultInStatusBeetween)
	ctx.Step(`^OAuth2JWT1Path: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithValidTokenShouldResultInStatusBeetween)
}

func (o *oauthJWTOnePathScenario) thereIsAnOauth2Endpoint() error {
	print("test")
	return helper.APIRuleWithRetries(batch.CreateResources, batch.UpdateResources, k8sClient, o.apiResource)
}

func (o *oauthJWTOnePathScenario) callingTheEndpointWithValidTokenShouldResultInStatusBeetween(path, tokenType string, lower, higher int) error {
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

 		headerVal := token.AccessToken

 		return helper.CallEndpointWithHeadersWithRetries(headerVal, "oauth2-access-token", fmt.Sprintf("%s%s", o.url, path), &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
	}
	return errors.New("should not happen")
}

func (o *oauthJWTOnePathScenario) callingTheEndpointWithInvalidTokenShouldResultInStatusBeetween(path string, lower, higher int) error {
	return helper.CallEndpointWithHeadersWithRetries(anyToken, authorizationHeaderName, fmt.Sprintf("%s%s", o.url, path), &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
}

func (o *oauthJWTOnePathScenario) callingTheEndpointWithoutTokenShouldResultInStatusBeetween(path string, lower, higher int) error {
	return helper.CallEndpointWithRetries(fmt.Sprintf("%s%s", o.url, path), &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
}
