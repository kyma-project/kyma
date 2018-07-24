package apicontroller

import (
	"testing"

	"github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma.cx/v1alpha2"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestApiConverter_ToGQL(t *testing.T) {
	t.Run("API definition given", func(t *testing.T) {
		api := fixApi("test")

		expected := &gqlschema.API{
			Name:     api.Name,
			Hostname: api.Spec.Hostname,
			Service: gqlschema.Service{
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

func fixApi(name string) *v1alpha2.Api {
	return &v1alpha2.Api{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
		Spec: v1alpha2.ApiSpec{
			Hostname: "test-service.dev.kyma.cx",
			Service: v1alpha2.Service{
				Name: "test-service",
				Port: 8080,
			},
			Authentication: v1alpha2.Authentication{
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
