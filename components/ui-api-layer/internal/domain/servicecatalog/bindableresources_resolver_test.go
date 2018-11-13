package servicecatalog

import (
	"context"
	"testing"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog/automock"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/magiconair/properties/assert"
	"github.com/stretchr/testify/require"
)

func TestBindableResourcesResolver_ListBindableResources(t *testing.T) {
	// GIVEN
	svc := automock.NewBindableResourcesLister()
	svc.On("ListResources", fixUsageKindResourceNamespace()).
		Return(fixBindableResourcesOutputItems(), nil).
		Once()
	defer svc.AssertExpectations(t)
	resolver := newBindableResourcesResolver(svc)

	// WHEN
	result, err := resolver.ListBindableResources(context.Background(), fixUsageKindResourceNamespace())

	// THEN
	require.NoError(t, err)
	assert.Equal(t, fixBindableResourcesOutputItems(), result)
}

func fixBindableResourcesOutputItems() []gqlschema.BindableResourcesOutputItem {
	return []gqlschema.BindableResourcesOutputItem{
		{
			Kind:        "deployment",
			DisplayName: "Deployments",
			Resources:   []gqlschema.UsageKindResource{},
		},
	}
}
