package rafter_test

import (
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter/listener"
	resourceFake "github.com/kyma-project/kyma/components/console-backend-service/internal/resource/fake"
	testingUtils "github.com/kyma-project/kyma/components/console-backend-service/internal/testing"
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestClusterAssetGroupService_Find(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		clusterAssetGroup1 := fixSimpleClusterAssetGroup("1", nil)
		clusterAssetGroup2 := fixSimpleClusterAssetGroup("2", nil)
		clusterAssetGroup3 := fixSimpleClusterAssetGroup("3", nil)

		service := fixFakeClusterAssetGroupService(t, clusterAssetGroup1, clusterAssetGroup2, clusterAssetGroup3)

		result, err := service.Find("1")
		require.NoError(t, err)
		assert.Equal(t, clusterAssetGroup1, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		service := fixFakeClusterAssetGroupService(t)

		result, err := service.Find("1")
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestClusterAssetGroupService_List(t *testing.T) {
	t.Run("Success without parameters", func(t *testing.T) {
		clusterAssetGroup1 := fixSimpleClusterAssetGroup("1", map[string]string{
			rafter.OrderLabel: "1",
		})
		clusterAssetGroup2 := fixSimpleClusterAssetGroup("2", map[string]string{
			rafter.OrderLabel: "2",
		})
		clusterAssetGroup3 := fixSimpleClusterAssetGroup("3", map[string]string{
			rafter.OrderLabel: "3",
		})
		expected := []*v1beta1.ClusterAssetGroup{
			clusterAssetGroup1,
			clusterAssetGroup2,
			clusterAssetGroup3,
		}

		service := fixFakeClusterAssetGroupService(t, clusterAssetGroup1, clusterAssetGroup2, clusterAssetGroup3)

		result, err := service.List(nil, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Success with all parameters", func(t *testing.T) {
		viewContext := "viewContext"
		groupName := "groupName"
		clusterAssetGroup1 := fixSimpleClusterAssetGroup("1", map[string]string{
			rafter.ViewContextLabel: viewContext,
			rafter.GroupNameLabel:   groupName,
		})
		clusterAssetGroup2 := fixSimpleClusterAssetGroup("2", map[string]string{
			rafter.ViewContextLabel: "viewContext2",
			rafter.GroupNameLabel:   "groupName2",
		})
		clusterAssetGroup3 := fixSimpleClusterAssetGroup("3", map[string]string{
			rafter.ViewContextLabel: "viewContext3",
			rafter.GroupNameLabel:   "groupName3",
		})
		expected := []*v1beta1.ClusterAssetGroup{
			clusterAssetGroup1,
		}

		service := fixFakeClusterAssetGroupService(t, clusterAssetGroup1, clusterAssetGroup2, clusterAssetGroup3)

		result, err := service.List(&viewContext, &groupName)
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Success with viewContext parameter", func(t *testing.T) {
		viewContext := "viewContext"
		clusterAssetGroup1 := fixSimpleClusterAssetGroup("1", map[string]string{
			rafter.ViewContextLabel: viewContext,
		})
		clusterAssetGroup2 := fixSimpleClusterAssetGroup("2", map[string]string{
			rafter.ViewContextLabel: "viewContext2",
		})
		clusterAssetGroup3 := fixSimpleClusterAssetGroup("3", map[string]string{
			rafter.ViewContextLabel: "viewContext3",
		})
		expected := []*v1beta1.ClusterAssetGroup{
			clusterAssetGroup1,
		}

		service := fixFakeClusterAssetGroupService(t, clusterAssetGroup1, clusterAssetGroup2, clusterAssetGroup3)

		result, err := service.List(&viewContext, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Success with groupName parameter", func(t *testing.T) {
		groupName := "groupName"
		clusterAssetGroup1 := fixSimpleClusterAssetGroup("1", map[string]string{
			rafter.GroupNameLabel: groupName,
		})
		clusterAssetGroup2 := fixSimpleClusterAssetGroup("2", map[string]string{
			rafter.GroupNameLabel: "groupName2",
		})
		clusterAssetGroup3 := fixSimpleClusterAssetGroup("3", map[string]string{
			rafter.GroupNameLabel: "groupName3",
		})
		expected := []*v1beta1.ClusterAssetGroup{
			clusterAssetGroup1,
		}

		service := fixFakeClusterAssetGroupService(t, clusterAssetGroup1, clusterAssetGroup2, clusterAssetGroup3)

		result, err := service.List(nil, &groupName)
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Success with sorting function", func(t *testing.T) {
		clusterAssetGroup1 := fixSimpleClusterAssetGroup("1", map[string]string{
			rafter.OrderLabel: "1",
		})
		clusterAssetGroup2 := fixSimpleClusterAssetGroup("2", map[string]string{
			rafter.OrderLabel: "2",
		})
		clusterAssetGroup3 := fixSimpleClusterAssetGroup("3", map[string]string{
			rafter.OrderLabel: "3",
		})
		clusterAssetGroup4 := fixSimpleClusterAssetGroup("4", nil)
		expected := []*v1beta1.ClusterAssetGroup{
			clusterAssetGroup1,
			clusterAssetGroup2,
			clusterAssetGroup3,
			clusterAssetGroup4,
		}

		service := fixFakeClusterAssetGroupService(t, clusterAssetGroup2, clusterAssetGroup1, clusterAssetGroup4, clusterAssetGroup3)

		result, err := service.List(nil, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		service := fixFakeClusterAssetGroupService(t)

		result, err := service.List(nil, nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestClusterAssetGroupService_Subscribe(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		service := fixFakeClusterAssetGroupService(t)

		clusterAssetGroupListener := listener.NewClusterAssetGroup(nil, nil, nil)
		service.Subscribe(clusterAssetGroupListener)
	})

	t.Run("Duplicated", func(t *testing.T) {
		service := fixFakeClusterAssetGroupService(t)

		clusterAssetGroupListener := listener.NewClusterAssetGroup(nil, nil, nil)
		service.Subscribe(clusterAssetGroupListener)
		service.Subscribe(clusterAssetGroupListener)
	})

	t.Run("Multiple", func(t *testing.T) {
		service := fixFakeClusterAssetGroupService(t)

		clusterAssetGroupListenerA := listener.NewClusterAssetGroup(nil, nil, nil)
		clusterAssetGroupListenerB := listener.NewClusterAssetGroup(nil, nil, nil)

		service.Subscribe(clusterAssetGroupListenerA)
		service.Subscribe(clusterAssetGroupListenerB)
	})

	t.Run("Nil", func(t *testing.T) {
		service := fixFakeClusterAssetGroupService(t)

		service.Subscribe(nil)
	})
}

func TestClusterAssetGroupService_Unsubscribe(t *testing.T) {
	t.Run("Existing", func(t *testing.T) {
		service := fixFakeClusterAssetGroupService(t)

		clusterAssetGroupListener := listener.NewClusterAssetGroup(nil, nil, nil)
		service.Subscribe(clusterAssetGroupListener)
		service.Unsubscribe(clusterAssetGroupListener)
	})

	t.Run("Duplicated", func(t *testing.T) {
		service := fixFakeClusterAssetGroupService(t)

		clusterAssetGroupListener := listener.NewClusterAssetGroup(nil, nil, nil)
		service.Subscribe(clusterAssetGroupListener)
		service.Subscribe(clusterAssetGroupListener)

		service.Unsubscribe(clusterAssetGroupListener)
	})

	t.Run("Multiple", func(t *testing.T) {
		service := fixFakeClusterAssetGroupService(t)

		clusterAssetGroupListenerA := listener.NewClusterAssetGroup(nil, nil, nil)
		clusterAssetGroupListenerB := listener.NewClusterAssetGroup(nil, nil, nil)

		service.Subscribe(clusterAssetGroupListenerA)
		service.Subscribe(clusterAssetGroupListenerB)

		service.Unsubscribe(clusterAssetGroupListenerA)
	})

	t.Run("Nil", func(t *testing.T) {
		service := fixFakeClusterAssetGroupService(t)

		service.Unsubscribe(nil)
	})
}

func fixFakeClusterAssetGroupService(t *testing.T, objects ...runtime.Object) *rafter.ClusterAssetGroupService {
	serviceFactory, err := resourceFake.NewFakeServiceFactory(v1beta1.AddToScheme, objects...)
	require.NoError(t, err)

	service, err := rafter.NewClusterAssetGroupService(serviceFactory)
	require.NoError(t, err)

	testingUtils.WaitForInformerStartAtMost(t, time.Second, service.Informer)

	return service
}

func fixSimpleClusterAssetGroup(name string, labels map[string]string) *v1beta1.ClusterAssetGroup {
	return &v1beta1.ClusterAssetGroup{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterAssetGroup",
			APIVersion: v1beta1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
	}
}
