package k8s_test

import (
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/listener"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
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

func TestSecretService_Find(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		instanceName := "testExample"
		namespace := "testNamespace"

		secret := fixSecret(instanceName, namespace, nil)
		secretInformer, _ := fixSecretInformer(secret)

		svc := k8s.NewSecretService(secretInformer, nil)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, secretInformer)

		instance, err := svc.Find(instanceName, namespace)
		require.NoError(t, err)
		assert.Equal(t, secret, instance)
	})

	t.Run("NotFound", func(t *testing.T) {
		secretInformer, _ := fixSecretInformer()

		svc := k8s.NewSecretService(secretInformer, nil)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, secretInformer)

		instance, err := svc.Find("doesntExist", "notExistingNamespace")
		require.NoError(t, err)
		assert.Nil(t, instance)
	})

	t.Run("NoTypeMetaReturned", func(t *testing.T) {
		secretName := "testExample"
		namespace := "testNamespace"

		expectedSecret := fixSecret(secretName, namespace, nil)
		returnedSecret := fixSecretWithoutTypeMeta(secretName, namespace, nil)
		secretInformer, _ := fixSecretInformer(returnedSecret)

		svc := k8s.NewSecretService(secretInformer, nil)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, secretInformer)

		instance, err := svc.Find(secretName, namespace)
		require.NoError(t, err)
		assert.Equal(t, expectedSecret, instance)
	})
}

func TestSecretService_List(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		namespace := "testNamespace"
		secret1 := fixSecret("secret1", namespace, nil)
		secret2 := fixSecret("secret2", namespace, nil)
		secret3 := fixSecret("secret3", "differentNamespace", nil)

		secretInformer, _ := fixSecretInformer(secret1, secret2, secret3)

		svc := k8s.NewSecretService(secretInformer, nil)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, secretInformer)

		secrets, err := svc.List(namespace, pager.PagingParams{})
		require.NoError(t, err)
		assert.Equal(t, []*v1.Secret{
			secret1, secret2,
		}, secrets)
	})

	t.Run("NotFound", func(t *testing.T) {
		secretInformer, _ := fixSecretInformer()

		svc := k8s.NewSecretService(secretInformer, nil)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, secretInformer)

		var emptyArray []*v1.Secret
		secrets, err := svc.List("notExistingNamespace", pager.PagingParams{})
		require.NoError(t, err)
		assert.Equal(t, emptyArray, secrets)
	})

	t.Run("NoTypeMetaReturned", func(t *testing.T) {
		namespace := "testNamespace"
		returnedSecret1 := fixSecretWithoutTypeMeta("secret1", namespace, nil)
		returnedSecret2 := fixSecretWithoutTypeMeta("secret2", namespace, nil)
		returnedSecret3 := fixSecretWithoutTypeMeta("secret3", "differentNamespace", nil)
		expectedSecret1 := fixSecret("secret1", namespace, nil)
		expectedSecret2 := fixSecret("secret2", namespace, nil)

		secretInformer, _ := fixSecretInformer(returnedSecret1, returnedSecret2, returnedSecret3)

		svc := k8s.NewSecretService(secretInformer, nil)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, secretInformer)

		secrets, err := svc.List(namespace, pager.PagingParams{})
		require.NoError(t, err)
		assert.Equal(t, []*v1.Secret{
			expectedSecret1, expectedSecret2,
		}, secrets)
	})
}

func TestSecretService_Subscribe(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		secretInformer, _ := fixSecretInformer()
		svc := k8s.NewSecretService(secretInformer, nil)
		secretListener := listener.NewSecret(nil, nil, nil)
		svc.Subscribe(secretListener)
	})

	t.Run("Duplicated", func(t *testing.T) {
		secretInformer, _ := fixSecretInformer()
		svc := k8s.NewSecretService(secretInformer, nil)
		secretListener := listener.NewSecret(nil, nil, nil)

		svc.Subscribe(secretListener)
		svc.Subscribe(secretListener)
	})

	t.Run("Multiple", func(t *testing.T) {
		secretInformer, _ := fixSecretInformer()
		svc := k8s.NewSecretService(secretInformer, nil)
		secretListenerA := listener.NewSecret(nil, nil, nil)
		secretListenerB := listener.NewSecret(nil, nil, nil)

		svc.Subscribe(secretListenerA)
		svc.Subscribe(secretListenerB)
	})

	t.Run("Nil", func(t *testing.T) {
		secretInformer, _ := fixSecretInformer()
		svc := k8s.NewSecretService(secretInformer, nil)

		svc.Subscribe(nil)
	})
}

func TestSecretService_Unsubscribe(t *testing.T) {
	t.Run("Existing", func(t *testing.T) {
		secretInformer, _ := fixSecretInformer()
		svc := k8s.NewSecretService(secretInformer, nil)
		secretListener := listener.NewSecret(nil, nil, nil)
		svc.Subscribe(secretListener)

		svc.Unsubscribe(secretListener)
	})

	t.Run("Duplicated", func(t *testing.T) {
		secretInformer, _ := fixSecretInformer()
		svc := k8s.NewSecretService(secretInformer, nil)
		secretListener := listener.NewSecret(nil, nil, nil)
		svc.Subscribe(secretListener)
		svc.Subscribe(secretListener)

		svc.Unsubscribe(secretListener)
	})

	t.Run("Multiple", func(t *testing.T) {
		secretInformer, _ := fixSecretInformer()
		svc := k8s.NewSecretService(secretInformer, nil)
		secretListenerA := listener.NewSecret(nil, nil, nil)
		secretListenerB := listener.NewSecret(nil, nil, nil)
		svc.Subscribe(secretListenerA)
		svc.Subscribe(secretListenerB)

		svc.Unsubscribe(secretListenerA)
	})

	t.Run("Nil", func(t *testing.T) {
		secretInformer, _ := fixSecretInformer()
		svc := k8s.NewSecretService(secretInformer, nil)

		svc.Unsubscribe(nil)
	})
}

func fixSecretWithoutTypeMeta(name, namespace string, labels map[string]string) *v1.Secret {
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
	}
}

func fixSecretInformer(objects ...runtime.Object) (cache.SharedIndexInformer, corev1.CoreV1Interface) {
	client := fake.NewSimpleClientset(objects...)
	informerFactory := informers.NewSharedInformerFactory(client, 0)
	informer := informerFactory.Core().V1().Secrets().Informer()

	return informer, client.CoreV1()
}

func fixFailingSecretInformer(objects ...runtime.Object) (cache.SharedIndexInformer, corev1.CoreV1Interface) {
	client := fake.NewSimpleClientset(objects...)
	client.PrependReactor("update", "secrets", failingReactor)
	informerFactory := informers.NewSharedInformerFactory(client, 0)
	informer := informerFactory.Core().V1().Secrets().Informer()

	return informer, client.CoreV1()
}

func fixSecret(name, namespace string, labels map[string]string) *v1.Secret {
	secret := fixSecretWithoutTypeMeta(name, namespace, labels)
	secret.TypeMeta = metav1.TypeMeta{
		Kind:       "Secret",
		APIVersion: "v1",
	}
	return secret
}
