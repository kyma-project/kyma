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

const (
	AssetGroupNamespace = "AssetGroupNamespace"
)

func TestAssetGroupService_Find(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		assetGroup1 := fixSimpleAssetGroup("1", nil)
		assetGroup2 := fixSimpleAssetGroup("2", nil)
		assetGroup3 := fixSimpleAssetGroup("3", nil)

		service := fixFakeAssetGroupService(t, assetGroup1, assetGroup2, assetGroup3)

		result, err := service.Find(AssetGroupNamespace, "1")
		require.NoError(t, err)
		assert.Equal(t, assetGroup1, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		service := fixFakeAssetGroupService(t)

		result, err := service.Find(AssetGroupNamespace, "1")
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestAssetGroupService_Subscribe(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		service := fixFakeAssetGroupService(t)

		assetGroupListener := listener.NewAssetGroup(nil, nil, nil)
		service.Subscribe(assetGroupListener)
	})

	t.Run("Duplicated", func(t *testing.T) {
		service := fixFakeAssetGroupService(t)

		assetGroupListener := listener.NewAssetGroup(nil, nil, nil)
		service.Subscribe(assetGroupListener)
		service.Subscribe(assetGroupListener)
	})

	t.Run("Multiple", func(t *testing.T) {
		service := fixFakeAssetGroupService(t)

		assetGroupListenerA := listener.NewAssetGroup(nil, nil, nil)
		assetGroupListenerB := listener.NewAssetGroup(nil, nil, nil)

		service.Subscribe(assetGroupListenerA)
		service.Subscribe(assetGroupListenerB)
	})

	t.Run("Nil", func(t *testing.T) {
		service := fixFakeAssetGroupService(t)

		service.Subscribe(nil)
	})
}

func TestAssetGroupService_Unsubscribe(t *testing.T) {
	t.Run("Existing", func(t *testing.T) {
		service := fixFakeAssetGroupService(t)

		assetGroupListener := listener.NewAssetGroup(nil, nil, nil)
		service.Subscribe(assetGroupListener)
		service.Unsubscribe(assetGroupListener)
	})

	t.Run("Duplicated", func(t *testing.T) {
		service := fixFakeAssetGroupService(t)

		assetGroupListener := listener.NewAssetGroup(nil, nil, nil)
		service.Subscribe(assetGroupListener)
		service.Subscribe(assetGroupListener)

		service.Unsubscribe(assetGroupListener)
	})

	t.Run("Multiple", func(t *testing.T) {
		service := fixFakeAssetGroupService(t)

		assetGroupListenerA := listener.NewAssetGroup(nil, nil, nil)
		assetGroupListenerB := listener.NewAssetGroup(nil, nil, nil)

		service.Subscribe(assetGroupListenerA)
		service.Subscribe(assetGroupListenerB)

		service.Unsubscribe(assetGroupListenerA)
	})

	t.Run("Nil", func(t *testing.T) {
		service := fixFakeAssetGroupService(t)

		service.Unsubscribe(nil)
	})
}

func fixFakeAssetGroupService(t *testing.T, objects ...runtime.Object) *rafter.AssetGroupService {
	serviceFactory, err := resourceFake.NewFakeServiceFactory(v1beta1.AddToScheme, objects...)
	require.NoError(t, err)

	service, err := rafter.NewAssetGroupService(serviceFactory)
	require.NoError(t, err)

	testingUtils.WaitForInformerStartAtMost(t, time.Second, service.Informer)

	return service
}

func fixSimpleAssetGroup(name string, labels map[string]string) *v1beta1.AssetGroup {
	return &v1beta1.AssetGroup{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AssetGroup",
			APIVersion: v1beta1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: AssetGroupNamespace,
			Labels:    labels,
		},
	}
}
