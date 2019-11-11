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
	"testing"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

const (
	AssetNamespace = "AssetNamespace"
)

func TestAssetService_Find(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		asset1 := fixSimpleAsset("1", nil)
		asset2 := fixSimpleAsset("2", nil)
		asset3 := fixSimpleAsset("3", nil)

		service := fixFakeAssetService(t, asset1, asset2, asset3)

		result, err := service.Find(AssetNamespace, "1")
		require.NoError(t, err)
		assert.Equal(t, asset1, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		service := fixFakeAssetService(t)

		result, err := service.Find(AssetNamespace, "1")
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestAssetService_ListForAssetGroupByType(t *testing.T) {
	t.Run("Success with assetGroupName paramater", func(t *testing.T) {
		asset1 := fixSimpleAsset("1", map[string]string{
			rafter.AssetGroupLabel: "exampleAssetGroupA",
		})
		asset2 := fixSimpleAsset("2", map[string]string{
			rafter.AssetGroupLabel: "exampleAssetGroupB",
		})
		asset3 := fixSimpleAsset("3", map[string]string{
			rafter.AssetGroupLabel: "exampleAssetGroupC",
		})
		expected := []*v1beta1.Asset{asset1,}

		service := fixFakeAssetService(t, asset1, asset2, asset3)

		result, err := service.ListForAssetGroupByType(AssetNamespace, "exampleAssetGroupA", nil)
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Success with whole paramaters", func(t *testing.T) {
		asset1 := fixSimpleAsset("1", map[string]string{
			rafter.AssetGroupLabel: "exampleAssetGroupA",
			rafter.TypeLabel: "markdown",
		})
		asset2 := fixSimpleAsset("2", map[string]string{
			rafter.AssetGroupLabel: "exampleAssetGroupB",
			rafter.TypeLabel: "json",
		})
		asset3 := fixSimpleAsset("3", map[string]string{
			rafter.AssetGroupLabel: "exampleAssetGroupC",
			rafter.TypeLabel: "yaml",
		})
		expected := []*v1beta1.Asset{asset1,}

		service := fixFakeAssetService(t, asset1, asset2, asset3)

		result, err := service.ListForAssetGroupByType(AssetNamespace, "exampleAssetGroupA", []string{"markdown"})
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		service := fixFakeAssetService(t)

		result, err := service.ListForAssetGroupByType(AssetNamespace, "exampleDocsTopic", nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestAssetService_Subscribe(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		service := fixFakeAssetService(t)

		assetListener := listener.NewAsset(nil, nil, nil)
		service.Subscribe(assetListener)
	})

	t.Run("Duplicated", func(t *testing.T) {
		service := fixFakeAssetService(t)

		assetListener := listener.NewAsset(nil, nil, nil)
		service.Subscribe(assetListener)
		service.Subscribe(assetListener)
	})

	t.Run("Multiple", func(t *testing.T) {
		service := fixFakeAssetService(t)

		assetListenerA := listener.NewAsset(nil, nil, nil)
		assetListenerB := listener.NewAsset(nil, nil, nil)

		service.Subscribe(assetListenerA)
		service.Subscribe(assetListenerB)
	})

	t.Run("Nil", func(t *testing.T) {
		service := fixFakeAssetService(t)

		service.Subscribe(nil)
	})
}

func TestAssetService_Unsubscribe(t *testing.T) {
	t.Run("Existing", func(t *testing.T) {
		service := fixFakeAssetService(t)

		assetListener := listener.NewAsset(nil, nil, nil)
		service.Subscribe(assetListener)
		service.Unsubscribe(assetListener)
	})

	t.Run("Duplicated", func(t *testing.T) {
		service := fixFakeAssetService(t)

		assetListener := listener.NewAsset(nil, nil, nil)
		service.Subscribe(assetListener)
		service.Subscribe(assetListener)

		service.Unsubscribe(assetListener)
	})

	t.Run("Multiple", func(t *testing.T) {
		service := fixFakeAssetService(t)

		assetListenerA := listener.NewAsset(nil, nil, nil)
		assetListenerB := listener.NewAsset(nil, nil, nil)

		service.Subscribe(assetListenerA)
		service.Subscribe(assetListenerB)

		service.Unsubscribe(assetListenerA)
	})

	t.Run("Nil", func(t *testing.T) {
		service := fixFakeAssetService(t)

		service.Unsubscribe(nil)
	})
}

func fixFakeAssetService(t *testing.T, objects ...runtime.Object) *rafter.AssetService {
	serviceFactory, err := resourceFake.NewFakeServiceFactory(v1beta1.AddToScheme, objects...)
	require.NoError(t, err)

	service, err := rafter.NewAssetService(serviceFactory)
	require.NoError(t, err)

	testingUtils.WaitForInformerStartAtMost(t, time.Second, service.Informer)

	return service
}

func fixSimpleAsset(name string, labels map[string]string) *v1beta1.Asset {
	return &v1beta1.Asset{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Asset",
			APIVersion: v1beta1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: AssetNamespace,
			Labels:    labels,
		},
	}
}