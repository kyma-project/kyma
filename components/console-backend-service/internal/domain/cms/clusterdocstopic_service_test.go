package cms_test

import (
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/cms"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/cms/listener"
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

func TestClusterDocsTopicService_Find(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		clusterDocsTopics := []runtime.Object{
			fixUnstructuredClusterDocsTopic(map[string]interface{}{
				"name": "exampleClassA",
			}),
			fixUnstructuredClusterDocsTopic(map[string]interface{}{
				"name": "exampleClassB",
			}),
			fixUnstructuredClusterDocsTopic(map[string]interface{}{
				"name": "exampleClassC",
			}),
		}
		expected := fixClusterDocsTopic("exampleClassA", nil)

		informer := fixClusterDocsTopicInformer(clusterDocsTopics...)

		svc, err := cms.NewClusterDocsTopicService(informer)
		require.NoError(t, err)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		result, err := svc.Find("exampleClassA")
		require.NoError(t, err)

		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		informer := fixClusterDocsTopicInformer()

		svc, err := cms.NewClusterDocsTopicService(informer)
		require.NoError(t, err)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		clusterDocsTopic, err := svc.Find("exampleClass")
		require.NoError(t, err)
		assert.Nil(t, clusterDocsTopic)
	})
}

func TestClusterDocsTopicService_List(t *testing.T) {
	t.Run("Success without parameters", func(t *testing.T) {
		clusterDocsTopics := []runtime.Object{
			fixUnstructuredClusterDocsTopic(map[string]interface{}{
				"name": "1",
				"labels": map[string]interface{}{
					cms.OrderLabel: "1",
				},
			}),
			fixUnstructuredClusterDocsTopic(map[string]interface{}{
				"name": "2",
				"labels": map[string]interface{}{
					cms.OrderLabel: "2",
				},
			}),
			fixUnstructuredClusterDocsTopic(map[string]interface{}{
				"name": "3",
				"labels": map[string]interface{}{
					cms.OrderLabel: "3",
				},
			}),
		}
		expected := []*v1alpha1.ClusterDocsTopic{
			fixClusterDocsTopic("1", map[string]string{
				cms.OrderLabel: "1",
			}),
			fixClusterDocsTopic("2", map[string]string{
				cms.OrderLabel: "2",
			}),
			fixClusterDocsTopic("3", map[string]string{
				cms.OrderLabel: "3",
			}),
		}

		informer := fixClusterDocsTopicInformer(clusterDocsTopics...)

		svc, err := cms.NewClusterDocsTopicService(informer)
		require.NoError(t, err)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		result, err := svc.List(nil, nil)
		require.NoError(t, err)

		assert.Equal(t, expected, result)
	})

	t.Run("Success with all parameters", func(t *testing.T) {
		viewContext := "viewContext"
		groupName := "groupName"
		clusterDocsTopics := []runtime.Object{
			fixUnstructuredClusterDocsTopic(map[string]interface{}{
				"name": "1",
				"labels": map[string]interface{}{
					cms.ViewContextLabel: viewContext,
					cms.GroupNameLabel:   groupName,
				},
			}),
			fixUnstructuredClusterDocsTopic(map[string]interface{}{
				"name": "2",
			}),
			fixUnstructuredClusterDocsTopic(map[string]interface{}{
				"name": "3",
			}),
		}
		expected := []*v1alpha1.ClusterDocsTopic{
			fixClusterDocsTopic("1", map[string]string{
				cms.ViewContextLabel: viewContext,
				cms.GroupNameLabel:   groupName,
			}),
		}

		informer := fixClusterDocsTopicInformer(clusterDocsTopics...)

		svc, err := cms.NewClusterDocsTopicService(informer)
		require.NoError(t, err)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		result, err := svc.List(&viewContext, &groupName)
		require.NoError(t, err)

		assert.Equal(t, expected, result)
	})

	t.Run("Success with viewContext parameter", func(t *testing.T) {
		viewContext := "viewContext"
		clusterDocsTopics := []runtime.Object{
			fixUnstructuredClusterDocsTopic(map[string]interface{}{
				"name": "1",
				"labels": map[string]interface{}{
					cms.ViewContextLabel: viewContext,
				},
			}),
			fixUnstructuredClusterDocsTopic(map[string]interface{}{
				"name": "2",
			}),
			fixUnstructuredClusterDocsTopic(map[string]interface{}{
				"name": "3",
			}),
		}
		expected := []*v1alpha1.ClusterDocsTopic{
			fixClusterDocsTopic("1", map[string]string{
				cms.ViewContextLabel: viewContext,
			}),
		}

		informer := fixClusterDocsTopicInformer(clusterDocsTopics...)

		svc, err := cms.NewClusterDocsTopicService(informer)
		require.NoError(t, err)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		result, err := svc.List(&viewContext, nil)
		require.NoError(t, err)

		assert.Equal(t, expected, result)
	})

	t.Run("Success with groupName parameter", func(t *testing.T) {
		groupName := "groupName"
		clusterDocsTopics := []runtime.Object{
			fixUnstructuredClusterDocsTopic(map[string]interface{}{
				"name": "1",
				"labels": map[string]interface{}{
					cms.GroupNameLabel: groupName,
				},
			}),
			fixUnstructuredClusterDocsTopic(map[string]interface{}{
				"name": "2",
			}),
			fixUnstructuredClusterDocsTopic(map[string]interface{}{
				"name": "3",
			}),
		}
		expected := []*v1alpha1.ClusterDocsTopic{
			fixClusterDocsTopic("1", map[string]string{
				cms.GroupNameLabel: groupName,
			}),
		}

		informer := fixClusterDocsTopicInformer(clusterDocsTopics...)

		svc, err := cms.NewClusterDocsTopicService(informer)
		require.NoError(t, err)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		result, err := svc.List(nil, &groupName)
		require.NoError(t, err)

		assert.Equal(t, expected, result)
	})

	t.Run("Success with sorting function", func(t *testing.T) {
		clusterDocsTopics := []runtime.Object{
			fixUnstructuredClusterDocsTopic(map[string]interface{}{
				"name": "2",
				"labels": map[string]interface{}{
					cms.OrderLabel: "2",
				},
			}),
			fixUnstructuredClusterDocsTopic(map[string]interface{}{
				"name": "4",
			}),
			fixUnstructuredClusterDocsTopic(map[string]interface{}{
				"name": "3",
				"labels": map[string]interface{}{
					cms.OrderLabel: "3",
				},
			}),
			fixUnstructuredClusterDocsTopic(map[string]interface{}{
				"name": "1",
				"labels": map[string]interface{}{
					cms.OrderLabel: "1",
				},
			}),
		}
		expected := []*v1alpha1.ClusterDocsTopic{
			fixClusterDocsTopic("1", map[string]string{
				cms.OrderLabel: "1",
			}),
			fixClusterDocsTopic("2", map[string]string{
				cms.OrderLabel: "2",
			}),
			fixClusterDocsTopic("3", map[string]string{
				cms.OrderLabel: "3",
			}),
			fixClusterDocsTopic("4", nil),
		}

		informer := fixClusterDocsTopicInformer(clusterDocsTopics...)

		svc, err := cms.NewClusterDocsTopicService(informer)
		require.NoError(t, err)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		result, err := svc.List(nil, nil)
		require.NoError(t, err)

		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		informer := fixClusterDocsTopicInformer()

		svc, err := cms.NewClusterDocsTopicService(informer)
		require.NoError(t, err)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		clusterDocsTopics, err := svc.List(nil, nil)
		require.NoError(t, err)
		assert.Nil(t, clusterDocsTopics)
	})
}

func TestClusterDocsTopicService_Subscribe(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		svc, err := cms.NewClusterDocsTopicService(fixClusterDocsTopicInformer())
		require.NoError(t, err)

		clusterDocsTopicListener := listener.NewClusterDocsTopic(nil, nil, nil)
		svc.Subscribe(clusterDocsTopicListener)
	})

	t.Run("Duplicated", func(t *testing.T) {
		svc, err := cms.NewClusterDocsTopicService(fixClusterDocsTopicInformer())
		require.NoError(t, err)

		clusterDocsTopicListener := listener.NewClusterDocsTopic(nil, nil, nil)
		svc.Subscribe(clusterDocsTopicListener)
		svc.Subscribe(clusterDocsTopicListener)
	})

	t.Run("Multiple", func(t *testing.T) {
		svc, err := cms.NewClusterDocsTopicService(fixClusterDocsTopicInformer())
		require.NoError(t, err)

		clusterDocsTopicListenerA := listener.NewClusterDocsTopic(nil, nil, nil)
		clusterDocsTopicListenerB := listener.NewClusterDocsTopic(nil, nil, nil)

		svc.Subscribe(clusterDocsTopicListenerA)
		svc.Subscribe(clusterDocsTopicListenerB)
	})

	t.Run("Nil", func(t *testing.T) {
		svc, err := cms.NewClusterDocsTopicService(fixClusterDocsTopicInformer())
		require.NoError(t, err)

		svc.Subscribe(nil)
	})
}

func TestClusterDocsTopicService_Unsubscribe(t *testing.T) {
	t.Run("Existing", func(t *testing.T) {
		svc, err := cms.NewClusterDocsTopicService(fixClusterDocsTopicInformer())
		require.NoError(t, err)

		clusterDocsTopicListener := listener.NewClusterDocsTopic(nil, nil, nil)

		svc.Subscribe(clusterDocsTopicListener)
		svc.Unsubscribe(clusterDocsTopicListener)
	})

	t.Run("Duplicated", func(t *testing.T) {
		svc, err := cms.NewClusterDocsTopicService(fixClusterDocsTopicInformer())
		require.NoError(t, err)

		clusterDocsTopicListener := listener.NewClusterDocsTopic(nil, nil, nil)
		svc.Subscribe(clusterDocsTopicListener)
		svc.Subscribe(clusterDocsTopicListener)

		svc.Unsubscribe(clusterDocsTopicListener)
	})

	t.Run("Multiple", func(t *testing.T) {
		svc, err := cms.NewClusterDocsTopicService(fixClusterDocsTopicInformer())
		require.NoError(t, err)

		clusterDocsTopicListenerA := listener.NewClusterDocsTopic(nil, nil, nil)
		clusterDocsTopicListenerB := listener.NewClusterDocsTopic(nil, nil, nil)

		svc.Subscribe(clusterDocsTopicListenerA)
		svc.Subscribe(clusterDocsTopicListenerB)

		svc.Unsubscribe(clusterDocsTopicListenerA)
	})

	t.Run("Nil", func(t *testing.T) {
		svc, err := cms.NewClusterDocsTopicService(fixClusterDocsTopicInformer())
		require.NoError(t, err)

		svc.Unsubscribe(nil)
	})
}

func fixUnstructuredClusterDocsTopic(metadata map[string]interface{}) *unstructured.Unstructured {
	return testingUtils.NewUnstructured(v1alpha1.SchemeGroupVersion.String(), "ClusterDocsTopic", metadata, nil, nil)
}

func fixClusterDocsTopic(name string, labels map[string]string) *v1alpha1.ClusterDocsTopic {
	return &v1alpha1.ClusterDocsTopic{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterDocsTopic",
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
	}
}

func fixClusterDocsTopicInformer(objects ...runtime.Object) cache.SharedIndexInformer {
	fakeClient := fake.NewSimpleDynamicClient(runtime.NewScheme(), objects...)
	informerFactory := dynamicinformer.NewDynamicSharedInformerFactory(fakeClient, 0)

	informer := informerFactory.ForResource(schema.GroupVersionResource{
		Version:  v1alpha1.SchemeGroupVersion.Version,
		Group:    v1alpha1.SchemeGroupVersion.Group,
		Resource: "clusterdocstopics",
	}).Informer()

	return informer
}
