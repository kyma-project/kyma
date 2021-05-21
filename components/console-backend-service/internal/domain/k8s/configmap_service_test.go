package k8s_test

import (
	"context"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/apierror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/listener"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
	testingUtils "github.com/kyma-project/kyma/components/console-backend-service/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
)

func TestConfigMapService_Find(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		instanceName := "testExample"
		namespace := "testNamespace"

		configMap := fixConfigMap(instanceName, namespace, nil)
		configMapInformer, _ := fixConfigMapInformer(configMap)

		svc := k8s.NewConfigMapService(configMapInformer, nil)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, configMapInformer)

		instance, err := svc.Find(instanceName, namespace)
		require.NoError(t, err)
		assert.Equal(t, configMap, instance)
	})

	t.Run("NotFound", func(t *testing.T) {
		configMapInformer, _ := fixConfigMapInformer()

		svc := k8s.NewConfigMapService(configMapInformer, nil)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, configMapInformer)

		instance, err := svc.Find("doesntExist", "notExistingNamespace")
		require.NoError(t, err)
		assert.Nil(t, instance)
	})

	t.Run("NoTypeMetaReturned", func(t *testing.T) {
		instanceName := "testExample"
		namespace := "testNamespace"

		expectedConfigMap := fixConfigMap(instanceName, namespace, nil)
		returnedConfigMap := fixConfigMapWithoutTypeMeta(instanceName, namespace, nil)
		configMapInformer, _ := fixConfigMapInformer(returnedConfigMap)

		svc := k8s.NewConfigMapService(configMapInformer, nil)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, configMapInformer)

		instance, err := svc.Find(instanceName, namespace)
		require.NoError(t, err)
		assert.Equal(t, expectedConfigMap, instance)
	})
}

func TestConfigMapService_List(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		namespace := "testNamespace"
		configMap1 := fixConfigMap("configMap1", namespace, nil)
		configMap2 := fixConfigMap("configMap2", namespace, nil)
		configMap3 := fixConfigMap("configMap3", "differentNamespace", nil)

		configMapInformer, _ := fixConfigMapInformer(configMap1, configMap2, configMap3)

		svc := k8s.NewConfigMapService(configMapInformer, nil)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, configMapInformer)

		configMaps, err := svc.List(namespace, pager.PagingParams{})
		require.NoError(t, err)
		assert.Equal(t, []*v1.ConfigMap{
			configMap1, configMap2,
		}, configMaps)
	})

	t.Run("NotFound", func(t *testing.T) {
		configMapInformer, _ := fixConfigMapInformer()

		svc := k8s.NewConfigMapService(configMapInformer, nil)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, configMapInformer)

		var emptyArray []*v1.ConfigMap
		configMaps, err := svc.List("notExistingNamespace", pager.PagingParams{})
		require.NoError(t, err)
		assert.Equal(t, emptyArray, configMaps)
	})

	t.Run("NoTypeMetaReturned", func(t *testing.T) {
		namespace := "testNamespace"
		returnedConfigMap1 := fixConfigMapWithoutTypeMeta("configMap1", namespace, nil)
		returnedConfigMap2 := fixConfigMapWithoutTypeMeta("configMap2", namespace, nil)
		returnedConfigMap3 := fixConfigMapWithoutTypeMeta("configMap3", "differentNamespace", nil)
		expectedConfigMap1 := fixConfigMap("configMap1", namespace, nil)
		expectedConfigMap2 := fixConfigMap("configMap2", namespace, nil)

		configMapInformer, _ := fixConfigMapInformer(returnedConfigMap1, returnedConfigMap2, returnedConfigMap3)

		svc := k8s.NewConfigMapService(configMapInformer, nil)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, configMapInformer)

		configMaps, err := svc.List(namespace, pager.PagingParams{})
		require.NoError(t, err)
		assert.Equal(t, []*v1.ConfigMap{
			expectedConfigMap1, expectedConfigMap2,
		}, configMaps)
	})
}

func TestConfigMapService_Update(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		exampleName := "exampleConfigMap"
		exampleNamespace := "exampleNamespace"
		exampleConfigMap := fixConfigMap(exampleName, exampleNamespace, nil)
		configMapInformer, client := fixConfigMapInformer(exampleConfigMap)
		svc := k8s.NewConfigMapService(configMapInformer, client)

		update := exampleConfigMap.DeepCopy()
		update.Labels = map[string]string{
			"example": "example",
		}

		configMap, err := svc.Update(exampleName, exampleNamespace, *update)
		require.NoError(t, err)
		assert.Equal(t, update, configMap)

		configMap, err = client.ConfigMaps(exampleNamespace).Get(context.Background(), exampleName, metav1.GetOptions{})
		require.NoError(t, err)
		assert.Equal(t, update, configMap)
	})

	t.Run("NotFound", func(t *testing.T) {
		exampleName := "exampleConfigMap"
		exampleNamespace := "exampleNamespace"
		exampleConfigMap := fixConfigMap(exampleName, exampleNamespace, nil)
		configMapInformer, client := fixConfigMapInformer()
		svc := k8s.NewConfigMapService(configMapInformer, client)

		update := exampleConfigMap.DeepCopy()
		update.Labels = map[string]string{
			"example": "example",
		}

		configMap, err := svc.Update(exampleName, exampleNamespace, *update)
		require.Error(t, err)
		assert.Nil(t, configMap)

		configMap, err = client.ConfigMaps(exampleNamespace).Get(context.Background(), exampleName, metav1.GetOptions{})
		require.Error(t, err)
		assert.Nil(t, configMap)
	})

	t.Run("NameMismatch", func(t *testing.T) {
		exampleName := "exampleConfigMap"
		exampleNamespace := "exampleNamespace"
		exampleConfigMap := fixConfigMap(exampleName, exampleNamespace, nil)
		configMapInformer, client := fixConfigMapInformer(exampleConfigMap)
		svc := k8s.NewConfigMapService(configMapInformer, client)

		update := exampleConfigMap.DeepCopy()
		update.Name = "NameMismatch"
		update.Labels = map[string]string{
			"example": "example",
		}

		configMap, err := svc.Update(exampleName, exampleNamespace, *update)
		require.Error(t, err)
		assert.True(t, apierror.IsInvalid(err))
		assert.Nil(t, configMap)

		configMap, err = client.ConfigMaps(exampleNamespace).Get(context.Background(), exampleName, metav1.GetOptions{})
		require.NoError(t, err)
		assert.Equal(t, exampleConfigMap, configMap)
	})

	t.Run("NamespaceMismatch", func(t *testing.T) {
		exampleName := "exampleConfigMap"
		exampleNamespace := "exampleNamespace"
		exampleConfigMap := fixConfigMap(exampleName, exampleNamespace, nil)
		configMapInformer, client := fixConfigMapInformer(exampleConfigMap)
		svc := k8s.NewConfigMapService(configMapInformer, client)

		update := exampleConfigMap.DeepCopy()
		update.Namespace = "NamespaceMismatch"
		update.Labels = map[string]string{
			"example": "example",
		}

		configMap, err := svc.Update(exampleName, exampleNamespace, *update)
		require.Error(t, err)
		assert.True(t, apierror.IsInvalid(err))
		assert.Nil(t, configMap)

		configMap, err = client.ConfigMaps(exampleNamespace).Get(context.Background(), exampleName, metav1.GetOptions{})
		require.NoError(t, err)
		assert.Equal(t, exampleConfigMap, configMap)
	})

	t.Run("InvalidUpdate", func(t *testing.T) {
		exampleName := "exampleConfigMap"
		exampleNamespace := "exampleNamespace"
		exampleConfigMap := fixConfigMap(exampleName, exampleNamespace, nil)
		configMapInformer, client := fixFailingConfigMapInformer(exampleConfigMap)
		svc := k8s.NewConfigMapService(configMapInformer, client)

		update := exampleConfigMap.DeepCopy()
		update.Labels = map[string]string{
			"example": "example",
		}

		configMap, err := svc.Update(exampleName, exampleNamespace, *update)
		require.Error(t, err)
		assert.Nil(t, configMap)

		configMap, err = client.ConfigMaps(exampleNamespace).Get(context.Background(), exampleName, metav1.GetOptions{})
		require.NoError(t, err)
		assert.Equal(t, exampleConfigMap, configMap)
	})

	t.Run("TypeMetaChanged", func(t *testing.T) {
		exampleName := "exampleConfigMap"
		exampleNamespace := "exampleNamespace"
		exampleConfigMap := fixConfigMap(exampleName, exampleNamespace, nil)
		configMapInformer, client := fixConfigMapInformer(exampleConfigMap)
		svc := k8s.NewConfigMapService(configMapInformer, client)

		update := exampleConfigMap.DeepCopy()
		update.Kind = "OtherKind"
		configMap, err := svc.Update(exampleName, exampleNamespace, *update)
		require.Error(t, err)
		assert.True(t, apierror.IsInvalid(err))
		assert.Nil(t, configMap)

		update.Kind = "ConfigMap"
		update.APIVersion = "OtherVersion"
		configMap, err = svc.Update(exampleName, exampleNamespace, *update)
		require.Error(t, err)
		assert.True(t, apierror.IsInvalid(err))
		assert.Nil(t, configMap)

		configMap, err = client.ConfigMaps(exampleNamespace).Get(context.Background(), exampleName, metav1.GetOptions{})
		require.NoError(t, err)
		assert.Equal(t, exampleConfigMap, configMap)
	})

	t.Run("NoTypeMetaReturned", func(t *testing.T) {
		exampleName := "exampleConfigMap"
		exampleNamespace := "exampleNamespace"
		returnedConfigMap := fixConfigMapWithoutTypeMeta(exampleName, exampleNamespace, nil)
		expectedConfigMap := fixConfigMap(exampleName, exampleNamespace, nil)
		configMapInformer, client := fixConfigMapInformer(returnedConfigMap)
		svc := k8s.NewConfigMapService(configMapInformer, client)

		update := expectedConfigMap.DeepCopy()
		update.Labels = map[string]string{
			"example": "example",
		}

		configMap, err := svc.Update(exampleName, exampleNamespace, *update)
		require.NoError(t, err)
		assert.Equal(t, update, configMap)

		configMap, err = client.ConfigMaps(exampleNamespace).Get(context.Background(), exampleName, metav1.GetOptions{})
		require.NoError(t, err)
		assert.Equal(t, update, configMap)
	})
}

func TestConfigMapService_Delete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		exampleName := "exampleConfigMap"
		exampleNamespace := "exampleNamespace"
		exampleConfigMap := fixConfigMap(exampleName, exampleNamespace, nil)
		configMapInformer, client := fixConfigMapInformer(exampleConfigMap)
		svc := k8s.NewConfigMapService(configMapInformer, client)

		err := svc.Delete(exampleName, exampleNamespace)

		require.NoError(t, err)
		_, err = client.ConfigMaps(exampleNamespace).Get(context.Background(), exampleName, metav1.GetOptions{})
		assert.True(t, errors.IsNotFound(err))
	})
}

func TestConfigMapService_Subscribe(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		configMapInformer, _ := fixConfigMapInformer()
		svc := k8s.NewConfigMapService(configMapInformer, nil)
		configMapListener := listener.NewConfigMap(nil, nil, nil)
		svc.Subscribe(configMapListener)
	})

	t.Run("Duplicated", func(t *testing.T) {
		configMapInformer, _ := fixConfigMapInformer()
		svc := k8s.NewConfigMapService(configMapInformer, nil)
		configMapListener := listener.NewConfigMap(nil, nil, nil)

		svc.Subscribe(configMapListener)
		svc.Subscribe(configMapListener)
	})

	t.Run("Multiple", func(t *testing.T) {
		configMapInformer, _ := fixConfigMapInformer()
		svc := k8s.NewConfigMapService(configMapInformer, nil)
		configMapListenerA := listener.NewConfigMap(nil, nil, nil)
		configMapListenerB := listener.NewConfigMap(nil, nil, nil)

		svc.Subscribe(configMapListenerA)
		svc.Subscribe(configMapListenerB)
	})

	t.Run("Nil", func(t *testing.T) {
		configMapInformer, _ := fixConfigMapInformer()
		svc := k8s.NewConfigMapService(configMapInformer, nil)

		svc.Subscribe(nil)
	})
}

func TestConfigMapService_Unsubscribe(t *testing.T) {
	t.Run("Existing", func(t *testing.T) {
		configMapInformer, _ := fixConfigMapInformer()
		svc := k8s.NewConfigMapService(configMapInformer, nil)
		configMapListener := listener.NewConfigMap(nil, nil, nil)
		svc.Subscribe(configMapListener)

		svc.Unsubscribe(configMapListener)
	})

	t.Run("Duplicated", func(t *testing.T) {
		configMapInformer, _ := fixConfigMapInformer()
		svc := k8s.NewConfigMapService(configMapInformer, nil)
		configMapListener := listener.NewConfigMap(nil, nil, nil)
		svc.Subscribe(configMapListener)
		svc.Subscribe(configMapListener)

		svc.Unsubscribe(configMapListener)
	})

	t.Run("Multiple", func(t *testing.T) {
		configMapInformer, _ := fixConfigMapInformer()
		svc := k8s.NewConfigMapService(configMapInformer, nil)
		configMapListenerA := listener.NewConfigMap(nil, nil, nil)
		configMapListenerB := listener.NewConfigMap(nil, nil, nil)
		svc.Subscribe(configMapListenerA)
		svc.Subscribe(configMapListenerB)

		svc.Unsubscribe(configMapListenerA)
	})

	t.Run("Nil", func(t *testing.T) {
		configMapInformer, _ := fixConfigMapInformer()
		svc := k8s.NewConfigMapService(configMapInformer, nil)

		svc.Unsubscribe(nil)
	})
}

func fixConfigMap(name, namespace string, labels map[string]string) *v1.ConfigMap {
	configMap := fixConfigMapWithoutTypeMeta(name, namespace, labels)
	configMap.TypeMeta = metav1.TypeMeta{
		Kind:       "ConfigMap",
		APIVersion: "v1",
	}
	return configMap
}

func fixConfigMapWithoutTypeMeta(name, namespace string, labels map[string]string) *v1.ConfigMap {
	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
	}
}

func fixConfigMapInformer(objects ...runtime.Object) (cache.SharedIndexInformer, corev1.CoreV1Interface) {
	client := fake.NewSimpleClientset(objects...)
	informerFactory := informers.NewSharedInformerFactory(client, 0)
	informer := informerFactory.Core().V1().ConfigMaps().Informer()

	return informer, client.CoreV1()
}

func fixFailingConfigMapInformer(objects ...runtime.Object) (cache.SharedIndexInformer, corev1.CoreV1Interface) {
	client := fake.NewSimpleClientset(objects...)
	client.PrependReactor("update", "configmaps", failingReactor)
	informerFactory := informers.NewSharedInformerFactory(client, 0)
	informer := informerFactory.Core().V1().ConfigMaps().Informer()

	return informer, client.CoreV1()
}
