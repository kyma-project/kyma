package k8s

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	v1 "k8s.io/api/core/v1"
)

type namespaceConverter struct{}

func (c *namespaceConverter) ToGQL(in *v1.Namespace) (*gqlschema.Namespace, error) {
	if in == nil {
		return nil, nil
	}

	return &gqlschema.Namespace{
		Name:   in.Name,
		Labels: in.Labels,
		Status: string(in.Status.Phase),
	}, nil
}

func (c *namespaceConverter) ToGQLs(in []*v1.Namespace) ([]gqlschema.Namespace, error) {
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

func (c *namespaceConverter) ToGQLWithAdditionalData(in NamespaceWithAdditionalData) (*gqlschema.Namespace, error) {
	return &gqlschema.Namespace{
		Name:   in.namespace.Name,
		Labels: in.namespace.Labels,
		Status: string(in.namespace.Status.Phase),
		IsSystemNamespace: in.isSystemNamespace,
	}, nil
}

func (c *namespaceConverter) ToGQLsWithAdditionalData(in []NamespaceWithAdditionalData) ([]gqlschema.Namespace, error) {
	var result []gqlschema.Namespace
	for _, u := range in {
		converted, err := c.ToGQLWithAdditionalData(u)
		if err != nil {
			return nil, err
		}

		if converted != nil {
			result = append(result, *converted)
		}
	}
	return result, nil
}
