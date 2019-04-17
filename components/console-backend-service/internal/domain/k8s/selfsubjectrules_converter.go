package k8s

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	v1 "k8s.io/api/authorization/v1"
)

type selfSubjectRulesConverter struct {
}

func (c *selfSubjectRulesConverter) ToGQL(in *v1.SelfSubjectRulesReview) (*gqlschema.SelfSubjectRules, error) {
	if in == nil {
		return nil, nil
	}

	resourceRulesSlice := make([]*gqlschema.ResourceRule, len(in.Status.ResourceRules))
	for i, resourceRule := range in.Status.ResourceRules {
		resourceRulesSlice[i] = &gqlschema.ResourceRule{
			Verbs:     resourceRule.Verbs,
			APIGroups: resourceRule.APIGroups,
			Resources: resourceRule.Resources,
		}
	}

	out := &gqlschema.SelfSubjectRules{
		ResourceRules: resourceRulesSlice,
	}
	return out, nil
}
