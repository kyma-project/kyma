package servicecatalog

import (
	"github.com/kyma-project/kyma/components/binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
)

type gqlUsageKindConverter interface {
	ToGQL(usageKind *v1alpha1.UsageKind) *gqlschema.UsageKind
	ToGQLs(usageKinds []*v1alpha1.UsageKind) []gqlschema.UsageKind
}

type usageKindConverter struct{}

func (c *usageKindConverter) ToGQL(usageKind *v1alpha1.UsageKind) *gqlschema.UsageKind {
	if usageKind == nil {
		return nil
	}

	return &gqlschema.UsageKind{
		Name:        usageKind.Name,
		Kind:        usageKind.Spec.Resource.Kind,
		Version:     usageKind.Spec.Resource.Version,
		Group:       usageKind.Spec.Resource.Group,
		DisplayName: usageKind.Spec.DisplayName,
	}
}

func (c *usageKindConverter) ToGQLs(usageKinds []*v1alpha1.UsageKind) []gqlschema.UsageKind {
	var result []gqlschema.UsageKind
	for _, item := range usageKinds {
		converted := c.ToGQL(item)
		if converted != nil {
			result = append(result, *converted)
		}
	}
	return result
}
