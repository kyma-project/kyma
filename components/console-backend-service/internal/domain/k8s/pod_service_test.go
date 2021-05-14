package k8s_test

import (
	"context"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/apierror"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/listener"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s"
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

func TestPodService_Find(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		instanceName := "testExample"
		namespace := "testNamespace"

		pod := fixPod(instanceName, namespace, nil)
		podInformer, _ := fixPodInformer(pod)

		svc := k8s.NewPodService(podInformer, nil)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, podInformer)

		instance, err := svc.Find(instanceName, namespace)
		require.NoError(t, err)
		assert.Equal(t, pod, instance)
	})

	t.Run("NotFound", func(t *testing.T) {
		podInformer, _ := fixPodInformer()

		svc := k8s.NewPodService(podInformer, nil)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, podInformer)

		instance, err := svc.Find("doesntExist", "notExistingNamespace")
		require.NoError(t, err)
		assert.Nil(t, instance)
	})

	t.Run("NoTypeMetaReturned", func(t *testing.T) {
		instanceName := "testExample"
		namespace := "testNamespace"

		expectedPod := fixPod(instanceName, namespace, nil)
		returnedPod := fixPodWithoutTypeMeta(instanceName, namespace, nil)
		podInformer, _ := fixPodInformer(returnedPod)

		svc := k8s.NewPodService(podInformer, nil)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, podInformer)

		instance, err := svc.Find(instanceName, namespace)
		require.NoError(t, err)
		assert.Equal(t, expectedPod, instance)
	})
}

func TestPodService_List(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		namespace := "testNamespace"
		pod1 := fixPod("pod1", namespace, nil)
		pod2 := fixPod("pod2", namespace, nil)
		pod3 := fixPod("pod3", "differentNamespace", nil)

		podInformer, _ := fixPodInformer(pod1, pod2, pod3)

		svc := k8s.NewPodService(podInformer, nil)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, podInformer)

		pods, err := svc.List(namespace, pager.PagingParams{})
		require.NoError(t, err)
		assert.Equal(t, []*v1.Pod{
			pod1, pod2,
		}, pods)
	})

	t.Run("NotFound", func(t *testing.T) {
		podInformer, _ := fixPodInformer()

		svc := k8s.NewPodService(podInformer, nil)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, podInformer)

		var emptyArray []*v1.Pod
		pods, err := svc.List("notExistingNamespace", pager.PagingParams{})
		require.NoError(t, err)
		assert.Equal(t, emptyArray, pods)
	})

	t.Run("NoTypeMetaReturned", func(t *testing.T) {
		namespace := "testNamespace"
		returnedPod1 := fixPodWithoutTypeMeta("pod1", namespace, nil)
		returnedPod2 := fixPodWithoutTypeMeta("pod2", namespace, nil)
		returnedPod3 := fixPodWithoutTypeMeta("pod3", "differentNamespace", nil)
		expectedPod1 := fixPod("pod1", namespace, nil)
		expectedPod2 := fixPod("pod2", namespace, nil)

		podInformer, _ := fixPodInformer(returnedPod1, returnedPod2, returnedPod3)

		svc := k8s.NewPodService(podInformer, nil)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, podInformer)

		pods, err := svc.List(namespace, pager.PagingParams{})
		require.NoError(t, err)
		assert.Equal(t, []*v1.Pod{
			expectedPod1, expectedPod2,
		}, pods)
	})
}

func TestPodService_Subscribe(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		podInformer, _ := fixPodInformer()
		svc := k8s.NewPodService(podInformer, nil)
		podListener := listener.NewPod(nil, nil, nil)
		svc.Subscribe(podListener)
	})

	t.Run("Duplicated", func(t *testing.T) {
		podInformer, _ := fixPodInformer()
		svc := k8s.NewPodService(podInformer, nil)
		podListener := listener.NewPod(nil, nil, nil)

		svc.Subscribe(podListener)
		svc.Subscribe(podListener)
	})

	t.Run("Multiple", func(t *testing.T) {
		podInformer, _ := fixPodInformer()
		svc := k8s.NewPodService(podInformer, nil)
		podListenerA := listener.NewPod(nil, nil, nil)
		podListenerB := listener.NewPod(nil, nil, nil)

		svc.Subscribe(podListenerA)
		svc.Subscribe(podListenerB)
	})

	t.Run("Nil", func(t *testing.T) {
		podInformer, _ := fixPodInformer()
		svc := k8s.NewPodService(podInformer, nil)

		svc.Subscribe(nil)
	})
}

func TestPodService_Unsubscribe(t *testing.T) {
	t.Run("Existing", func(t *testing.T) {
		podInformer, _ := fixPodInformer()
		svc := k8s.NewPodService(podInformer, nil)
		podListener := listener.NewPod(nil, nil, nil)
		svc.Subscribe(podListener)

		svc.Unsubscribe(podListener)
	})

	t.Run("Duplicated", func(t *testing.T) {
		podInformer, _ := fixPodInformer()
		svc := k8s.NewPodService(podInformer, nil)
		podListener := listener.NewPod(nil, nil, nil)
		svc.Subscribe(podListener)
		svc.Subscribe(podListener)

		svc.Unsubscribe(podListener)
	})

	t.Run("Multiple", func(t *testing.T) {
		podInformer, _ := fixPodInformer()
		svc := k8s.NewPodService(podInformer, nil)
		podListenerA := listener.NewPod(nil, nil, nil)
		podListenerB := listener.NewPod(nil, nil, nil)
		svc.Subscribe(podListenerA)
		svc.Subscribe(podListenerB)

		svc.Unsubscribe(podListenerA)
	})

	t.Run("Nil", func(t *testing.T) {
		podInformer, _ := fixPodInformer()
		svc := k8s.NewPodService(podInformer, nil)

		svc.Unsubscribe(nil)
	})
}

func TestPodService_Update(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		exampleName := "examplePod"
		exampleNamespace := "exampleNamespace"
		examplePod := fixPod(exampleName, exampleNamespace, nil)
		podInformer, client := fixPodInformer(examplePod)
		svc := k8s.NewPodService(podInformer, client)

		update := examplePod.DeepCopy()
		update.Labels = map[string]string{
			"example": "example",
		}

		pod, err := svc.Update(exampleName, exampleNamespace, *update)
		require.NoError(t, err)
		assert.Equal(t, update, pod)

		pod, err = client.Pods(exampleNamespace).Get(context.Background(), exampleName, metav1.GetOptions{})
		require.NoError(t, err)
		assert.Equal(t, update, pod)
	})

	t.Run("NotFound", func(t *testing.T) {
		exampleName := "examplePod"
		exampleNamespace := "exampleNamespace"
		examplePod := fixPod(exampleName, exampleNamespace, nil)
		podInformer, client := fixPodInformer()
		svc := k8s.NewPodService(podInformer, client)

		update := examplePod.DeepCopy()
		update.Labels = map[string]string{
			"example": "example",
		}

		pod, err := svc.Update(exampleName, exampleNamespace, *update)
		require.Error(t, err)
		assert.Nil(t, pod)

		pod, err = client.Pods(exampleNamespace).Get(context.Background(), exampleName, metav1.GetOptions{})
		require.Error(t, err)
		assert.Nil(t, pod)
	})

	t.Run("NameMismatch", func(t *testing.T) {
		exampleName := "examplePod"
		exampleNamespace := "exampleNamespace"
		examplePod := fixPod(exampleName, exampleNamespace, nil)
		podInformer, client := fixPodInformer(examplePod)
		svc := k8s.NewPodService(podInformer, client)

		update := examplePod.DeepCopy()
		update.Name = "NameMismatch"
		update.Labels = map[string]string{
			"example": "example",
		}

		pod, err := svc.Update(exampleName, exampleNamespace, *update)
		require.Error(t, err)
		assert.True(t, apierror.IsInvalid(err))
		assert.Nil(t, pod)

		pod, err = client.Pods(exampleNamespace).Get(context.Background(), exampleName, metav1.GetOptions{})
		require.NoError(t, err)
		assert.Equal(t, examplePod, pod)
	})

	t.Run("NamespaceMismatch", func(t *testing.T) {
		exampleName := "examplePod"
		exampleNamespace := "exampleNamespace"
		examplePod := fixPod(exampleName, exampleNamespace, nil)
		podInformer, client := fixPodInformer(examplePod)
		svc := k8s.NewPodService(podInformer, client)

		update := examplePod.DeepCopy()
		update.Namespace = "NamespaceMismatch"
		update.Labels = map[string]string{
			"example": "example",
		}

		pod, err := svc.Update(exampleName, exampleNamespace, *update)
		require.Error(t, err)
		assert.True(t, apierror.IsInvalid(err))
		assert.Nil(t, pod)

		pod, err = client.Pods(exampleNamespace).Get(context.Background(), exampleName, metav1.GetOptions{})
		require.NoError(t, err)
		assert.Equal(t, examplePod, pod)
	})

	t.Run("InvalidUpdate", func(t *testing.T) {
		exampleName := "examplePod"
		exampleNamespace := "exampleNamespace"
		examplePod := fixPod(exampleName, exampleNamespace, nil)
		podInformer, client := fixFailingPodInformer(examplePod)
		svc := k8s.NewPodService(podInformer, client)

		update := examplePod.DeepCopy()
		update.Labels = map[string]string{
			"example": "example",
		}

		pod, err := svc.Update(exampleName, exampleNamespace, *update)
		require.Error(t, err)
		assert.Nil(t, pod)

		pod, err = client.Pods(exampleNamespace).Get(context.Background(), exampleName, metav1.GetOptions{})
		require.NoError(t, err)
		assert.Equal(t, examplePod, pod)
	})

	t.Run("TypeMetaChanged", func(t *testing.T) {
		exampleName := "examplePod"
		exampleNamespace := "exampleNamespace"
		examplePod := fixPod(exampleName, exampleNamespace, nil)
		podInformer, client := fixPodInformer(examplePod)
		svc := k8s.NewPodService(podInformer, client)

		update := examplePod.DeepCopy()
		update.Kind = "OtherKind"
		pod, err := svc.Update(exampleName, exampleNamespace, *update)
		require.Error(t, err)
		assert.True(t, apierror.IsInvalid(err))
		assert.Nil(t, pod)

		update.Kind = "Pod"
		update.APIVersion = "OtherVersion"
		pod, err = svc.Update(exampleName, exampleNamespace, *update)
		require.Error(t, err)
		assert.True(t, apierror.IsInvalid(err))
		assert.Nil(t, pod)

		pod, err = client.Pods(exampleNamespace).Get(context.Background(), exampleName, metav1.GetOptions{})
		require.NoError(t, err)
		assert.Equal(t, examplePod, pod)
	})

	t.Run("NoTypeMetaReturned", func(t *testing.T) {
		exampleName := "examplePod"
		exampleNamespace := "exampleNamespace"
		returnedPod := fixPodWithoutTypeMeta(exampleName, exampleNamespace, nil)
		expectedPod := fixPod(exampleName, exampleNamespace, nil)
		podInformer, client := fixPodInformer(returnedPod)
		svc := k8s.NewPodService(podInformer, client)

		update := expectedPod.DeepCopy()
		update.Labels = map[string]string{
			"example": "example",
		}

		pod, err := svc.Update(exampleName, exampleNamespace, *update)
		require.NoError(t, err)
		assert.Equal(t, update, pod)

		pod, err = client.Pods(exampleNamespace).Get(context.Background(), exampleName, metav1.GetOptions{})
		require.NoError(t, err)
		assert.Equal(t, update, pod)
	})
}

func TestPodService_Delete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		exampleName := "examplePod"
		exampleNamespace := "exampleNamespace"
		examplePod := fixPod(exampleName, exampleNamespace, nil)
		podInformer, client := fixPodInformer(examplePod)
		svc := k8s.NewPodService(podInformer, client)

		err := svc.Delete(exampleName, exampleNamespace)

		require.NoError(t, err)
		_, err = client.Pods(exampleNamespace).Get(context.Background(), exampleName, metav1.GetOptions{})
		assert.True(t, errors.IsNotFound(err))
	})
}

func fixPod(name, namespace string, labels map[string]string) *v1.Pod {
	pod := fixPodWithoutTypeMeta(name, namespace, labels)
	pod.TypeMeta = metav1.TypeMeta{
		Kind:       "Pod",
		APIVersion: "v1",
	}
	return pod
}

func fixPodWithoutTypeMeta(name, namespace string, labels map[string]string) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
	}
}

func fixPodInformer(objects ...runtime.Object) (cache.SharedIndexInformer, corev1.CoreV1Interface) {
	client := fake.NewSimpleClientset(objects...)
	informerFactory := informers.NewSharedInformerFactory(client, 0)
	informer := informerFactory.Core().V1().Pods().Informer()

	return informer, client.CoreV1()
}

func fixFailingPodInformer(objects ...runtime.Object) (cache.SharedIndexInformer, corev1.CoreV1Interface) {
	client := fake.NewSimpleClientset(objects...)
	client.PrependReactor("update", "pods", failingReactor)
	informerFactory := informers.NewSharedInformerFactory(client, 0)
	informer := informerFactory.Core().V1().Pods().Informer()

	return informer, client.CoreV1()
}
