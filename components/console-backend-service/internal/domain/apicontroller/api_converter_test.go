package apicontroller

import (
	"testing"

	"github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma-project.io/v1alpha2"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestApiConverter_ToGQL(t *testing.T) {
	t.Run("API definition given", func(t *testing.T) {
		api := fixApi("test")

		expected := &gqlschema.API{
			Name:     api.Name,
			Hostname: api.Spec.Hostname,
			Service: gqlschema.ApiService{
				Name: api.Spec.Service.Name,
				Port: api.Spec.Service.Port,
			},
			AuthenticationPolicies: []gqlschema.AuthenticationPolicy{
				{
					Type:    gqlschema.AuthenticationPolicyTypeJwt,
					Issuer:  api.Spec.Authentication[0].Jwt.Issuer,
					JwksURI: api.Spec.Authentication[0].Jwt.JwksUri,
				},
			},
		}

		converter := apiConverter{}
		result := converter.ToGQL(api)

		assert.Equal(t, expected, result)
	})

	t.Run("Nil given", func(t *testing.T) {
		converter := apiConverter{}
		result := converter.ToGQL(nil)

		require.Nil(t, result)
	})
}

func TestApiConverter_ToGQLs(t *testing.T) {
	t.Run("An array of APIs given", func(t *testing.T) {
		apis := []*v1alpha2.Api{
			fixApi("test-1"),
			fixApi("test-2"),
		}

		converter := apiConverter{}
		result := converter.ToGQLs(apis)

		assert.Equal(t, len(apis), len(result))
		assert.NotEqual(t, apis[0].Name, apis[1].Name)
	})

	t.Run("An array of APIs with nil given", func(t *testing.T) {
		apis := []*v1alpha2.Api{
			fixApi("test-1"),
			nil,
		}

		converter := apiConverter{}
		result := converter.ToGQLs(apis)

		assert.Equal(t, len(apis)-1, len(result))
	})

	t.Run("An empty array given", func(t *testing.T) {
		apis := []*v1alpha2.Api{}

		converter := apiConverter{}
		result := converter.ToGQLs(apis)

		assert.Empty(t, result)
	})
}

func TestApiConverter_ToApi(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		expected := fixApi("test")

		converter := apiConverter{}
		apiInput := fixApiInput()
		result := converter.ToApi("test", "test", apiInput)

		assert.Equal(t, expected, result)
	})
}

func fixApi(name string) *v1alpha2.Api {
	return &v1alpha2.Api{
		TypeMeta: v1.TypeMeta{
			Kind:       "API",
			APIVersion: "authentication.kyma-project.io/v1alpha2",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: "test",
		},
		Spec: v1alpha2.ApiSpec{
			Hostname: "test-service.dev.kyma.cx",
			Service: v1alpha2.Service{
				Name: "test-service",
				Port: 8080,
			},
			Authentication: []v1alpha2.AuthenticationRule{
				{
					Type: v1alpha2.JwtType,
					Jwt: v1alpha2.JwtAuthentication{
						Issuer:  "sample-issuer",
						JwksUri: "http://sample-issuer/keys",
					},
				},
			},
		},
	}
}

func fixApiInput() gqlschema.APIInput {
	return gqlschema.APIInput{
		Hostname:    "test-service.dev.kyma.cx",
		ServiceName: "test-service",
		ServicePort: 8080,
		JwksURI:     "http://sample-issuer/keys",
		Issuer:      "sample-issuer",
		DisableIstioAuthPolicyMTLS: nil,
		AuthenticationEnabled:      nil,
	}
}
