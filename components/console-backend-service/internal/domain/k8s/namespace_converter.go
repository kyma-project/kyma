package k8s

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	v1 "k8s.io/api/core/v1"
)

type namespaceConverter struct {
	systemNamespaces []string
}

func newNamespaceConverter(systemNamespaces []string) *namespaceConverter {
	return &namespaceConverter{
		systemNamespaces: systemNamespaces,
	}
}

func (c *namespaceConverter) ToGQL(in *v1.Namespace) *gqlschema.Namespace {
	if in == nil {
		return nil
	}

	labels := map[string]string{}
	if in.Labels != nil {
		labels = in.Labels
	}

	isSystem := isSystemNamespace(*in, c.systemNamespaces)
	return &gqlschema.Namespace{
		Name:              in.Name,
		Labels:            labels,
		Status:            string(in.Status.Phase),
		IsSystemNamespace: isSystem,
	}
}

func (c *namespaceConverter) ToListItemGQL(in *v1.Namespace) *gqlschema.NamespaceListItem {
	if in == nil {
		return nil
	}

	labels := map[string]string{}
	if in.Labels != nil {
		labels = in.Labels
	}

	isSystem := isSystemNamespace(*in, c.systemNamespaces)
	return &gqlschema.NamespaceListItem{
		Name:              in.Name,
		Labels:            labels,
		Status:            string(in.Status.Phase),
		IsSystemNamespace: isSystem,
	}
}

func (c *namespaceConverter) ToGQLs(in []*v1.Namespace) []*gqlschema.NamespaceListItem {
	var result []*gqlschema.NamespaceListItem
	for _, u := range in {
		converted := c.ToListItemGQL(u)

		if converted != nil {
			result = append(result, converted)
		}
	}
	return result
}

func isSystemNamespace(namespace v1.Namespace, sysNamespaces []string) bool {
	for _, sysNs := range sysNamespaces {
		if sysNs == namespace.Name {
			return true
		}
	}
	return false
}
