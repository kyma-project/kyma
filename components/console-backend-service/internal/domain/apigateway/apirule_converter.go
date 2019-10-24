package apigateway

import (
	"encoding/json"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kyma-incubator/api-gateway/api/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/apigateway/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/pkg/errors"
)

type apiRuleConverter struct{}

func toGQLJSON(config *runtime.RawExtension) (gqlschema.JSON, error) {
	jsonByte, err := json.Marshal(config)
	if err != nil {
		return nil, errors.Wrapf(err, "while marshalling %s with Config `%s`", pretty.APIRule, config)
	}

	jsonMap := make(map[string]interface{})
	err = json.Unmarshal(jsonByte, &jsonMap)
	if err != nil {
		return nil, errors.Wrapf(err, "while unmarshalling %s with Config `%s` to map", pretty.APIRule, config)
	}

	var result gqlschema.JSON
	err = result.UnmarshalGQL(jsonMap)
	if err != nil {
		return nil, errors.Wrapf(err, "while unmarshalling %s with Config `%s` to GQL JSON", pretty.APIRule, config)
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
		Status: &gqlschema.APIRuleStatuses{
			APIRuleStatus: gqlschema.APIRuleStatus{
				Code: string(in.Status.APIRuleStatus.Code),
				Desc: &in.Status.APIRuleStatus.Description,
			},
			AccessRuleStatus: gqlschema.APIRuleStatus{
				Code: string(in.Status.AccessRuleStatus.Code),
				Desc: &in.Status.AccessRuleStatus.Description,
			},
			VirtualServiceStatus: gqlschema.APIRuleStatus{
				Code: string(in.Status.VirtualServiceStatus.Code),
				Desc: &in.Status.VirtualServiceStatus.Description,
			},
		},
	}, nil
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

func (c *apiRuleConverter) ToApiRule(name string, namespace string, in gqlschema.APIRuleInput) *v1alpha1.APIRule {
	//TODO create APIRule obj

	return &v1alpha1.APIRule{}
}
