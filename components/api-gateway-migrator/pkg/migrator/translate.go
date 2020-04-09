package migrator

import (
	"encoding/json"
	"fmt"

	gatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	oldapi "github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma-project.io/v1alpha2"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

func translateToApiRule(oldApi *oldapi.Api, gateway string) (*gatewayv1alpha1.APIRule, error) {

	newApi := gatewayv1alpha1.APIRule{}
	newApi.Name = oldApi.Name
	newApi.Kind = "APIRule"
	newApi.APIVersion = "gateway.kyma-project.io/v1alpha1"
	newApi.Namespace = oldApi.Namespace

	newApi.Spec = gatewayv1alpha1.APIRuleSpec{}

	port := uint32(oldApi.Spec.Service.Port)
	isExternal := false

	newApi.Spec.Service = &gatewayv1alpha1.Service{
		Name:       &oldApi.Spec.Service.Name,
		Port:       &port,
		Host:       &oldApi.Spec.Hostname,
		IsExternal: &isExternal,
	}

	newApi.Spec.Gateway = &gateway

	rules, err := configureRules(oldApi.Spec.Authentication)
	if err != nil {
		return nil, err
	}

	newApi.Spec.Rules = rules

	return &newApi, nil
}

func configureRules(oldApiRules []oldapi.AuthenticationRule) ([]gatewayv1alpha1.Rule, error) {
	res := []gatewayv1alpha1.Rule{}

	newRule := gatewayv1alpha1.Rule{}
	newRule.Path = "/.*"
	newRule.Methods = []string{"GET", "PUT", "POST", "DELETE", "PATCH", "HEAD", "OPTIONS"}
	newRule.Mutators = nil

	if len(oldApiRules) > 0 {

		//Collect JWT params and set these to main newRule entry - this one will go through Oathkeeper.
		jwksUrls := []string{}
		trustedIssuers := []string{}

		for _, oldApiRule := range oldApiRules {
			jwksUrls = append(jwksUrls, oldApiRule.Jwt.JwksUri)
			trustedIssuers = append(trustedIssuers, oldApiRule.Jwt.Issuer)
		}

		jwtAuthenticator, err := createJWTAuthenticator(jwksUrls, trustedIssuers)
		if err != nil {
			return nil, err
		}
		newRule.AccessStrategies = []*rulev1alpha1.Authenticator{jwtAuthenticator}

		//Create subsequent rules for possible excludedPaths. Look only for the first Rule.
		//These must be the same for all the other, otherwise it's filtered out
		oldApiRule := oldApiRules[0]
		if oldApiRule.Jwt.TriggerRule != nil {

			rulesForExcludedPaths := []gatewayv1alpha1.Rule{}

			allowHandler := rulev1alpha1.Handler{
				Name: "allow",
			}

			as1 := rulev1alpha1.Authenticator{
				Handler: &allowHandler,
			}

			accessStrategies := []*rulev1alpha1.Authenticator{&as1}

			for _, ar := range oldApiRule.Jwt.TriggerRule.ExcludedPaths {
				ruleForExcludedPath := gatewayv1alpha1.Rule{}
				ruleForExcludedPath.Path = translatePath(fmt.Sprint(ar.ExprType), ar.Value)
				ruleForExcludedPath.Methods = []string{"GET", "PUT", "POST", "DELETE", "PATCH", "HEAD", "OPTIONS"}
				ruleForExcludedPath.AccessStrategies = accessStrategies
				ruleForExcludedPath.Mutators = nil
				rulesForExcludedPaths = append(rulesForExcludedPaths, ruleForExcludedPath)
			}

			res = append(res, rulesForExcludedPaths...)
		}
	} else {
		//When no rules is defined, old Api allowed for unauthenticated requests.
		allowHandler := rulev1alpha1.Handler{
			Name: "allow",
		}

		as1 := rulev1alpha1.Authenticator{
			Handler: &allowHandler,
		}

		newRule.AccessStrategies = []*rulev1alpha1.Authenticator{&as1}
	}

	res = append(res, newRule)
	return res, nil
}

func createJWTAuthenticator(jwksUrls []string, trustedIssuers []string) (*rulev1alpha1.Authenticator, error) {
	jwtConfig := &JwtConfig{
		JwksURLs:       jwksUrls,
		TrustedIssuers: trustedIssuers,
	}

	jwtConfigJSON, err := json.Marshal(jwtConfig)
	if err != nil {
		return nil, fmt.Errorf("could not marshal JWT config: %v", err)
	}

	rawConfig := &runtime.RawExtension{
		Raw: jwtConfigJSON,
	}

	jwtHandler := rulev1alpha1.Handler{
		Name:   "jwt",
		Config: rawConfig,
	}

	return &rulev1alpha1.Authenticator{
		Handler: &jwtHandler,
	}, nil
}

func translatePath(pathType, apiPath string) string {
	switch pathType {
	case fmt.Sprint(oldapi.ExactMatch), fmt.Sprint(oldapi.RegexMatch):
		return apiPath
	case fmt.Sprint(oldapi.PrefixMatch):
		return fmt.Sprintf("%s%s", apiPath, ".*")
	case fmt.Sprint(oldapi.SuffixMatch):
		return fmt.Sprintf("%s%s", ".*", apiPath)
	default:
		return apiPath
	}
}

// JwtConfig Config
type JwtConfig struct {
	//RequiredScope []string `json:"required_scope"`
	JwksURLs       []string `json:"jwks_urls"`
	TrustedIssuers []string `json:"trusted_issuers"`
}
