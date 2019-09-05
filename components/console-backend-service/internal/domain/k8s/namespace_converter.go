package k8s

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	v1 "k8s.io/api/core/v1"
)

type namespaceConverter struct{
	systemNamespaces []string
}

func newNamespaceConverter(systemNamespaces []string) *namespaceConverter {
	return &namespaceConverter{
		systemNamespaces: systemNamespaces,
	}
}

func (c *namespaceConverter) ToGQL(in *v1.Namespace) (*gqlschema.Namespace, error) {
	if in == nil {
		return nil, nil
	}

	isSystem := isSystemNamespace(*in, c.systemNamespaces)
	return &gqlschema.Namespace{
		Name:              in.Name,
		Labels:            in.Labels,
		Status:            string(in.Status.Phase),
		IsSystemNamespace: isSystem,
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

func isSystemNamespace(namespace v1.Namespace, sysNamespaces []string) bool {
	for _, sysNs := range sysNamespaces {
		if sysNs == namespace.Name {
			return true
		}
	}
	return false
}
