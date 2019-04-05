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

const (
	DocsTopicNamespace = "DocsTopicNamespace"
)

func TestDocsTopicService_Find(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		docsTopics := []runtime.Object{
			fixUnstructuredDocsTopic(map[string]interface{}{
				"name":      "exampleClassA",
				"namespace": DocsTopicNamespace,
			}),
			fixUnstructuredDocsTopic(map[string]interface{}{
				"name":      "exampleClassB",
				"namespace": DocsTopicNamespace,
			}),
			fixUnstructuredDocsTopic(map[string]interface{}{
				"name":      "exampleClassC",
				"namespace": DocsTopicNamespace,
			}),
		}
		expected := fixDocsTopic("exampleClassA", nil)

		informer := fixDocsTopicInformer(docsTopics...)

		svc, err := cms.NewDocsTopicService(informer)
		require.NoError(t, err)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		result, err := svc.Find(DocsTopicNamespace, "exampleClassA")
		require.NoError(t, err)

		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		informer := fixDocsTopicInformer()

		svc, err := cms.NewDocsTopicService(informer)
		require.NoError(t, err)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		docsTopics, err := svc.Find(DocsTopicNamespace, "exampleClass")
		require.NoError(t, err)
		assert.Nil(t, docsTopics)
	})
}

func TestDocsTopicService_Subscribe(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		svc, err := cms.NewDocsTopicService(fixDocsTopicInformer())
		require.NoError(t, err)

		docsTopicListener := listener.NewDocsTopic(nil, nil, nil)
		svc.Subscribe(docsTopicListener)
	})

	t.Run("Duplicated", func(t *testing.T) {
		svc, err := cms.NewDocsTopicService(fixDocsTopicInformer())
		require.NoError(t, err)

		docsTopicListener := listener.NewDocsTopic(nil, nil, nil)
		svc.Subscribe(docsTopicListener)
		svc.Subscribe(docsTopicListener)
	})

	t.Run("Multiple", func(t *testing.T) {
		svc, err := cms.NewDocsTopicService(fixDocsTopicInformer())
		require.NoError(t, err)

		docsTopicListenerA := listener.NewDocsTopic(nil, nil, nil)
		docsTopicListenerB := listener.NewDocsTopic(nil, nil, nil)

		svc.Subscribe(docsTopicListenerA)
		svc.Subscribe(docsTopicListenerB)
	})

	t.Run("Nil", func(t *testing.T) {
		svc, err := cms.NewDocsTopicService(fixDocsTopicInformer())
		require.NoError(t, err)

		svc.Subscribe(nil)
	})
}

func TestDocsTopicService_Unsubscribe(t *testing.T) {
	t.Run("Existing", func(t *testing.T) {
		svc, err := cms.NewDocsTopicService(fixDocsTopicInformer())
		require.NoError(t, err)

		docsTopicListener := listener.NewDocsTopic(nil, nil, nil)

		svc.Subscribe(docsTopicListener)
		svc.Unsubscribe(docsTopicListener)
	})

	t.Run("Duplicated", func(t *testing.T) {
		svc, err := cms.NewDocsTopicService(fixDocsTopicInformer())
		require.NoError(t, err)

		docsTopicListener := listener.NewDocsTopic(nil, nil, nil)
		svc.Subscribe(docsTopicListener)
		svc.Subscribe(docsTopicListener)

		svc.Unsubscribe(docsTopicListener)
	})

	t.Run("Multiple", func(t *testing.T) {
		svc, err := cms.NewDocsTopicService(fixDocsTopicInformer())
		require.NoError(t, err)

		docsTopicListenerA := listener.NewDocsTopic(nil, nil, nil)
		docsTopicListenerB := listener.NewDocsTopic(nil, nil, nil)

		svc.Subscribe(docsTopicListenerA)
		svc.Subscribe(docsTopicListenerB)

		svc.Unsubscribe(docsTopicListenerA)
	})

	t.Run("Nil", func(t *testing.T) {
		svc, err := cms.NewDocsTopicService(fixDocsTopicInformer())
		require.NoError(t, err)

		svc.Unsubscribe(nil)
	})
}

func fixUnstructuredDocsTopic(metadata map[string]interface{}) *unstructured.Unstructured {
	return testingUtils.NewUnstructured(v1alpha1.SchemeGroupVersion.String(), "DocsTopic", metadata, nil, nil)
}

func fixDocsTopic(name string, labels map[string]string) *v1alpha1.DocsTopic {
	return &v1alpha1.DocsTopic{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DocsTopic",
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: DocsTopicNamespace,
			Labels:    labels,
		},
	}
}

func fixDocsTopicInformer(objects ...runtime.Object) cache.SharedIndexInformer {
	fakeClient := fake.NewSimpleDynamicClient(runtime.NewScheme(), objects...)
	informerFactory := dynamicinformer.NewDynamicSharedInformerFactory(fakeClient, 0)

	informer := informerFactory.ForResource(schema.GroupVersionResource{
		Version:  v1alpha1.SchemeGroupVersion.Version,
		Group:    v1alpha1.SchemeGroupVersion.Group,
		Resource: "docstopics",
	}).Informer()

	return informer
}
