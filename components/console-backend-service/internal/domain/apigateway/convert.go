package apigateway

import (
	"bytes"

	"github.com/kyma-incubator/api-gateway/api/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/apigateway/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
)

type ApiRuleUnstructuredExtractor struct{}

func (ext ApiRuleUnstructuredExtractor) Do(obj interface{}) (*v1alpha1.APIRule, error) {
	u, err := toUnstructured(obj)
	if err != nil {
		return nil, err
	}

	return fromUnstructured(u)
}

func toUnstructured(obj interface{}) (*unstructured.Unstructured, error) {
	if obj == nil {
		return nil, nil
	}

	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting resource %s %s to unstructured", pretty.APIRule, obj)
	}
	if len(u) == 0 {
		return nil, nil
	}

	return &unstructured.Unstructured{Object: u}, nil
}

func fromUnstructured(obj *unstructured.Unstructured) (*v1alpha1.APIRule, error) {
	if obj == nil {
		return nil, nil
	}
	var apiRule v1alpha1.APIRule
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &apiRule)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting unstructured to resource %s %s", pretty.APIRule, obj.Object)
	}

	return &apiRule, nil
}

type apiRuleConverter struct{}

func (c *apiRuleConverter) ToGQL(in *v1alpha1.APIRule) (*v1alpha1.APIRule, error) {
	return in, nil
}

func (c *apiRuleConverter) ToGQLs(in []*v1alpha1.APIRule) ([]*v1alpha1.APIRule, error) {
	return in, nil
}

func fromGQLJSON(config gqlschema.JSON) *runtime.RawExtension {
	result := runtime.RawExtension{}
	if config != nil {
		var buf bytes.Buffer
		config.MarshalGQL(&buf)
		result.Raw = buf.Bytes()
	}
	return &result
}

func (c *apiRuleConverter) ToApiRule(name string, namespace string, in gqlschema.APIRuleInput) *v1alpha1.APIRule {
	hostPort := uint32(in.ServicePort)
	return &v1alpha1.APIRule{
		TypeMeta: v1.TypeMeta{
			APIVersion: "gateway.kyma-project.io/v1alpha1",
			Kind:       "APIRule",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.APIRuleSpec{
			Service: &v1alpha1.Service{
				Name: &in.ServiceName,
				Port: &hostPort,
				Host: &in.Host,
			},
			Gateway: &in.Gateway,
			Rules:   toRules(in.Rules),
		},
	}
}

func toRules(ruleInputs []*gqlschema.RuleInput) []v1alpha1.Rule {
	var rules []v1alpha1.Rule
	for _, rule := range ruleInputs {

		rules = append(rules, v1alpha1.Rule{
			Path:             rule.Path,
			Methods:          rule.Methods,
			AccessStrategies: toAccessStrategies(rule.AccessStrategies),
			Mutators:         toMutators(rule.Mutators),
		})
	}
	return rules
}

func toAccessStrategies(accessStrategyInputs []*gqlschema.APIRuleConfigInput) []*alpha1.Authenticator {
	var accessStrategies []*alpha1.Authenticator
	for _, accessStrategy := range accessStrategyInputs {
		accessStrategies = append(accessStrategies, &alpha1.Authenticator{
			Handler: &alpha1.Handler{
				Name:   accessStrategy.Name,
				Config: fromGQLJSON(accessStrategy.Config),
			},
		})
	}
	return accessStrategies
}

func toMutators(mutatorInputs []*gqlschema.APIRuleConfigInput) []*alpha1.Mutator {
	var mutators []*alpha1.Mutator
	for _, mutator := range mutatorInputs {
		mutators = append(mutators, &alpha1.Mutator{
			Handler: &alpha1.Handler{
				Name:   mutator.Name,
				Config: fromGQLJSON(mutator.Config),
			},
		})
	}
	return mutators
}
