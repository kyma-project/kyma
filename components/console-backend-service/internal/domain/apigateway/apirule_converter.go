package apigateway

import (
	"github.com/kyma-incubator/api-gateway/api/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
)

type apiRuleConverter struct{}

func (c *apiRuleConverter) ToGQL(in *v1alpha1.APIRule) *gqlschema.APIRule {
	if in == nil {
		return nil
	}

	//TODO iterate over rules & create APIRule obj

	return &gqlschema.APIRule{}
}

func (c *apiRuleConverter) ToGQLs(in []*v1alpha1.APIRule) []gqlschema.APIRule {
	var result []gqlschema.APIRule
	for _, item := range in {
		converted := c.ToGQL(item)

		if converted != nil {
			result = append(result, *converted)
		}
	}

	return result
}

func (c *apiRuleConverter) ToApiRule(name string, namespace string, in gqlschema.APIRuleInput) *v1alpha1.APIRule {
	//TODO create APIRule obj

	return &v1alpha1.APIRule{}
}
