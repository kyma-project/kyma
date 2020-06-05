package k8s

import (
	"encoding/json"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	v1 "k8s.io/api/authorization/v1"
)

type selfSubjectRulesConverter struct {
}

func (c *selfSubjectRulesConverter) ToGQL(in *v1.SelfSubjectRulesReview) ([]gqlschema.ResourceRule, error) {
	if in == nil {
		return nil, nil
	}

	result := make([]gqlschema.ResourceRule, len(in.Status.ResourceRules))
	for i, resourceRule := range in.Status.ResourceRules {
		result[i] = gqlschema.ResourceRule{
			Verbs:     resourceRule.Verbs,
			APIGroups: resourceRule.APIGroups,
			Resources: resourceRule.Resources,
		}
	}

	return result, nil
}

func (c *selfSubjectRulesConverter) ToBytes(in *v1.SelfSubjectRulesReview) ([]byte, error) {

	return json.Marshal(in)
}
