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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
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
	assert.Equal(t, result, fixUsageKindsList())
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
	assert.Empty(t, result)
}

func TestUsageKindService_ListResources_Empty(t *testing.T) {
	// There is no any test for non-empty response because of a bug in fake dynamic scClient List() method.
	// The bug is fixed in scClient-go version 1.12-rc.1 but this version is not compatible with service-catalog:
	// vendor/github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1/register.go:77:36: too many arguments in call to scheme.AddFieldLabelConversionFunc

	// GIVEN
	usageKind := fixUsageKind("fix-A")

	existingFunction := newUnstructured(usageKind.Spec.Resource.Version, usageKind.Spec.Resource.Kind, "test", "test")
	scheme := runtime.NewScheme()
	dynamicClient := fakeDynamic.NewSimpleDynamicClient(scheme, existingFunction)

	client := fake.NewSimpleClientset()
	informerFactory := externalversions.NewSharedInformerFactory(client, 0)

	informer := informerFactory.Servicecatalog().V1alpha1().UsageKinds().Informer()
	svc := newUsageKindService(client.ServicecatalogV1alpha1(), dynamicClient, informer)
	testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

	// WHEN
	result, err := svc.ListResources("test")
	require.NoError(t, err)

	// THEN
	assert.Empty(t, result)
}

func newUnstructured(apiVersion, kind, namespace, name string) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": apiVersion,
			"kind":       kind,
			"metadata": map[string]interface{}{
				"namespace": namespace,
				"name":      name,
			},
		},
	}
	return obj
}
