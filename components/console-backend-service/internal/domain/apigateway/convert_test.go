package apigateway

import (
	"testing"

	"github.com/kyma-incubator/api-gateway/api/v1alpha1"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApiRuleConverter_ToGQL(t *testing.T) {
	t.Run("APIRule definition given", func(t *testing.T) {
		name1 := "test-apiRule1"
		namespace := "test-namespace"
		hostname := "test-hostname1"
		serviceName := "test-service-name1"
		servicePort1 := uint32(8080)
		gateway1 := "gateway1"

		apiRule := fixTestApiRule(name1, namespace, hostname, serviceName, servicePort1, gateway1)

		expected := gqlschema.APIRule{
			Name: name1,
			Service: gqlschema.APIRuleService{
				Host: hostname,
				Name: serviceName,
				Port: int(servicePort1),
			},
			Gateway: gateway1,
			Rules: []gqlschema.Rule{
				{
					Path:    "*",
					Methods: []string{"GET"},
					AccessStrategies: []gqlschema.APIRuleConfig{
						{
							Name: "allow",
						},
					},
				},
			},
		}

		converter := apiRuleConverter{}
		result, err := converter.ToGQL(apiRule)

		require.NoError(t, err)
		assert.Equal(t, expected.Name, result.Name)
		assert.Equal(t, expected.Gateway, result.Gateway)
		assert.Equal(t, expected.Service, result.Service)
		assert.Equal(t, expected.Rules[0].Path, result.Rules[0].Path)
		assert.Equal(t, expected.Rules[0].Methods, result.Rules[0].Methods)
		assert.Equal(t, expected.Rules[0].AccessStrategies[0].Name, result.Rules[0].AccessStrategies[0].Name)
	})

	t.Run("Nil given", func(t *testing.T) {
		converter := apiRuleConverter{}
		result, err := converter.ToGQL(nil)

		require.NoError(t, err)
		require.Nil(t, result)
	})
}

func TestApiConverter_ToGQLs(t *testing.T) {

	name1 := "test-apiRule1"
	namespace := "test-namespace"
	hostname := "test-hostname1"
	serviceName := "test-service-name1"
	servicePort1 := uint32(8080)
	gateway1 := "gateway1"

	name2 := "test-apiRule2"
	servicePort2 := uint32(8080)
	gateway2 := "gateway2"

	t.Run("An array of APIRules given", func(t *testing.T) {
		apiRules := []*v1alpha1.APIRule{
			fixTestApiRule(name1, namespace, hostname, serviceName, servicePort1, gateway1),
			fixTestApiRule(name2, "different-namespace", hostname, serviceName, servicePort2, gateway2),
		}

		converter := apiRuleConverter{}
		result, err := converter.ToGQLs(apiRules)

		require.NoError(t, err)
		assert.Equal(t, len(apiRules), len(result))
		assert.NotEqual(t, apiRules[0].Name, apiRules[1].Name)
	})

	t.Run("An array of APIRules with nil given", func(t *testing.T) {
		apiRules := []*v1alpha1.APIRule{
			fixTestApiRule(name1, namespace, hostname, serviceName, servicePort1, gateway1),
			nil,
		}

		converter := apiRuleConverter{}
		result, err := converter.ToGQLs(apiRules)

		require.NoError(t, err)
		assert.Equal(t, len(apiRules)-1, len(result))
	})

	t.Run("An empty array given", func(t *testing.T) {
		apiRules := []*v1alpha1.APIRule{}

		converter := apiRuleConverter{}
		result, err := converter.ToGQLs(apiRules)

		require.NoError(t, err)
		assert.Empty(t, result)
	})
}

func TestApiRuleConverter_ToApiRule(t *testing.T) {
	name1 := "test-apiRule1"
	namespace := "test-namespace"
	hostname := "test-hostname1"
	serviceName := "test-service-name1"
	servicePort1 := uint32(8080)
	gateway1 := "gateway1"

	t.Run("Success", func(t *testing.T) {
		expected := fixTestApiRule(name1, namespace, hostname, serviceName, servicePort1, gateway1)

		converter := apiRuleConverter{}
		apiRuleInput := fixApiRuleInput(hostname, serviceName, int(servicePort1), gateway1)
		result := converter.ToApiRule(name, namespace, apiRuleInput)

		assert.Equal(t, expected.Name, result.Name)
		assert.Equal(t, expected.Spec.Gateway, result.Spec.Gateway)
		assert.Equal(t, expected.Spec.Service, result.Spec.Service)
		assert.Equal(t, expected.Spec.Rules[0].Path, result.Spec.Rules[0].Path)
		assert.Equal(t, expected.Spec.Rules[0].Methods, result.Spec.Rules[0].Methods)
		assert.Equal(t, expected.Spec.Rules[0].AccessStrategies[0].Name, result.Spec.Rules[0].AccessStrategies[0].Name)
	})
}

func fixApiRuleInput(hostname string, serviceName string, servicePort int, gateway string) gqlschema.APIRuleInput {
	return gqlschema.APIRuleInput{
		Host:        hostname,
		ServiceName: serviceName,
		ServicePort: servicePort,
		Gateway:     gateway,
		Rules: []gqlschema.RuleInput{
			{
				Path:    "*",
				Methods: []string{"GET"},
				AccessStrategies: []gqlschema.APIRuleConfigInput{
					{
						Name: "allow",
					},
				},
			},
		},
	}
}
