package servicecatalog

import (
	"testing"

	"github.com/kyma-project/kyma/components/binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/magiconair/properties/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestUsageKindConverter_ToGQL(t *testing.T) {
	// GIVEN
	conv := usageKindConverter{}
	name := "harry"

	// WHEN
	result := conv.ToGQL(fixUsageKind(name))

	// THEN
	assert.Equal(t, result, fixGQLBindingTarget(name))
}

func fixUsageKind(name string) *v1alpha1.UsageKind {
	return &v1alpha1.UsageKind{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
		Spec: v1alpha1.UsageKindSpec{
			DisplayName: fixUsageKindDisplayName(),
			Resource: &v1alpha1.ResourceReference{
				Group:   fixUsageKindGroup(),
				Kind:    fixUsageKindKind(),
				Version: fixUsageKindVersion(),
			},
		},
	}
}

func fixGQLBindingTarget(name string) *gqlschema.UsageKind {
	return &gqlschema.UsageKind{
		Name:        name,
		Group:       fixUsageKindGroup(),
		Kind:        fixUsageKindKind(),
		Version:     fixUsageKindVersion(),
		DisplayName: fixUsageKindDisplayName(),
	}
}

func fixUsageKindGroup() string {
	return "kubeless.io"
}

func fixUsageKindKind() string {
	return "function"
}

func fixUsageKindVersion() string {
	return "v1beta1"
}

func fixUsageKindDisplayName() string {
	return "target"
}
