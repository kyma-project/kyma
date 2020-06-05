package servicecatalogaddons_test

import (
	"context"
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBindableResourcesResolver_ListBindableResources(t *testing.T) {
	// GIVEN
	svc := automock.NewBindableResourcesLister()
	svc.On("ListResources", fixUsageKindResourceNamespace()).
		Return(fixBindableResourcesOutputItems(), nil).
		Once()
	defer svc.AssertExpectations(t)
	resolver := servicecatalogaddons.NewBindableResourcesResolver(svc)

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
