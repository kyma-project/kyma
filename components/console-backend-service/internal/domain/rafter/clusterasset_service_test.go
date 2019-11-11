package rafter_test

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter/listener"
	resourceFake "github.com/kyma-project/kyma/components/console-backend-service/internal/resource/fake"
	testingUtils "github.com/kyma-project/kyma/components/console-backend-service/internal/testing"
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

func TestClusterAssetService_Find(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		clusterAsset1 := fixSimpleClusterAsset("1", nil)
		clusterAsset2 := fixSimpleClusterAsset("2", nil)
		clusterAsset3 := fixSimpleClusterAsset("3", nil)

		service := fixFakeClusterAssetService(t, clusterAsset1, clusterAsset2, clusterAsset3)

		result, err := service.Find("1")
		require.NoError(t, err)
		assert.Equal(t, clusterAsset1, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		service := fixFakeClusterAssetService(t)

		result, err := service.Find("1")
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestClusterAssetService_ListForAssetGroupByType(t *testing.T) {
	t.Run("Success with assetGroupName paramater", func(t *testing.T) {
		clusterAsset1 := fixSimpleClusterAsset("1", map[string]string{
			rafter.AssetGroupLabel: "exampleAssetGroupA",
		})
		clusterAsset2 := fixSimpleClusterAsset("2", map[string]string{
			rafter.AssetGroupLabel: "exampleAssetGroupB",
		})
		clusterAsset3 := fixSimpleClusterAsset("3", map[string]string{
			rafter.AssetGroupLabel: "exampleAssetGroupC",
		})
		expected := []*v1beta1.ClusterAsset{clusterAsset1,}

		service := fixFakeClusterAssetService(t, clusterAsset1, clusterAsset2, clusterAsset3)

		result, err := service.ListForClusterAssetGroupByType("exampleAssetGroupA", nil)
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Success with whole paramaters", func(t *testing.T) {
		clusterAsset1 := fixSimpleClusterAsset("1", map[string]string{
			rafter.AssetGroupLabel: "exampleAssetGroupA",
			rafter.TypeLabel: "markdown",
		})
		clusterAsset2 := fixSimpleClusterAsset("2", map[string]string{
			rafter.AssetGroupLabel: "exampleAssetGroupB",
			rafter.TypeLabel: "json",
		})
		clusterAsset3 := fixSimpleClusterAsset("3", map[string]string{
			rafter.AssetGroupLabel: "exampleAssetGroupC",
			rafter.TypeLabel: "yaml",
		})
		expected := []*v1beta1.ClusterAsset{clusterAsset1,}

		service := fixFakeClusterAssetService(t, clusterAsset1, clusterAsset2, clusterAsset3)

		result, err := service.ListForClusterAssetGroupByType( "exampleAssetGroupA", []string{"markdown"})
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		service := fixFakeClusterAssetService(t)

		result, err := service.ListForClusterAssetGroupByType("exampleDocsTopic", nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestClusterAssetService_Subscribe(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		service := fixFakeClusterAssetService(t)

		clusterAssetListener := listener.NewClusterAsset(nil, nil, nil)
		service.Subscribe(clusterAssetListener)
	})

	t.Run("Duplicated", func(t *testing.T) {
		service := fixFakeClusterAssetService(t)

		clusterAssetListener := listener.NewClusterAsset(nil, nil, nil)
		service.Subscribe(clusterAssetListener)
		service.Subscribe(clusterAssetListener)
	})

	t.Run("Multiple", func(t *testing.T) {
		service := fixFakeClusterAssetService(t)

		clusterAssetListenerA := listener.NewClusterAsset(nil, nil, nil)
		clusterAssetListenerB := listener.NewClusterAsset(nil, nil, nil)

		service.Subscribe(clusterAssetListenerA)
		service.Subscribe(clusterAssetListenerB)
	})

	t.Run("Nil", func(t *testing.T) {
		service := fixFakeClusterAssetService(t)

		service.Subscribe(nil)
	})
}

func TestClusterAssetService_Unsubscribe(t *testing.T) {
	t.Run("Existing", func(t *testing.T) {
		service := fixFakeClusterAssetService(t)

		clusterAssetListener := listener.NewClusterAsset(nil, nil, nil)
		service.Subscribe(clusterAssetListener)
		service.Unsubscribe(clusterAssetListener)
	})

	t.Run("Duplicated", func(t *testing.T) {
		service := fixFakeClusterAssetService(t)

		clusterAssetListener := listener.NewClusterAsset(nil, nil, nil)
		service.Subscribe(clusterAssetListener)
		service.Subscribe(clusterAssetListener)

		service.Unsubscribe(clusterAssetListener)
	})

	t.Run("Multiple", func(t *testing.T) {
		service := fixFakeClusterAssetService(t)

		clusterAssetListenerA := listener.NewClusterAsset(nil, nil, nil)
		clusterAssetListenerB := listener.NewClusterAsset(nil, nil, nil)

		service.Subscribe(clusterAssetListenerA)
		service.Subscribe(clusterAssetListenerB)

		service.Unsubscribe(clusterAssetListenerA)
	})

	t.Run("Nil", func(t *testing.T) {
		service := fixFakeClusterAssetService(t)

		service.Unsubscribe(nil)
	})
}

func fixFakeClusterAssetService(t *testing.T, objects ...runtime.Object) *rafter.ClusterAssetService {
	serviceFactory, err := resourceFake.NewFakeServiceFactory(v1beta1.AddToScheme, objects...)
	require.NoError(t, err)

	service, err := rafter.NewClusterAssetService(serviceFactory)
	require.NoError(t, err)

	testingUtils.WaitForInformerStartAtMost(t, time.Second, service.Informer)

	return service
}

func fixSimpleClusterAsset(name string, labels map[string]string) *v1beta1.ClusterAsset {
	return &v1beta1.ClusterAsset{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterAsset",
			APIVersion: v1beta1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
	}
}