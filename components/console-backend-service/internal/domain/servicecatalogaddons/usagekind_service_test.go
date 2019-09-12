package servicecatalogaddons_test

import (
	"testing"
	"time"

	"github.com/kyma-project/helm-broker/pkg/apis/addons/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
	testingUtils "github.com/kyma-project/kyma/components/console-backend-service/internal/testing"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/dynamic/dynamicinformer"
	sbu "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	dynamicFake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/tools/cache"
)

func TestUsageKindService_List_Success(t *testing.T) {
	// GIVEN
	usageKindA := fixUsageKind("fix-A")
	usageKindB := fixUsageKind("fix-B")

	client, err := newDynamicClient(usageKindA, usageKindB)
	require.NoError(t, err)
	informer := newUkFakeInformer(client)

	svc := servicecatalogaddons.NewUsageKindService(client, informer)
	testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

	// WHEN
	result, err := svc.List(pager.PagingParams{})
	require.NoError(t, err)

	// THEN
	assert.Equal(t, result, fixUsageKindsList())
}

func TestUsageKindService_List_Empty(t *testing.T) {
	// GIVEN
	scheme := runtime.NewScheme()
	client := dynamicFake.NewSimpleDynamicClient(scheme)
	informer := newUkFakeInformer(client)
	svc := servicecatalogaddons.NewUsageKindService(client, informer)
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
	dynamicClient := dynamicFake.NewSimpleDynamicClient(scheme, existingFunction)
	informer := newUkFakeInformer(dynamicClient)

	svc := servicecatalogaddons.NewUsageKindService(dynamicClient, informer)
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

func newSbuFakeInformer(dynamic dynamic.Interface) cache.SharedIndexInformer {
	return dynamicinformer.NewDynamicSharedInformerFactory(dynamic, 10).ForResource(bindingUsageGVR).Informer()
}

func newUkFakeInformer(dynamic dynamic.Interface) cache.SharedIndexInformer {
	return dynamicinformer.NewDynamicSharedInformerFactory(dynamic, 10).ForResource(usageKindsGVR).Informer()
}

func newDynamicClient(objects ...runtime.Object) (*dynamicFake.FakeDynamicClient, error) {
	scheme := runtime.NewScheme()
	err := v1alpha1.AddToScheme(scheme)
	if err != nil {
		return &dynamicFake.FakeDynamicClient{}, err
	}
	err = sbu.AddToScheme(scheme)
	if err != nil {
		return &dynamicFake.FakeDynamicClient{}, err
	}

	result := make([]runtime.Object, len(objects))
	for i, obj := range objects {
		converted, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
		result[i] = &unstructured.Unstructured{Object: converted}
	}
	return dynamicFake.NewSimpleDynamicClient(scheme, result...), nil
}
