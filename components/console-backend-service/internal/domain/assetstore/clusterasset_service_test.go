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

func TestClusterAssetService_Find(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		clusterAssets := []runtime.Object{
			fixUnstructuredClusterAsset(map[string]interface{}{
				"name": "1",
			}),
			fixUnstructuredClusterAsset(map[string]interface{}{
				"name": "2",
			}),
			fixUnstructuredClusterAsset(map[string]interface{}{
				"name": "3",
			}),
		}

		expected := fixClusterAsset("1", nil)

		informer := fixClusterAssetInformer(clusterAssets...)

		svc, err := assetstore.NewClusterAssetService(informer)
		require.NoError(t, err)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		result, err := svc.Find("1")
		require.NoError(t, err)

		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		informer := fixClusterAssetInformer()

		svc, err := assetstore.NewClusterAssetService(informer)
		require.NoError(t, err)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		result, err := svc.Find("1")
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestClusterAssetService_ListForDocsTopicByType(t *testing.T) {
	t.Run("Success with docsTopicName paramater", func(t *testing.T) {
		clusterAssets := []runtime.Object{
			fixUnstructuredClusterAsset(map[string]interface{}{
				"name": "1",
				"labels": map[string]interface{}{
					assetstore.CmsDocsTopicLabel: "exampleDocsTopicA",
				},
			}),
			fixUnstructuredClusterAsset(map[string]interface{}{
				"name": "2",
				"labels": map[string]interface{}{
					assetstore.CmsDocsTopicLabel: "exampleDocsTopicB",
				},
			}),
			fixUnstructuredClusterAsset(map[string]interface{}{
				"name": "3",
				"labels": map[string]interface{}{
					assetstore.CmsDocsTopicLabel: "exampleDocsTopicC",
				},
			}),
		}

		expected := []*v1alpha2.ClusterAsset{
			fixClusterAsset("1", map[string]string{
				assetstore.CmsDocsTopicLabel: "exampleDocsTopicA",
			}),
		}

		informer := fixClusterAssetInformer(clusterAssets...)

		svc, err := assetstore.NewClusterAssetService(informer)
		require.NoError(t, err)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		result, err := svc.ListForDocsTopicByType("exampleDocsTopicA", nil)
		require.NoError(t, err)

		assert.Equal(t, expected, result)
	})

	t.Run("Success with whole paramaters", func(t *testing.T) {
		clusterAssets := []runtime.Object{
			fixUnstructuredClusterAsset(map[string]interface{}{
				"name": "1",
				"labels": map[string]interface{}{
					assetstore.CmsDocsTopicLabel: "exampleDocsTopic",
					assetstore.CmsTypeLabel:      "markdown",
				},
			}),
			fixUnstructuredClusterAsset(map[string]interface{}{
				"name": "2",
				"labels": map[string]interface{}{
					assetstore.CmsDocsTopicLabel: "exampleDocsTopic",
					assetstore.CmsTypeLabel:      "json",
				},
			}),
			fixUnstructuredClusterAsset(map[string]interface{}{
				"name": "3",
				"labels": map[string]interface{}{
					assetstore.CmsDocsTopicLabel: "exampleDocsTopic",
					assetstore.CmsTypeLabel:      "yaml",
				},
			}),
		}

		expected := []*v1alpha2.ClusterAsset{
			fixClusterAsset("1", map[string]string{
				assetstore.CmsDocsTopicLabel: "exampleDocsTopic",
				assetstore.CmsTypeLabel:      "markdown",
			}),
		}

		informer := fixClusterAssetInformer(clusterAssets...)

		svc, err := assetstore.NewClusterAssetService(informer)
		require.NoError(t, err)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		result, err := svc.ListForDocsTopicByType("exampleDocsTopic", []string{"markdown"})
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

func TestClusterAssetService_Subscribe(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		svc, err := assetstore.NewClusterAssetService(fixClusterAssetInformer())
		require.NoError(t, err)

		clusterAssetListener := listener.NewClusterAsset(nil, nil, nil)
		svc.Subscribe(clusterAssetListener)
	})

	t.Run("Duplicated", func(t *testing.T) {
		svc, err := assetstore.NewClusterAssetService(fixClusterAssetInformer())
		require.NoError(t, err)

		clusterAssetListener := listener.NewClusterAsset(nil, nil, nil)
		svc.Subscribe(clusterAssetListener)
		svc.Subscribe(clusterAssetListener)
	})

	t.Run("Multiple", func(t *testing.T) {
		svc, err := assetstore.NewClusterAssetService(fixClusterAssetInformer())
		require.NoError(t, err)

		clusterAssetListenerA := listener.NewClusterAsset(nil, nil, nil)
		clusterAssetListenerB := listener.NewClusterAsset(nil, nil, nil)

		svc.Subscribe(clusterAssetListenerA)
		svc.Subscribe(clusterAssetListenerB)
	})

	t.Run("Nil", func(t *testing.T) {
		svc, err := assetstore.NewClusterAssetService(fixClusterAssetInformer())
		require.NoError(t, err)

		svc.Subscribe(nil)
	})
}

func TestClusterAssetService_Unsubscribe(t *testing.T) {
	t.Run("Existing", func(t *testing.T) {
		svc, err := assetstore.NewClusterAssetService(fixClusterAssetInformer())
		require.NoError(t, err)

		clusterAssetListener := listener.NewClusterAsset(nil, nil, nil)
		svc.Subscribe(clusterAssetListener)
		svc.Unsubscribe(clusterAssetListener)
	})

	t.Run("Duplicated", func(t *testing.T) {
		svc, err := assetstore.NewClusterAssetService(fixClusterAssetInformer())
		require.NoError(t, err)

		clusterAssetListener := listener.NewClusterAsset(nil, nil, nil)
		svc.Subscribe(clusterAssetListener)
		svc.Subscribe(clusterAssetListener)

		svc.Unsubscribe(clusterAssetListener)
	})

	t.Run("Multiple", func(t *testing.T) {
		svc, err := assetstore.NewClusterAssetService(fixClusterAssetInformer())
		require.NoError(t, err)

		clusterAssetListenerA := listener.NewClusterAsset(nil, nil, nil)
		clusterAssetListenerB := listener.NewClusterAsset(nil, nil, nil)

		svc.Subscribe(clusterAssetListenerA)
		svc.Subscribe(clusterAssetListenerB)

		svc.Unsubscribe(clusterAssetListenerA)
	})

	t.Run("Nil", func(t *testing.T) {
		svc, err := assetstore.NewClusterAssetService(fixClusterAssetInformer())
		require.NoError(t, err)

		svc.Unsubscribe(nil)
	})
}

func fixUnstructuredClusterAsset(metadata map[string]interface{}) *unstructured.Unstructured {
	return testingUtils.NewUnstructured(v1alpha2.SchemeGroupVersion.String(), "ClusterAsset", metadata, nil, nil)
}

func fixClusterAsset(name string, labels map[string]string) *v1alpha2.ClusterAsset {
	return &v1alpha2.ClusterAsset{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterAsset",
			APIVersion: v1alpha2.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
	}
}

func fixClusterAssetInformer(objects ...runtime.Object) cache.SharedIndexInformer {
	fakeClient := fake.NewSimpleDynamicClient(runtime.NewScheme(), objects...)
	informerFactory := dynamicinformer.NewDynamicSharedInformerFactory(fakeClient, 0)

	informer := informerFactory.ForResource(schema.GroupVersionResource{
		Version:  v1alpha2.SchemeGroupVersion.Version,
		Group:    v1alpha2.SchemeGroupVersion.Group,
		Resource: "clusterassets",
	}).Informer()

	return informer
}
