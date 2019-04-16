package assetstore_test

import (
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/assetstore"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/assetstore/listener"
	testingUtils "github.com/kyma-project/kyma/components/console-backend-service/internal/testing"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/dynamic/dynamicinformer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/tools/cache"
)

const (
	AssetNamespace = "AssetNamespace"
)

func TestAssetService_Find(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		assets := []runtime.Object{
			fixUnstructuredAsset(map[string]interface{}{
				"name":      "1",
				"namespace": AssetNamespace,
			}),
			fixUnstructuredAsset(map[string]interface{}{
				"name":      "2",
				"namespace": AssetNamespace,
			}),
			fixUnstructuredAsset(map[string]interface{}{
				"name":      "3",
				"namespace": AssetNamespace,
			}),
		}

		expected := fixAsset("1", nil)

		informer := fixAssetInformer(assets...)

		svc, err := assetstore.NewAssetService(informer)
		require.NoError(t, err)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		result, err := svc.Find(AssetNamespace, "1")
		require.NoError(t, err)

		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		informer := fixAssetInformer()

		svc, err := assetstore.NewAssetService(informer)
		require.NoError(t, err)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		result, err := svc.Find(AssetNamespace, "1")
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestAssetService_ListForDocsTopicByType(t *testing.T) {
	t.Run("Success with docsTopicName paramater", func(t *testing.T) {
		assets := []runtime.Object{
			fixUnstructuredAsset(map[string]interface{}{
				"name":      "1",
				"namespace": AssetNamespace,
				"labels": map[string]interface{}{
					assetstore.CmsDocsTopicLabel: "exampleDocsTopicA",
				},
			}),
			fixUnstructuredAsset(map[string]interface{}{
				"name":      "2",
				"namespace": AssetNamespace,
				"labels": map[string]interface{}{
					assetstore.CmsDocsTopicLabel: "exampleDocsTopicB",
				},
			}),
			fixUnstructuredAsset(map[string]interface{}{
				"name":      "3",
				"namespace": AssetNamespace,
				"labels": map[string]interface{}{
					assetstore.CmsDocsTopicLabel: "exampleDocsTopicC",
				},
			}),
		}

		expected := []*v1alpha2.Asset{
			fixAsset("1", map[string]string{
				assetstore.CmsDocsTopicLabel: "exampleDocsTopicA",
			}),
		}

		informer := fixAssetInformer(assets...)

		svc, err := assetstore.NewAssetService(informer)
		require.NoError(t, err)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		result, err := svc.ListForDocsTopicByType(AssetNamespace, "exampleDocsTopicA", nil)
		require.NoError(t, err)

		assert.Equal(t, expected, result)
	})

	t.Run("Success with whole paramaters", func(t *testing.T) {
		assets := []runtime.Object{
			fixUnstructuredAsset(map[string]interface{}{
				"name":      "1",
				"namespace": AssetNamespace,
				"labels": map[string]interface{}{
					assetstore.CmsDocsTopicLabel: "exampleDocsTopic",
					assetstore.CmsTypeLabel:      "markdown",
				},
			}),
			fixUnstructuredAsset(map[string]interface{}{
				"name":      "2",
				"namespace": AssetNamespace,
				"labels": map[string]interface{}{
					assetstore.CmsDocsTopicLabel: "exampleDocsTopic",
					assetstore.CmsTypeLabel:      "json",
				},
			}),
			fixUnstructuredAsset(map[string]interface{}{
				"name":      "3",
				"namespace": AssetNamespace,
				"labels": map[string]interface{}{
					assetstore.CmsDocsTopicLabel: "exampleDocsTopic",
					assetstore.CmsTypeLabel:      "yaml",
				},
			}),
		}

		expected := []*v1alpha2.Asset{
			fixAsset("1", map[string]string{
				assetstore.CmsDocsTopicLabel: "exampleDocsTopic",
				assetstore.CmsTypeLabel:      "markdown",
			}),
		}

		informer := fixAssetInformer(assets...)

		svc, err := assetstore.NewAssetService(informer)
		require.NoError(t, err)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		result, err := svc.ListForDocsTopicByType(AssetNamespace, "exampleDocsTopic", []string{"markdown"})
		require.NoError(t, err)

		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		informer := fixClusterAssetInformer()

		svc, err := assetstore.NewClusterAssetService(informer)
		require.NoError(t, err)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		result, err := svc.ListForDocsTopicByType("exampleDocsTopic", nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestAssetService_Subscribe(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		svc, err := assetstore.NewAssetService(fixAssetInformer())
		require.NoError(t, err)

		assetListener := listener.NewAsset(nil, nil, nil)
		svc.Subscribe(assetListener)
	})

	t.Run("Duplicated", func(t *testing.T) {
		svc, err := assetstore.NewAssetService(fixAssetInformer())
		require.NoError(t, err)

		assetListener := listener.NewAsset(nil, nil, nil)
		svc.Subscribe(assetListener)
		svc.Subscribe(assetListener)
	})

	t.Run("Multiple", func(t *testing.T) {
		svc, err := assetstore.NewAssetService(fixAssetInformer())
		require.NoError(t, err)

		assetListenerA := listener.NewAsset(nil, nil, nil)
		assetListenerB := listener.NewAsset(nil, nil, nil)

		svc.Subscribe(assetListenerA)
		svc.Subscribe(assetListenerB)
	})

	t.Run("Nil", func(t *testing.T) {
		svc, err := assetstore.NewAssetService(fixAssetInformer())
		require.NoError(t, err)

		svc.Subscribe(nil)
	})
}

func TestAssetService_Unsubscribe(t *testing.T) {
	t.Run("Existing", func(t *testing.T) {
		svc, err := assetstore.NewAssetService(fixAssetInformer())
		require.NoError(t, err)

		assetListener := listener.NewAsset(nil, nil, nil)
		svc.Subscribe(assetListener)
		svc.Unsubscribe(assetListener)
	})

	t.Run("Duplicated", func(t *testing.T) {
		svc, err := assetstore.NewAssetService(fixAssetInformer())
		require.NoError(t, err)

		assetListener := listener.NewAsset(nil, nil, nil)
		svc.Subscribe(assetListener)
		svc.Subscribe(assetListener)

		svc.Unsubscribe(assetListener)
	})

	t.Run("Multiple", func(t *testing.T) {
		svc, err := assetstore.NewAssetService(fixAssetInformer())
		require.NoError(t, err)

		assetListenerA := listener.NewAsset(nil, nil, nil)
		assetListenerB := listener.NewAsset(nil, nil, nil)

		svc.Subscribe(assetListenerA)
		svc.Subscribe(assetListenerB)

		svc.Unsubscribe(assetListenerA)
	})

	t.Run("Nil", func(t *testing.T) {
		svc, err := assetstore.NewAssetService(fixAssetInformer())
		require.NoError(t, err)

		svc.Unsubscribe(nil)
	})
}

func fixUnstructuredAsset(metadata map[string]interface{}) *unstructured.Unstructured {
	return testingUtils.NewUnstructured(v1alpha2.SchemeGroupVersion.String(), "Asset", metadata, nil, nil)
}

func fixAsset(name string, labels map[string]string) *v1alpha2.Asset {
	return &v1alpha2.Asset{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Asset",
			APIVersion: v1alpha2.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: AssetNamespace,
			Labels:    labels,
		},
	}
}

func fixAssetInformer(objects ...runtime.Object) cache.SharedIndexInformer {
	fakeClient := fake.NewSimpleDynamicClient(runtime.NewScheme(), objects...)
	informerFactory := dynamicinformer.NewDynamicSharedInformerFactory(fakeClient, 0)

	informer := informerFactory.ForResource(schema.GroupVersionResource{
		Version:  v1alpha2.SchemeGroupVersion.Version,
		Group:    v1alpha2.SchemeGroupVersion.Group,
		Resource: "assets",
	}).Informer()

	return informer
}
