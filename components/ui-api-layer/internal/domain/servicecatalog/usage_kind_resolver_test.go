package servicecatalog

import (
	"context"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/binding-usage-controller/pkg/client/clientset/versioned/fake"
	"github.com/kyma-project/kyma/components/binding-usage-controller/pkg/client/informers/externalversions"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	testingUtils "github.com/kyma-project/kyma/components/ui-api-layer/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	fakeDynamic "k8s.io/client-go/dynamic/fake"
)

func TestUsageKindResolver_ListUsageKinds(t *testing.T) {
	// GIVEN
	usageKindA := fixUsageKind("fix-A")
	usageKindB := fixUsageKind("fix-B")

	client := fake.NewSimpleClientset(usageKindA, usageKindB)
	informerFactory := externalversions.NewSharedInformerFactory(client, 0)

	informer := informerFactory.Servicecatalog().V1alpha1().UsageKinds().Informer()
	testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)
	svc := newUsageKindService(client.ServicecatalogV1alpha1(), nil, informer)
	resolver := newUsageKindResolver(svc)
	// WHEN
	resp, err := resolver.ListUsageKinds(context.Background(), nil, nil)
	require.NoError(t, err)

	// THEN
	assert.Equal(t, resp, fixUsageKindsGQL())
}

func TestUsageKindResolver_ListUsageKindResources_Empty(t *testing.T) {
	// GIVEN
	usageKindA := fixUsageKind("fix-A")

	dynamicClient := &fakeDynamic.FakeClientPool{}

	client := fake.NewSimpleClientset(usageKindA)
	informerFactory := externalversions.NewSharedInformerFactory(client, 0)

	informer := informerFactory.Servicecatalog().V1alpha1().UsageKinds().Informer()
	testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)
	svc := newUsageKindService(client.ServicecatalogV1alpha1(), dynamicClient, informer)
	resolver := newUsageKindResolver(svc)
	// WHEN
	resp, err := resolver.ListServiceUsageKindResources(context.Background(), "fix-A", "default")
	require.NoError(t, err)

	// THEN
	assert.Equal(t, []gqlschema.UsageKindResource{}, resp)
}

func fixUsageKindsGQL() []gqlschema.UsageKind {
	return []gqlschema.UsageKind{
		{
			Name:        "fix-A",
			DisplayName: fixUsageKindDisplayName(),
			Group:       fixUsageKindGroup(),
			Kind:        fixUsageKindKind(),
			Version:     fixUsageKindVersion(),
		},
		{
			Name:        "fix-B",
			DisplayName: fixUsageKindDisplayName(),
			Group:       fixUsageKindGroup(),
			Kind:        fixUsageKindKind(),
			Version:     fixUsageKindVersion(),
		},
	}
}
