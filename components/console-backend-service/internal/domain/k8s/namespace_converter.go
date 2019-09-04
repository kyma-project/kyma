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
		Status: in.Status.Phase,
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

func (c *namespaceConverter) ToGQLWithPods(in NamespaceWithAdditionalData) (*gqlschema.Namespace, error) {
	podC := podConverter{}
	pods, _ := podC.ToGQLs(in.pods)

	return &gqlschema.Namespace{
		Name:   in.namespace.Name,
		Labels: in.namespace.Labels,
		Status: in.namespace.Status.Phase,
		IsSystemNamespace: in.isSystemNamespace,
		Pods: pods,
	}, nil
}

func (c *namespaceConverter) ToGQLsWithPods(in []NamespaceWithAdditionalData) ([]gqlschema.Namespace, error) {
	var result []gqlschema.Namespace
	for _, u := range in {
		converted, err := c.ToGQLWithPods(u)
		if err != nil {
			return nil, err
		}

		if converted != nil {
			result = append(result, *converted)
		}
	}
	return result, nil
}
