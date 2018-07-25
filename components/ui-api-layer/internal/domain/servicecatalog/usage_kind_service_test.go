package servicecatalog

import (
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/binding-usage-controller/pkg/client/clientset/versioned/fake"
	"github.com/kyma-project/kyma/components/binding-usage-controller/pkg/client/informers/externalversions"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"
	testingUtils "github.com/kyma-project/kyma/components/ui-api-layer/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	fakeDynamic "k8s.io/client-go/dynamic/fake"
)

func TestUsageKindService_List_Success(t *testing.T) {
	// GIVEN
	usageKindA := fixUsageKind("fix-A")
	usageKindB := fixUsageKind("fix-B")

	client := fake.NewSimpleClientset(usageKindA, usageKindB)
	informerFactory := externalversions.NewSharedInformerFactory(client, 0)

	informer := informerFactory.Servicecatalog().V1alpha1().UsageKinds().Informer()
	svc := newUsageKindService(client.ServicecatalogV1alpha1(), nil, informer)
	testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

	// WHEN
	result, err := svc.List(pager.PagingParams{})
	require.NoError(t, err)

	// THEN
	assert.Len(t, result, 2)

}

func TestUsageKindService_List_Empty(t *testing.T) {
	// GIVEN
	client := fake.NewSimpleClientset()
	informerFactory := externalversions.NewSharedInformerFactory(client, 0)

	informer := informerFactory.Servicecatalog().V1alpha1().UsageKinds().Informer()
	svc := newUsageKindService(client.ServicecatalogV1alpha1(), nil, informer)
	testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

	// WHEN
	result, err := svc.List(pager.PagingParams{})
	require.NoError(t, err)

	// THEN
	assert.Len(t, result, 0)
}

func TestUsageKindService_ListBindingResources_Success(t *testing.T) {
	// GIVEN
	usageKind := fixUsageKind("fix-A")

	dynamicOperations := &fakeDynamic.FakeClientPool{}

	client := fake.NewSimpleClientset(usageKind)
	informerFactory := externalversions.NewSharedInformerFactory(client, 0)
	informer := informerFactory.Servicecatalog().V1alpha1().UsageKinds().Informer()

	svc := newUsageKindService(client.ServicecatalogV1alpha1(), dynamicOperations, informer)
	testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

	// WHEN
	result, err := svc.ListUsageKindResources("fix-A", fixUsageKindResourceNamespace())
	require.NoError(t, err)

	// THEN
	assert.Len(t, result, 0)
}

func fixUsageKindResourceNamespace() string {
	return "space"
}
