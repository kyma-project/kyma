package listener_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter/listener"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter/listener/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
	"github.com/stretchr/testify/assert"
)

func TestAsset_OnAdd(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlAsset := new(gqlschema.Asset)
		asset := new(v1beta1.Asset)
		converter := automock.NewGqlAssetConverter()

		channel := make(chan gqlschema.AssetEvent, 1)
		defer close(channel)
		converter.On("ToGQL", asset).Return(gqlAsset, nil).Once()
		defer converter.AssertExpectations(t)
		assetListener := listener.NewAsset(channel, filterAssetTrue, converter)

		// when
		assetListener.OnAdd(asset)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeAdd, result.Type)
		assert.Equal(t, *gqlAsset, result.Asset)
	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		assetListener := listener.NewAsset(nil, filterAssetFalse, nil)

		// when
		assetListener.OnAdd(new(v1beta1.Asset))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		assetListener := listener.NewAsset(nil, filterAssetTrue, nil)

		// when
		assetListener.OnAdd(nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		asset := new(v1beta1.Asset)
		converter := automock.NewGqlAssetConverter()

		converter.On("ToGQL", asset).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		assetListener := listener.NewAsset(nil, filterAssetTrue, converter)

		// when
		assetListener.OnAdd(asset)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		assetListener := listener.NewAsset(nil, filterAssetTrue, nil)

		// when
		assetListener.OnAdd(new(struct{}))
	})
}

func TestAsset_OnDelete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlAsset := new(gqlschema.Asset)
		asset := new(v1beta1.Asset)
		converter := automock.NewGqlAssetConverter()

		channel := make(chan gqlschema.AssetEvent, 1)
		defer close(channel)
		converter.On("ToGQL", asset).Return(gqlAsset, nil).Once()
		defer converter.AssertExpectations(t)
		assetListener := listener.NewAsset(channel, filterAssetTrue, converter)

		// when
		assetListener.OnDelete(asset)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeDelete, result.Type)
		assert.Equal(t, *gqlAsset, result.Asset)

	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		assetListener := listener.NewAsset(nil, filterAssetFalse, nil)

		// when
		assetListener.OnDelete(new(v1beta1.Asset))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		assetListener := listener.NewAsset(nil, filterAssetTrue, nil)

		// when
		assetListener.OnDelete(nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		asset := new(v1beta1.Asset)
		converter := automock.NewGqlAssetConverter()

		converter.On("ToGQL", asset).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		assetListener := listener.NewAsset(nil, filterAssetTrue, converter)

		// when
		assetListener.OnDelete(asset)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		assetListener := listener.NewAsset(nil, filterAssetTrue, nil)

		// when
		assetListener.OnDelete(new(struct{}))
	})
}

func TestAsset_OnUpdate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlAsset := new(gqlschema.Asset)
		asset := new(v1beta1.Asset)
		converter := automock.NewGqlAssetConverter()

		channel := make(chan gqlschema.AssetEvent, 1)
		defer close(channel)
		converter.On("ToGQL", asset).Return(gqlAsset, nil).Once()
		defer converter.AssertExpectations(t)
		assetListener := listener.NewAsset(channel, filterAssetTrue, converter)

		// when
		assetListener.OnUpdate(asset, asset)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeUpdate, result.Type)
		assert.Equal(t, *gqlAsset, result.Asset)

	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		assetListener := listener.NewAsset(nil, filterAssetFalse, nil)

		// when
		assetListener.OnUpdate(new(v1beta1.Asset), new(v1beta1.Asset))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		assetListener := listener.NewAsset(nil, filterAssetTrue, nil)

		// when
		assetListener.OnUpdate(nil, nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		asset := new(v1beta1.Asset)
		converter := automock.NewGqlAssetConverter()

		converter.On("ToGQL", asset).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		assetListener := listener.NewAsset(nil, filterAssetTrue, converter)

		// when
		assetListener.OnUpdate(nil, asset)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		assetListener := listener.NewAsset(nil, filterAssetTrue, nil)

		// when
		assetListener.OnUpdate(new(struct{}), new(struct{}))
	})
}

func filterAssetTrue(o *v1beta1.Asset) bool {
	return true
}

func filterAssetFalse(o *v1beta1.Asset) bool {
	return false
}
