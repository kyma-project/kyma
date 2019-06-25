package apicontroller

import (
	"github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma-project.io/v1alpha2"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		Name:              in.Name,
		Hostname:          in.Spec.Hostname,
		CreationTimestamp: in.CreationTimestamp.Time,
		Service: gqlschema.ApiService{
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

func (c *apiConverter) ToV1Api(name string, namespace string, in gqlschema.APIInput, resourceVersion string) *v1alpha2.Api {

	return &v1alpha2.Api{
		TypeMeta: v1.TypeMeta{
			APIVersion: "authentication.kyma-project.io/v1alpha2",
			Kind:       "API",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			ResourceVersion: resourceVersion,
		},
		Spec: v1alpha2.ApiSpec{
			Service: v1alpha2.Service{
				Name: in.ServiceName,
				Port: in.ServicePort,
			},
			Hostname: in.Hostname,
			Authentication: []v1alpha2.AuthenticationRule{
				{
					Jwt: v1alpha2.JwtAuthentication{
						JwksUri: in.JwksURI,
						Issuer:  in.Issuer,
					},
					Type: v1alpha2.AuthenticationType("JWT"),
				},
			},
			DisableIstioAuthPolicyMTLS: in.DisableIstioAuthPolicyMTLS,
			AuthenticationEnabled:      in.AuthenticationEnabled,
		},
	}
}

func (c *apiConverter) parseAuthenticationPolicyType(in *v1alpha2.AuthenticationType) gqlschema.AuthenticationPolicyType {

	//Map everything to JWT type as for now, there is only one type
	return gqlschema.AuthenticationPolicyTypeJwt
}
