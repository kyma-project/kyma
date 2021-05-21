package k8s_test

import (
	"context"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
	testingUtils "github.com/kyma-project/kyma/components/console-backend-service/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apps "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	appsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	"k8s.io/client-go/tools/cache"
)

func TestReplicaSetService_Find(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		instanceName := "testExample"
		namespace := "testNamespace"

		replicaSet := fixReplicaSet(instanceName, namespace, map[string]string{"test": "test"})
		replicaSetInformer, _ := fixReplicaSetInformer(replicaSet)

		svc := k8s.NewReplicaSetService(replicaSetInformer, nil)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, replicaSetInformer)

		instance, err := svc.Find(instanceName, namespace)
		require.NoError(t, err)
		assert.Equal(t, replicaSet, instance)
	})

	t.Run("NotFound", func(t *testing.T) {
		replicaSetInformer, _ := fixReplicaSetInformer()

		svc := k8s.NewReplicaSetService(replicaSetInformer, nil)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, replicaSetInformer)

		instance, err := svc.Find("doesntExist", "notExistingNamespace")
		require.NoError(t, err)
		assert.Nil(t, instance)
	})

	t.Run("NoTypeMetaReturned", func(t *testing.T) {
		instanceName := "testExample"
		namespace := "testNamespace"

		expectedReplicaSet := fixReplicaSet(instanceName, namespace, map[string]string{"test": "test"})
		returnedReplicaSet := fixReplicaSetWithoutTypeMeta(instanceName, namespace, map[string]string{"test": "test"})
		replicaSetInformer, _ := fixReplicaSetInformer(returnedReplicaSet)

		svc := k8s.NewReplicaSetService(replicaSetInformer, nil)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, replicaSetInformer)

		instance, err := svc.Find(instanceName, namespace)
		require.NoError(t, err)
		assert.Equal(t, expectedReplicaSet, instance)
	})
}

func TestReplicaSetService_List(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		namespace := "testNamespace"
		replicaSet1 := fixReplicaSet("replicaSet1", namespace, nil)
		replicaSet2 := fixReplicaSet("replicaSet2", namespace, nil)
		replicaSet3 := fixReplicaSet("replicaSet3", "differentNamespace", nil)

		replicaSetInformer, _ := fixReplicaSetInformer(replicaSet1, replicaSet2, replicaSet3)

		svc := k8s.NewReplicaSetService(replicaSetInformer, nil)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, replicaSetInformer)

		replicaSets, err := svc.List(namespace, pager.PagingParams{})
		require.NoError(t, err)
		assert.Equal(t, []*apps.ReplicaSet{
			replicaSet1, replicaSet2,
		}, replicaSets)
	})

	t.Run("NotFound", func(t *testing.T) {
		replicaSetInformer, _ := fixReplicaSetInformer()

		svc := k8s.NewReplicaSetService(replicaSetInformer, nil)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, replicaSetInformer)

		var emptyArray []*apps.ReplicaSet
		replicaSets, err := svc.List("notExistingNamespace", pager.PagingParams{})
		require.NoError(t, err)
		assert.Equal(t, emptyArray, replicaSets)
	})

	t.Run("NoTypeMetaReturned", func(t *testing.T) {
		namespace := "testNamespace"
		returnedReplicaSet1 := fixReplicaSetWithoutTypeMeta("replicaSet1", namespace, nil)
		returnedReplicaSet2 := fixReplicaSetWithoutTypeMeta("replicaSet2", namespace, nil)
		returnedReplicaSet3 := fixReplicaSetWithoutTypeMeta("replicaSet3", "differentNamespace", nil)
		expectedReplicaSet1 := fixReplicaSet("replicaSet1", namespace, nil)
		expectedReplicaSet2 := fixReplicaSet("replicaSet2", namespace, nil)

		replicaSetInformer, _ := fixReplicaSetInformer(returnedReplicaSet1, returnedReplicaSet2, returnedReplicaSet3)

		svc := k8s.NewReplicaSetService(replicaSetInformer, nil)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, replicaSetInformer)

		replicaSets, err := svc.List(namespace, pager.PagingParams{})
		require.NoError(t, err)
		assert.Equal(t, []*apps.ReplicaSet{
			expectedReplicaSet1, expectedReplicaSet2,
		}, replicaSets)
	})
}

func TestReplicaSetService_Update(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		exampleName := "exampleReplicaSet"
		exampleNamespace := "exampleNamespace"
		exampleReplicaSet := fixReplicaSet(exampleName, exampleNamespace, map[string]string{"test": "test"})
		replicaSetInformer, client := fixReplicaSetInformer(exampleReplicaSet)
		svc := k8s.NewReplicaSetService(replicaSetInformer, client)

		update := exampleReplicaSet.DeepCopy()
		update.Labels = map[string]string{
			"example": "example",
		}

		replicaSet, err := svc.Update(exampleName, exampleNamespace, *update)
		require.NoError(t, err)
		assert.Equal(t, update, replicaSet)

		replicaSet, err = client.ReplicaSets(exampleNamespace).Get(context.Background(), exampleName, metav1.GetOptions{})
		require.NoError(t, err)
		assert.Equal(t, update, replicaSet)
	})

	t.Run("NotFound", func(t *testing.T) {
		exampleName := "exampleReplicaSet"
		exampleNamespace := "exampleNamespace"
		exampleReplicaSet := fixReplicaSet(exampleName, exampleNamespace, nil)
		replicaSetInformer, client := fixReplicaSetInformer()
		svc := k8s.NewReplicaSetService(replicaSetInformer, client)

		update := exampleReplicaSet.DeepCopy()
		update.Labels = map[string]string{
			"example": "example",
		}

		replicaSet, err := svc.Update(exampleName, exampleNamespace, *update)
		require.Error(t, err)
		assert.Nil(t, replicaSet)

		replicaSet, err = client.ReplicaSets(exampleNamespace).Get(context.Background(), exampleName, metav1.GetOptions{})
		require.Error(t, err)
		assert.Nil(t, replicaSet)
	})

	t.Run("NameMismatch", func(t *testing.T) {
		exampleName := "exampleReplicaSet"
		exampleNamespace := "exampleNamespace"
		exampleReplicaSet := fixReplicaSet(exampleName, exampleNamespace, nil)
		replicaSetInformer, client := fixReplicaSetInformer(exampleReplicaSet)
		svc := k8s.NewReplicaSetService(replicaSetInformer, client)

		update := exampleReplicaSet.DeepCopy()
		update.Name = "NameMismatch"
		update.Labels = map[string]string{
			"example": "example",
		}

		replicaSet, err := svc.Update(exampleName, exampleNamespace, *update)
		require.Error(t, err)
		assert.True(t, errors.IsInvalid(err))
		assert.Nil(t, replicaSet)

		replicaSet, err = client.ReplicaSets(exampleNamespace).Get(context.Background(), exampleName, metav1.GetOptions{})
		require.NoError(t, err)
		assert.Equal(t, exampleReplicaSet, replicaSet)
	})

	t.Run("NamespaceMismatch", func(t *testing.T) {
		exampleName := "exampleReplicaSet"
		exampleNamespace := "exampleNamespace"
		exampleReplicaSet := fixReplicaSet(exampleName, exampleNamespace, nil)
		replicaSetInformer, client := fixReplicaSetInformer(exampleReplicaSet)
		svc := k8s.NewReplicaSetService(replicaSetInformer, client)

		update := exampleReplicaSet.DeepCopy()
		update.Namespace = "NamespaceMismatch"
		update.Labels = map[string]string{
			"example": "example",
		}

		replicaSet, err := svc.Update(exampleName, exampleNamespace, *update)
		require.Error(t, err)
		assert.True(t, errors.IsInvalid(err))
		assert.Nil(t, replicaSet)

		replicaSet, err = client.ReplicaSets(exampleNamespace).Get(context.Background(), exampleName, metav1.GetOptions{})
		require.NoError(t, err)
		assert.Equal(t, exampleReplicaSet, replicaSet)
	})

	t.Run("InvalidUpdate", func(t *testing.T) {
		exampleName := "exampleReplicaSet"
		exampleNamespace := "exampleNamespace"
		exampleReplicaSet := fixReplicaSet(exampleName, exampleNamespace, nil)
		replicaSetInformer, client := fixFailingReplicaSetInformer(exampleReplicaSet)
		svc := k8s.NewReplicaSetService(replicaSetInformer, client)

		update := exampleReplicaSet.DeepCopy()
		update.Labels = map[string]string{
			"example": "example",
		}

		replicaSet, err := svc.Update(exampleName, exampleNamespace, *update)
		require.Error(t, err)
		assert.Nil(t, replicaSet)

		replicaSet, err = client.ReplicaSets(exampleNamespace).Get(context.Background(), exampleName, metav1.GetOptions{})
		require.NoError(t, err)
		assert.Equal(t, exampleReplicaSet, replicaSet)
	})

	t.Run("TypeMetaChanged", func(t *testing.T) {
		exampleName := "exampleReplicaSet"
		exampleNamespace := "exampleNamespace"
		exampleReplicaSet := fixReplicaSet(exampleName, exampleNamespace, nil)
		replicaSetInformer, client := fixReplicaSetInformer(exampleReplicaSet)
		svc := k8s.NewReplicaSetService(replicaSetInformer, client)

		update := exampleReplicaSet.DeepCopy()
		update.Kind = "OtherKind"
		replicaSet, err := svc.Update(exampleName, exampleNamespace, *update)
		require.Error(t, err)
		assert.True(t, errors.IsInvalid(err))
		assert.Nil(t, replicaSet)

		update.Kind = "ReplicaSet"
		update.APIVersion = "OtherVersion"
		replicaSet, err = svc.Update(exampleName, exampleNamespace, *update)
		require.Error(t, err)
		assert.True(t, errors.IsInvalid(err))
		assert.Nil(t, replicaSet)

		replicaSet, err = client.ReplicaSets(exampleNamespace).Get(context.Background(), exampleName, metav1.GetOptions{})
		require.NoError(t, err)
		assert.Equal(t, exampleReplicaSet, replicaSet)
	})

	t.Run("NoTypeMetaReturned", func(t *testing.T) {
		exampleName := "exampleReplicaSet"
		exampleNamespace := "exampleNamespace"
		returnedReplicaSet := fixReplicaSetWithoutTypeMeta(exampleName, exampleNamespace, nil)
		expectedReplicaSet := fixReplicaSet(exampleName, exampleNamespace, nil)
		replicaSetInformer, client := fixReplicaSetInformer(returnedReplicaSet)
		svc := k8s.NewReplicaSetService(replicaSetInformer, client)

		update := expectedReplicaSet.DeepCopy()
		update.Labels = map[string]string{
			"example": "example",
		}

		replicaSet, err := svc.Update(exampleName, exampleNamespace, *update)
		require.NoError(t, err)
		assert.Equal(t, update, replicaSet)

		replicaSet, err = client.ReplicaSets(exampleNamespace).Get(context.Background(), exampleName, metav1.GetOptions{})
		require.NoError(t, err)
		assert.Equal(t, update, replicaSet)
	})
}

func TestReplicaSetService_Delete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		exampleName := "exampleReplicaSet"
		exampleNamespace := "exampleNamespace"
		exampleReplicaSet := fixReplicaSet(exampleName, exampleNamespace, nil)
		replicaSetInformer, client := fixReplicaSetInformer(exampleReplicaSet)
		svc := k8s.NewReplicaSetService(replicaSetInformer, client)

		err := svc.Delete(exampleName, exampleNamespace)

		require.NoError(t, err)
		_, err = client.ReplicaSets(exampleNamespace).Get(context.Background(), exampleName, metav1.GetOptions{})
		assert.True(t, errors.IsNotFound(err))
	})

	t.Run("Delete Error", func(t *testing.T) {
		exampleName := "exampleReplicaSet"
		exampleNamespace := "exampleNamespace"
		exampleReplicaSet := fixReplicaSet(exampleName, exampleNamespace, nil)
		replicaSetInformer, client := fixFailingReplicaSetInformer(exampleReplicaSet)
		svc := k8s.NewReplicaSetService(replicaSetInformer, client)

		err := svc.Delete(exampleName, exampleNamespace)

		// then
		require.Error(t, err)
	})
}

func fixReplicaSetInformer(objects ...runtime.Object) (cache.SharedIndexInformer, appsv1.AppsV1Interface) {
	client := fake.NewSimpleClientset(objects...)
	informerFactory := informers.NewSharedInformerFactory(client, 0)
	return informerFactory.Apps().V1().ReplicaSets().Informer(), client.AppsV1()
}

func fixFailingReplicaSetInformer(objects ...runtime.Object) (cache.SharedIndexInformer, appsv1.AppsV1Interface) {
	client := fake.NewSimpleClientset(objects...)
	client.PrependReactor("update", "replicasets", failingReactor)
	client.PrependReactor("delete", "replicasets", failingReactor)
	informerFactory := informers.NewSharedInformerFactory(client, 0)
	return informerFactory.Apps().V1().ReplicaSets().Informer(), client.AppsV1()
}

func fixReplicaSet(name, namespace string, labels map[string]string) *apps.ReplicaSet {
	replicaSet := fixReplicaSetWithoutTypeMeta(name, namespace, labels)
	replicaSet.TypeMeta = metav1.TypeMeta{
		Kind:       "ReplicaSet",
		APIVersion: "apps/v1",
	}
	return replicaSet
}

func fixReplicaSetWithoutTypeMeta(name, namespace string, labels map[string]string) *apps.ReplicaSet {
	return &apps.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
	}
}
