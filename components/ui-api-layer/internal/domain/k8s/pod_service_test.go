package k8s_test

import (
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/k8s"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"
	testingUtils "github.com/kyma-project/kyma/components/ui-api-layer/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
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
		podInformer := fixPodInformer(pod)

		svc := k8s.NewPodService(podInformer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, podInformer)

		instance, err := svc.Find(instanceName, namespace)
		require.NoError(t, err)
		assert.Equal(t, pod, instance)
	})

	t.Run("NotFound", func(t *testing.T) {
		podInformer := fixPodInformer()

		svc := k8s.NewPodService(podInformer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, podInformer)

		instance, err := svc.Find("doesntExist", "notExistingNamespace")
		require.NoError(t, err)
		assert.Nil(t, instance)
	})
}

func TestPodService_List(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		namespace := "testNamespace"
		pod1 := fixPod("pod1", namespace, nil)
		pod2 := fixPod("pod2", namespace, nil)
		pod3 := fixPod("pod3", "differentNamespace", nil)

		podInformer := fixPodInformer(pod1, pod2, pod3)

		svc := k8s.NewPodService(podInformer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, podInformer)

		pods, err := svc.List(namespace, pager.PagingParams{})
		require.NoError(t, err)
		assert.Equal(t, []*v1.Pod{
			pod1, pod2,
		}, pods)
	})

	t.Run("NotFound", func(t *testing.T) {
		podInformer := fixPodInformer()

		svc := k8s.NewPodService(podInformer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, podInformer)

		var emptyArray []*v1.Pod
		pods, err := svc.List("notExistingNamespace", pager.PagingParams{})
		require.NoError(t, err)
		assert.Equal(t, emptyArray, pods)
	})
}

func fixPod(name, environment string, labels map[string]string) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: environment,
			Labels:    labels,
		},
	}
}

func fixPodInformer(objects ...runtime.Object) cache.SharedIndexInformer {
	client := fake.NewSimpleClientset(objects...)
	informerFactory := informers.NewSharedInformerFactory(client, 0)
	informer := informerFactory.Core().V1().Pods().Informer()

	return informer
}
