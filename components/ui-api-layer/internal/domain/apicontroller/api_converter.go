package apicontroller

import (
	"github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma.cx/v1alpha2"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
)

type apiConverter struct{}

func (ac *apiConverter) ToGQL(in *v1alpha2.Api) *gqlschema.API {
	if in == nil {
		return nil
	}

	var authenticationPolicies []gqlschema.AuthenticationPolicy
	for _, policy := range in.Spec.Authentication {

		authenticationPolicies = append(authenticationPolicies, gqlschema.AuthenticationPolicy{
			Type:    ac.parseAuthenticationPolicyType(&policy.Type),
			Issuer:  policy.Jwt.Issuer,
			JwksURI: policy.Jwt.JwksUri,
		})
	}

	return &gqlschema.API{
		Name:     in.Name,
		Hostname: in.Spec.Hostname,
		Service: gqlschema.Service{
			Name: in.Spec.Service.Name,
			Port: in.Spec.Service.Port,
		},
		AuthenticationPolicies: authenticationPolicies,
	}
}

func (c *apiConverter) ToGQLs(in []*v1alpha2.Api) []gqlschema.API {
	var result []gqlschema.API
	for _, item := range in {
		converted := c.ToGQL(item)

		if converted != nil {
			result = append(result, *converted)
		}
	}

	return result
}

func (c *apiConverter) parseAuthenticationPolicyType(in *v1alpha2.AuthenticationType) gqlschema.AuthenticationPolicyType {

	//Map everything to JWT type as for now, there is only one type
	return gqlschema.AuthenticationPolicyTypeJwt
}
