package k8s

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	v1 "k8s.io/api/core/v1"
)

type namespaceConverter struct{}

func (c *namespaceConverter) toGQL(in *v1.Namespace) (*gqlschema.Namespace, error) {
	if in == nil {
		return nil, nil
	}

	return &gqlschema.Namespace{
		Name:   in.Name,
		Labels: in.Labels,
	}, nil
}

func (c *namespaceConverter) ToGQLs(in []*v1.Namespace) ([]gqlschema.Namespace, error) {
	var result []gqlschema.Namespace
	for _, u := range in {
		converted, err := c.toGQL(u)
		if err != nil {
			return nil, err
		}

		if converted != nil {
			result = append(result, *converted)
		}
	}
	return result, nil
}
