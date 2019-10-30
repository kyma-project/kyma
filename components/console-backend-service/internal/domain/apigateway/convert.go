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

func toGQLJSON(config *runtime.RawExtension) (gqlschema.JSON, error) {
	result := gqlschema.JSON{}

	if config != nil {
		err := result.UnmarshalGQL(string(config.Raw))
		if err != nil {
			return nil, errors.Wrapf(err, "while unmarshalling %s with Config `%s` to GQL JSON", pretty.APIRule, config)
		}
	}

	return result, nil
}

func (c *apiRuleConverter) ToGQL(in *v1alpha1.APIRule) (*gqlschema.APIRule, error) {
	if in == nil {
		return nil, nil
	}

	var rules []gqlschema.Rule

	for _, rule := range in.Spec.Rules {
		var gqlRule gqlschema.Rule
		var gqlAccessStrategies []gqlschema.APIRuleConfig
		var gqlMutators []gqlschema.APIRuleConfig

		for _, accessStrategy := range rule.AccessStrategies {
			qglAccessStrategyConfig, err := toGQLJSON(accessStrategy.Config)
			if err != nil {
				return nil, err
			}

			gqlAccessStrategies = append(gqlAccessStrategies, gqlschema.APIRuleConfig{
				Name:   accessStrategy.Name,
				Config: qglAccessStrategyConfig,
			})
		}

		for _, mutator := range rule.Mutators {
			gqlMutatorConfig, err := toGQLJSON(mutator.Config)
			if err != nil {
				return nil, err
			}

			gqlMutators = append(gqlMutators, gqlschema.APIRuleConfig{
				Name:   mutator.Name,
				Config: gqlMutatorConfig,
			})
		}

		gqlRule.Path = rule.Path
		gqlRule.Methods = rule.Methods
		gqlRule.AccessStrategies = gqlAccessStrategies
		gqlRule.Mutators = gqlMutators

		rules = append(rules, gqlRule)
	}

	return &gqlschema.APIRule{
		Name: in.Name,
		Service: gqlschema.APIRuleService{
			Host: *in.Spec.Service.Host,
			Name: *in.Spec.Service.Name,
			Port: int(*in.Spec.Service.Port),
		},
		Gateway: *in.Spec.Gateway,
		Rules:   rules,
		Status: gqlschema.APIRuleStatuses{
			APIRuleStatus:        getResourceStatusOrNil(in.Status.APIRuleStatus),
			AccessRuleStatus:     getResourceStatusOrNil(in.Status.AccessRuleStatus),
			VirtualServiceStatus: getResourceStatusOrNil(in.Status.VirtualServiceStatus),
		},
	}, nil
}

func getResourceStatusOrNil(status *v1alpha1.APIRuleResourceStatus) *gqlschema.APIRuleStatus {
	if status == nil {
		return nil
	}
	return &gqlschema.APIRuleStatus{
		Code: string(status.Code),
		Desc: &status.Description,
	}
}

func (c *apiRuleConverter) ToGQLs(in []*v1alpha1.APIRule) ([]gqlschema.APIRule, error) {
	var result []gqlschema.APIRule
	for _, item := range in {
		converted, err := c.ToGQL(item)
		if err != nil {
			return nil, err
		}

		if converted != nil {
			result = append(result, *converted)
		}
	}

	return result, nil
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

func toRules(ruleInputs []gqlschema.RuleInput) []v1alpha1.Rule {
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

func toAccessStrategies(accessStrategyInputs []gqlschema.APIRuleConfigInput) []*alpha1.Authenticator {
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

func toMutators(mutatorInputs []gqlschema.APIRuleConfigInput) []*alpha1.Mutator {
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
