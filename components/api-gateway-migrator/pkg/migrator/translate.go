package migrator

import (
	"encoding/json"
	"fmt"

	gatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	oldapi "github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma-project.io/v1alpha2"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

func translateToApiRule(oldApi *oldapi.Api) (*gatewayv1alpha1.APIRule, error) {

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

	gateway := "kyma-gateway.kyma-system.svc.cluster.local" //TODO: Configurable?
	newApi.Spec.Gateway = &gateway

	rules, err := configureRules(oldApi.Spec.Authentication)
	if err != nil {
		return nil, err
	}

	newApi.Spec.Rules = rules

	return &newApi, nil
}

//TODO: Improve!
func configureRules(oldApiRules []oldapi.AuthenticationRule) ([]gatewayv1alpha1.Rule, error) {
	res := []gatewayv1alpha1.Rule{}

	newRule := gatewayv1alpha1.Rule{}
	newRule.Path = "/.*"
	newRule.Methods = []string{"GET", "PUT", "POST", "DELETE"}
	newRule.Mutators = nil

	jwksUrls := []string{}
	trustedIssuers := []string{}

	for _, oldApiRule := range oldApiRules {
		jwksUrls = append(jwksUrls, oldApiRule.Jwt.JwksUri)
		trustedIssuers = append(trustedIssuers, oldApiRule.Jwt.Issuer)

		if oldApiRule.Jwt.TriggerRule != nil {

			additionalRules := []gatewayv1alpha1.Rule{}

			allowHandler := rulev1alpha1.Handler{
				Name: "allow",
			}

			as1 := rulev1alpha1.Authenticator{
				&allowHandler,
			}

			accessStrategies := []*rulev1alpha1.Authenticator{&as1}

			for _, ar := range oldApiRule.Jwt.TriggerRule.ExcludedPaths {
				additionalRule := gatewayv1alpha1.Rule{}
				additionalRule.Path = ar.Value //TODO: Fix
				additionalRule.Methods = []string{"GET", "PUT", "POST", "DELETE"}
				additionalRule.AccessStrategies = accessStrategies
				additionalRule.Mutators = nil
				additionalRules = append(additionalRules, additionalRule)
			}

			res = append(res, additionalRules...)
		}
	}

	jwtAuthenticator, err := createJWTAuthenticator(jwksUrls, trustedIssuers)
	if err != nil {
		return nil, err
	}

	newRule.AccessStrategies = []*rulev1alpha1.Authenticator{jwtAuthenticator}
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
		return nil, fmt.Errorf("Could not marshal JWT config: %v", err)
	}

	rawConfig := &runtime.RawExtension{
		Raw: jwtConfigJSON,
	}

	jwtHandler := rulev1alpha1.Handler{
		Name:   "jwt",
		Config: rawConfig,
	}

	return &rulev1alpha1.Authenticator{
		&jwtHandler,
	}, nil
}

// JwtConfig Config
type JwtConfig struct {
	//RequiredScope []string `json:"required_scope"`
	JwksURLs       []string `json:"jwks_urls"`
	TrustedIssuers []string `json:"trusted_issuers"`
}
