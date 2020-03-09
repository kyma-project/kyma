package servicecatalogaddons_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"

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

func TestUsageKindService_List(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		kubelessRef := fixKubelessFunctionResourceReference()
		usageKindA := fixUsageKind("fix-A", kubelessRef)
		usageKindB := fixUsageKind("fix-B", kubelessRef)
		usageKinds := []*sbu.UsageKind{
			usageKindA,
			usageKindB,
		}

		client, err := newDynamicClient(usageKindA, usageKindB)
		require.NoError(t, err)
		informer := newUkFakeInformer(client)

		svc := servicecatalogaddons.NewUsageKindService(client, informer)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		// WHEN
		result, err := svc.List(pager.PagingParams{})
		require.NoError(t, err)

		// THEN
		assert.Equal(t, usageKinds, result)
	})

	t.Run("Empty", func(t *testing.T) {
		// GIVEN
		client, err := newDynamicClient()
		require.NoError(t, err)
		informer := newUkFakeInformer(client)

		svc := servicecatalogaddons.NewUsageKindService(client, informer)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		// WHEN
		result, err := svc.List(pager.PagingParams{})
		require.NoError(t, err)

		// THEN
		assert.Empty(t, result)
	})
}

func TestUsageKindService_ListResources(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		kubelessRef := fixKubelessFunctionResourceReference()
		usageKind := fixUsageKind("fix-A", kubelessRef)

		apiVersion := fmt.Sprintf("%s/%s", usageKind.Spec.Resource.Group, usageKind.Spec.Resource.Version)
		existingFunction := newUnstructured(apiVersion, usageKind.Spec.Resource.Kind, "test", "test", []interface{}{})
		expected := []gqlschema.BindableResourcesOutputItem{
			{
				Kind:        usageKind.Name,
				DisplayName: usageKind.Spec.DisplayName,
				Resources: []gqlschema.UsageKindResource{
					{
						Name:      "test",
						Namespace: "test",
					},
				},
			},
		}

		client, err := newDynamicClient(usageKind, existingFunction)
		require.NoError(t, err)
		informer := newUkFakeInformer(client)

		svc := servicecatalogaddons.NewUsageKindService(client, informer)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		// WHEN
		result, err := svc.ListResources("test")
		require.NoError(t, err)

		// THEN
		assert.Equal(t, expected, result)
	})

	t.Run("Empty", func(t *testing.T) {
		// GIVEN
		client, err := newDynamicClient()
		require.NoError(t, err)
		informer := newUkFakeInformer(client)

		svc := servicecatalogaddons.NewUsageKindService(client, informer)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		// WHEN
		result, err := svc.ListResources("test")
		require.NoError(t, err)

		// THEN
		assert.Empty(t, result)
	})

	t.Run("omitResourceByOwnerRefs", func(t *testing.T) {
		// GIVEN
		kubelessRef := fixKubelessFunctionResourceReference()
		usageKindA := fixUsageKind("fix-A", kubelessRef)
		kubelessOwnerRef := []interface{}{
			map[string]interface{}{
				"apiVersion": fmt.Sprintf("%s/%s", kubelessRef.Group, kubelessRef.Version),
				"kind":       kubelessRef.Kind,
			},
		}

		deploymentRef := fixDeploymentResourceReference()
		usageKindB := fixUsageKind("fix-B", deploymentRef)

		apiVersion := fmt.Sprintf("%s/%s", usageKindA.Spec.Resource.Group, usageKindA.Spec.Resource.Version)
		existingFunction := newUnstructured(apiVersion, usageKindA.Spec.Resource.Kind, "test", "test-A", []interface{}{})
		apiVersion = fmt.Sprintf("%s/%s", usageKindB.Spec.Resource.Group, usageKindB.Spec.Resource.Version)
		existingDeploymentA := newUnstructured(apiVersion, usageKindB.Spec.Resource.Kind, "test", "test-B", kubelessOwnerRef)
		existingDeploymentB := newUnstructured(apiVersion, usageKindB.Spec.Resource.Kind, "test", "test-C", []interface{}{})

		client, err := newDynamicClient(usageKindA, usageKindB, existingFunction, existingDeploymentA, existingDeploymentB)
		require.NoError(t, err)
		informer := newUkFakeInformer(client)

		svc := servicecatalogaddons.NewUsageKindService(client, informer)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		// WHEN
		result, err := svc.ListResources("test")
		require.NoError(t, err)

		// THEN
		for _, item := range result {
			if item.Kind == usageKindA.Name {
				require.Equal(t, len(item.Resources), 1)
				require.Equal(t, item.Resources[0].Name, "test-A")
			}
			if item.Kind == usageKindB.Name {
				require.Equal(t, len(item.Resources), 1)
				require.Equal(t, item.Resources[0].Name, "test-C")
			}
		}
	})
}

func newUnstructured(apiVersion, kind, namespace, name string, ownerRefs []interface{}) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": apiVersion,
			"kind":       kind,
			"metadata": map[string]interface{}{
				"namespace":       namespace,
				"name":            name,
				"ownerReferences": ownerRefs,
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
