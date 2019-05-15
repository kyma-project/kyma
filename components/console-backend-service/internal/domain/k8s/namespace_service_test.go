package k8s_test

import (
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s"
	testingUtils "github.com/kyma-project/kyma/components/console-backend-service/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
)

func TestNamespacesService_List(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		labelsWithEnvTrue := map[string]string{
			"test": "test",
			"env":  "true",
		}
		labelsWithEnvFalse := map[string]string{
			"test": "test",
			"env":  "false",
		}
		labelsWithoutEnv := map[string]string{
			"test": "test",
		}

		namespace1 := fixNamespace("namespace-name", labelsWithEnvTrue)
		namespace2 := fixNamespace("namespace-name-2", labelsWithEnvFalse)
		namespace3 := fixNamespace("namespace-name-3", labelsWithoutEnv)
		fixedInformer, _ := fixNamespaceInformer(namespace1, namespace2, namespace3)
		svc, err := k8s.NewNamespaceService(fixedInformer, nil)
		require.NoError(t, err)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, fixedInformer)

		namespaces, err := svc.List()
		require.NoError(t, err)
		assert.ElementsMatch(t, []*v1.Namespace{
			namespace1,
		}, namespaces)
	})

	t.Run("NotFound", func(t *testing.T) {
		labelsWithoutEnv := map[string]string{
			"test": "test",
		}

		namespace := fixNamespace("namespace-name", labelsWithoutEnv)
		fixedInformer, _ := fixNamespaceInformer(namespace)
		svc, err := k8s.NewNamespaceService(fixedInformer, nil)
		require.NoError(t, err)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, fixedInformer)

		namespaces, err := svc.List()
		require.NoError(t, err)
		assert.ElementsMatch(t, []*v1.Namespace{}, namespaces)
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
