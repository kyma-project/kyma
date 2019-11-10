package listener_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter/listener"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter/listener/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
	"github.com/stretchr/testify/assert"
)

func TestAssetGroup_OnAdd(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlAssetGroup := new(gqlschema.AssetGroup)
		assetGroup := new(v1beta1.AssetGroup)
		converter := automock.NewGqlAssetGroupConverter()

		channel := make(chan gqlschema.AssetGroupEvent, 1)
		defer close(channel)
		converter.On("ToGQL", assetGroup).Return(gqlAssetGroup, nil).Once()
		defer converter.AssertExpectations(t)
		assetGroupListener := listener.NewAssetGroup(channel, filterAssetGroupTrue, converter)

		// when
		assetGroupListener.OnAdd(assetGroup)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeAdd, result.Type)
		assert.Equal(t, *gqlAssetGroup, result.AssetGroup)
	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		assetGroupListener := listener.NewAssetGroup(nil, filterAssetGroupFalse, nil)

		// when
		assetGroupListener.OnAdd(new(v1beta1.AssetGroup))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		assetGroupListener := listener.NewAssetGroup(nil, filterAssetGroupTrue, nil)

		// when
		assetGroupListener.OnAdd(nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		assetGroup := new(v1beta1.AssetGroup)
		converter := automock.NewGqlAssetGroupConverter()

		converter.On("ToGQL", assetGroup).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		assetGroupListener := listener.NewAssetGroup(nil, filterAssetGroupTrue, converter)

		// when
		assetGroupListener.OnAdd(assetGroup)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		assetGroupListener := listener.NewAssetGroup(nil, filterAssetGroupTrue, nil)

		// when
		assetGroupListener.OnAdd(new(struct{}))
	})
}

func TestAssetGroup_OnDelete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlAssetGroup := new(gqlschema.AssetGroup)
		assetGroup := new(v1beta1.AssetGroup)
		converter := automock.NewGqlAssetGroupConverter()

		channel := make(chan gqlschema.AssetGroupEvent, 1)
		defer close(channel)
		converter.On("ToGQL", assetGroup).Return(gqlAssetGroup, nil).Once()
		defer converter.AssertExpectations(t)
		assetGroupListener := listener.NewAssetGroup(channel, filterAssetGroupTrue, converter)

		// when
		assetGroupListener.OnDelete(assetGroup)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeDelete, result.Type)
		assert.Equal(t, *gqlAssetGroup, result.AssetGroup)

	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		assetGroupListener := listener.NewAssetGroup(nil, filterAssetGroupFalse, nil)

		// when
		assetGroupListener.OnDelete(new(v1beta1.AssetGroup))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		assetGroupListener := listener.NewAssetGroup(nil, filterAssetGroupTrue, nil)

		// when
		assetGroupListener.OnDelete(nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		assetGroup := new(v1beta1.AssetGroup)
		converter := automock.NewGqlAssetGroupConverter()

		converter.On("ToGQL", assetGroup).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		assetGroupListener := listener.NewAssetGroup(nil, filterAssetGroupTrue, converter)

		// when
		assetGroupListener.OnDelete(assetGroup)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		assetGroupListener := listener.NewAssetGroup(nil, filterAssetGroupTrue, nil)

		// when
		assetGroupListener.OnDelete(new(struct{}))
	})
}

func TestAssetGroup_OnUpdate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlAssetGroup := new(gqlschema.AssetGroup)
		assetGroup := new(v1beta1.AssetGroup)
		converter := automock.NewGqlAssetGroupConverter()

		channel := make(chan gqlschema.AssetGroupEvent, 1)
		defer close(channel)
		converter.On("ToGQL", assetGroup).Return(gqlAssetGroup, nil).Once()
		defer converter.AssertExpectations(t)
		assetGroupListener := listener.NewAssetGroup(channel, filterAssetGroupTrue, converter)

		// when
		assetGroupListener.OnUpdate(assetGroup, assetGroup)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeUpdate, result.Type)
		assert.Equal(t, *gqlAssetGroup, result.AssetGroup)

	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		assetGroupListener := listener.NewAssetGroup(nil, filterAssetGroupFalse, nil)

		// when
		assetGroupListener.OnUpdate(new(v1beta1.AssetGroup), new(v1beta1.AssetGroup))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		assetGroupListener := listener.NewAssetGroup(nil, filterAssetGroupTrue, nil)

		// when
		assetGroupListener.OnUpdate(nil, nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		assetGroup := new(v1beta1.AssetGroup)
		converter := automock.NewGqlAssetGroupConverter()

		converter.On("ToGQL", assetGroup).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		assetGroupListener := listener.NewAssetGroup(nil, filterAssetGroupTrue, converter)

		// when
		assetGroupListener.OnUpdate(nil, assetGroup)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		assetGroupListener := listener.NewAssetGroup(nil, filterAssetGroupTrue, nil)

		// when
		assetGroupListener.OnUpdate(new(struct{}), new(struct{}))
	})
}

func filterAssetGroupTrue(o *v1beta1.AssetGroup) bool {
	return true
}

func filterAssetGroupFalse(o *v1beta1.AssetGroup) bool {
	return false
}
