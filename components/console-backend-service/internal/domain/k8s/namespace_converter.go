package k8s

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/types"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
)

type namespaceConverter struct{}

func (c *namespaceConverter) ToGQL(in types.NamespaceWithAdditionalData) (*gqlschema.Namespace, error) {
	return &gqlschema.Namespace{
		Name:              in.Namespace.Name,
		Labels:            in.Namespace.Labels,
		Status:            string(in.Namespace.Status.Phase),
		IsSystemNamespace: in.IsSystemNamespace,
	}, nil
}

func (c *namespaceConverter) ToGQLs(in []types.NamespaceWithAdditionalData) ([]gqlschema.Namespace, error) {
	var result []gqlschema.Namespace
	for _, u := range in {
		converted, err := c.ToGQL(u)
		if err != nil {
			return nil, err
		}

		if converted != nil {
			result = append(result, *converted)
		}
	}
	return result, nil
}
