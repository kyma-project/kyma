package servicecatalogaddons

import (
	"context"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	"github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned/fake"
	"github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/informers/externalversions"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalogaddons/automock"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"
	testingUtils "github.com/kyma-project/kyma/components/ui-api-layer/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsageKindResolver_ListUsageKinds(t *testing.T) {
	// GIVEN
	usageKindA := fixUsageKind("fix-A")
	usageKindB := fixUsageKind("fix-B")

	svc := automock.NewUsageKindServices()
	svc.On("List", pager.PagingParams{}).
		Return(fixUsageKindsList(), nil).
		Once()
	defer svc.AssertExpectations(t)

	client := fake.NewSimpleClientset(usageKindA, usageKindB)
	informerFactory := externalversions.NewSharedInformerFactory(client, 0)

	informer := informerFactory.Servicecatalog().V1alpha1().UsageKinds().Informer()
	testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)
	resolver := newUsageKindResolver(svc)

	// WHEN
	resp, err := resolver.ListUsageKinds(context.Background(), nil, nil)

	// THEN
	require.NoError(t, err)
	assert.Equal(t, fixUsageKindsListGQL(), resp)
}

func fixUsageKindsList() []*v1alpha1.UsageKind {
	return []*v1alpha1.UsageKind{
		fixUsageKind("fix-A"),
		fixUsageKind("fix-B"),
	}
}

func fixUsageKindsListGQL() []gqlschema.UsageKind {
	return []gqlschema.UsageKind{
		*fixUsageKindGQL("fix-A"),
		*fixUsageKindGQL("fix-B"),
	}
}
