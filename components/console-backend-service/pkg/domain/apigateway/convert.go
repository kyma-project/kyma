package apigateway

import (
	"encoding/json"

	"github.com/kyma-incubator/api-gateway/api/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/domain/apigateway/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/graph/model"
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

func toGQLJSON(config *runtime.RawExtension) (map[string]interface{}, error) {
	result := map[string]interface{}{}

	if config != nil {
		err := json.Unmarshal(config.Raw, &result)
		if err != nil {
			return nil, errors.Wrapf(err, "while unmarshalling %s with Config `%s` to GQL JSON", pretty.APIRule, config)
		}
	}

	return result, nil
}

func (c *apiRuleConverter) ToGQL(in *v1alpha1.APIRule) (*model.APIRule, error) {
	if in == nil {
		return nil, nil
	}

	var rules []model.Rule

	for _, rule := range in.Spec.Rules {
		var gqlRule model.Rule
		var gqlAccessStrategies []*model.APIRuleConfig
		var gqlMutators []*model.APIRuleConfig

		for _, accessStrategy := range rule.AccessStrategies {
			qglAccessStrategyConfig, err := toGQLJSON(accessStrategy.Config)
			if err != nil {
				return nil, err
			}

			gqlAccessStrategies = append(gqlAccessStrategies, &model.APIRuleConfig{
				Name:   accessStrategy.Name,
				Config: qglAccessStrategyConfig,
			})
		}

		for _, mutator := range rule.Mutators {
			gqlMutatorConfig, err := toGQLJSON(mutator.Config)
			if err != nil {
				return nil, err
			}

			gqlMutators = append(gqlMutators, &model.APIRuleConfig{
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

	return &model.APIRule{
		Name: in.Name,
		Service: &model.APIRuleService{
			Host: *in.Spec.Service.Host,
			Name: *in.Spec.Service.Name,
			Port: int(*in.Spec.Service.Port),
		},
		Gateway: *in.Spec.Gateway,
		Rules:   rules,
		Status: &model.APIRuleStatuses{
			APIRuleStatus:        getResourceStatusOrNil(in.Status.APIRuleStatus),
			AccessRuleStatus:     getResourceStatusOrNil(in.Status.AccessRuleStatus),
			VirtualServiceStatus: getResourceStatusOrNil(in.Status.VirtualServiceStatus),
		},
	}, nil
}

func getResourceStatusOrNil(status *v1alpha1.APIRuleResourceStatus) *model.APIRuleStatus {
	if status == nil {
		return nil
	}
	return &model.APIRuleStatus{
		Code: string(status.Code),
		Desc: &status.Description,
	}
}

func (c *apiRuleConverter) ToGQLs(in []*v1alpha1.APIRule) ([]model.APIRule, error) {
	var result []model.APIRule
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

func fromGQLJSON(config map[string]interface{}) (*runtime.RawExtension, error ) {
	result := runtime.RawExtension{}
	var err error
	if config != nil {
		result.Raw, err = json.Marshal(config)
	}

	return &result, err
}

func (c *apiRuleConverter) ToApiRule(name string, namespace string, in model.APIRuleInput) (*v1alpha1.APIRule,error) {
	hostPort := uint32(in.ServicePort)
	rules,err := toRules(in.Rules)
	if err != nil {
		return nil, err
	}

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
			Rules:   rules,
		},
	}, nil
}

func toRules(ruleInputs []*model.RuleInput) ([]v1alpha1.Rule, error) {
	var rules []v1alpha1.Rule
	for _, rule := range ruleInputs {
		accessStrategies,err := toAccessStrategies(rule.AccessStrategies)
		if err != nil {
			return nil, err
		}

		mutators,err := toMutators(rule.Mutators)
		if err != nil {
			return nil, err
		}

		rules = append(rules, v1alpha1.Rule{
			Path:             rule.Path,
			Methods:          rule.Methods,
			AccessStrategies: accessStrategies,
			Mutators:         mutators,
		})
	}
	return rules, nil
}

func toAccessStrategies(accessStrategyInputs []*model.APIRuleConfigInput) ([]*alpha1.Authenticator, error) {
	var accessStrategies []*alpha1.Authenticator
	for _, accessStrategy := range accessStrategyInputs {
		config,err := fromGQLJSON(accessStrategy.Config)
		if err != nil {
			return nil, err
		}
		accessStrategies = append(accessStrategies, &alpha1.Authenticator{
			Handler: &alpha1.Handler{
				Name:   accessStrategy.Name,
				Config: config,
			},
		})
	}
	return accessStrategies, nil
}

func toMutators(mutatorInputs []*model.APIRuleConfigInput) ([]*alpha1.Mutator, error) {
	var mutators []*alpha1.Mutator
	for _, mutator := range mutatorInputs {
		config,err := fromGQLJSON(mutator.Config)
		if err != nil {
			return nil, err
		}
		mutators = append(mutators, &alpha1.Mutator{
			Handler: &alpha1.Handler{
				Name:   mutator.Name,
				Config: config,
			},
		})
	}
	return mutators, nil
}
