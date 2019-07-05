package k8s_test

import (
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s"
	testingUtils "github.com/kyma-project/kyma/components/console-backend-service/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
)

func TestNamespacesService_List(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		labels := map[string]string{
			"test": "test",
			"env":  "true",
		}
		emptyLabels := map[string]string{}

		namespace1 := fixNamespace("namespace-name", labels)
		namespace2 := fixNamespace("namespace-name-2", emptyLabels)
		fixedInformer, _ := fixNamespaceInformer(namespace1, namespace2)
		svc, err := k8s.NewNamespaceService(fixedInformer, nil)
		require.NoError(t, err)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, fixedInformer)

		namespaces, err := svc.List()
		require.NoError(t, err)
		assert.ElementsMatch(t, []*v1.Namespace{
			namespace1, namespace2,
		}, namespaces)
	})

	t.Run("NotFound", func(t *testing.T) {
		fixedInformer, _ := fixNamespaceInformer()
		svc, err := k8s.NewNamespaceService(fixedInformer, nil)
		require.NoError(t, err)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, fixedInformer)

		namespaces, err := svc.List()
		require.NoError(t, err)
		assert.ElementsMatch(t, []*v1.Namespace{}, namespaces)
	})
}

func TestNamespacesService_Find(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		name := "namespace-name"
		labels := map[string]string{
			"test": "test",
			"env":  "true",
		}

		namespace1 := fixNamespace(name, labels)
		fixedInformer, _ := fixNamespaceInformer(namespace1)
		svc, err := k8s.NewNamespaceService(fixedInformer, nil)
		require.NoError(t, err)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, fixedInformer)

		namespace, err := svc.Find(name)
		require.NoError(t, err)
		assert.Equal(t, namespace1, namespace)
	})

	t.Run("Not Found", func(t *testing.T) {
		fixedInformer, _ := fixNamespaceInformer()
		svc, err := k8s.NewNamespaceService(fixedInformer, nil)
		require.NoError(t, err)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, fixedInformer)

		namespace, err := svc.Find("name")
		require.NoError(t, err)
		var empty *v1.Namespace
		assert.Equal(t, empty, namespace)
	})
}

func TestNamespacesService_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		name := "namespace"
		labels := map[string]string{
			"test": "test",
			"env":  "true",
		}

		fixedInformer, client := fixNamespaceInformer()
		svc, err := k8s.NewNamespaceService(fixedInformer, client)
		require.NoError(t, err)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, fixedInformer)

		namespace, err := svc.Create(name, labels)
		require.NoError(t, err)
		assert.Equal(t, name, namespace.Name)
		assert.Equal(t, labels["env"], namespace.Labels["env"])
		assert.Equal(t, labels["test"], namespace.Labels["test"])
	})
}

func TestNamespacesService_Delete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		name := "namespace"
		labels := map[string]string{
			"test": "test",
			"env":  "true",
		}
		namespace1 := fixNamespace(name, labels)
		fixedInformer, client := fixNamespaceInformer(namespace1)
		svc, err := k8s.NewNamespaceService(fixedInformer, client)
		require.NoError(t, err)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, fixedInformer)

		err = svc.Delete(name)
		require.NoError(t, err)
	})

	t.Run("Not Found", func(t *testing.T) {
		fixedInformer, client := fixNamespaceInformer()
		svc, err := k8s.NewNamespaceService(fixedInformer, client)
		require.NoError(t, err)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, fixedInformer)

		err = svc.Delete("name")
		require.Error(t, err)
	})
}

func TestNamespacesService_Update(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		name := "namespace"
		labels := map[string]string{
			"test": "test",
			"env":  "true",
		}
		newLabels := map[string]string{
			"test": "test2",
		}

		namespace1 := fixNamespace(name, labels)
		fixedInformer, client := fixNamespaceInformer(namespace1)
		svc, err := k8s.NewNamespaceService(fixedInformer, client)
		require.NoError(t, err)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, fixedInformer)

		namespace, err := svc.Update(name, newLabels)
		require.NoError(t, err)

		assert.Equal(t, name, namespace.Name)
		assert.Equal(t, newLabels["test"], namespace.Labels["test"])
	})
}

func fixNamespace(name string, labels map[string]string) *v1.Namespace {
	namespace := fixNamespaceWithoutTypeMeta(name, labels)
	namespace.TypeMeta = metav1.TypeMeta{
		Kind:       "Namespace",
		APIVersion: "v1",
	}
	return namespace
}
func fixNamespaceWithoutTypeMeta(name string, labels map[string]string) *v1.Namespace {
	return &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
	}
}

func fixNamespaceInformer(objects ...runtime.Object) (cache.SharedIndexInformer, corev1.CoreV1Interface) {
	client := fake.NewSimpleClientset(objects...)
	informerFactory := informers.NewSharedInformerFactory(client, 0)
	informer := informerFactory.Core().V1().Namespaces().Informer()

	return informer, client.CoreV1()
}
