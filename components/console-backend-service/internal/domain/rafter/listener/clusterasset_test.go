package listener_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter/listener"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter/listener/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
	"github.com/stretchr/testify/assert"
)

func TestClusterAsset_OnAdd(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlClusterAsset := new(gqlschema.ClusterAsset)
		clusterAsset := new(v1beta1.ClusterAsset)
		converter := automock.NewGqlClusterAssetConverter()

		channel := make(chan gqlschema.ClusterAssetEvent, 1)
		defer close(channel)
		converter.On("ToGQL", clusterAsset).Return(gqlClusterAsset, nil).Once()
		defer converter.AssertExpectations(t)
		clusterAssetListener := listener.NewClusterAsset(channel, filterClusterAssetTrue, converter)

		// when
		clusterAssetListener.OnAdd(clusterAsset)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeAdd, result.Type)
		assert.Equal(t, *gqlClusterAsset, result.ClusterAsset)
	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		clusterAssetListener := listener.NewClusterAsset(nil, filterClusterAssetFalse, nil)

		// when
		clusterAssetListener.OnAdd(new(v1beta1.ClusterAsset))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		clusterAssetListener := listener.NewClusterAsset(nil, filterClusterAssetTrue, nil)

		// when
		clusterAssetListener.OnAdd(nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		clusterAsset := new(v1beta1.ClusterAsset)
		converter := automock.NewGqlClusterAssetConverter()

		converter.On("ToGQL", clusterAsset).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		clusterAssetListener := listener.NewClusterAsset(nil, filterClusterAssetTrue, converter)

		// when
		clusterAssetListener.OnAdd(clusterAsset)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		clusterAssetListener := listener.NewClusterAsset(nil, filterClusterAssetTrue, nil)

		// when
		clusterAssetListener.OnAdd(new(struct{}))
	})
}

func TestClusterAsset_OnDelete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlClusterAsset := new(gqlschema.ClusterAsset)
		clusterAsset := new(v1beta1.ClusterAsset)
		converter := automock.NewGqlClusterAssetConverter()

		channel := make(chan gqlschema.ClusterAssetEvent, 1)
		defer close(channel)
		converter.On("ToGQL", clusterAsset).Return(gqlClusterAsset, nil).Once()
		defer converter.AssertExpectations(t)
		clusterAssetListener := listener.NewClusterAsset(channel, filterClusterAssetTrue, converter)

		// when
		clusterAssetListener.OnDelete(clusterAsset)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeDelete, result.Type)
		assert.Equal(t, *gqlClusterAsset, result.ClusterAsset)

	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		clusterAssetListener := listener.NewClusterAsset(nil, filterClusterAssetFalse, nil)

		// when
		clusterAssetListener.OnDelete(new(v1beta1.ClusterAsset))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		clusterAssetListener := listener.NewClusterAsset(nil, filterClusterAssetTrue, nil)

		// when
		clusterAssetListener.OnDelete(nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		clusterAsset := new(v1beta1.ClusterAsset)
		converter := automock.NewGqlClusterAssetConverter()

		converter.On("ToGQL", clusterAsset).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		clusterAssetListener := listener.NewClusterAsset(nil, filterClusterAssetTrue, converter)

		// when
		clusterAssetListener.OnDelete(clusterAsset)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		clusterAssetListener := listener.NewClusterAsset(nil, filterClusterAssetTrue, nil)

		// when
		clusterAssetListener.OnDelete(new(struct{}))
	})
}

func TestClusterAsset_OnUpdate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlClusterAsset := new(gqlschema.ClusterAsset)
		clusterAsset := new(v1beta1.ClusterAsset)
		converter := automock.NewGqlClusterAssetConverter()

		channel := make(chan gqlschema.ClusterAssetEvent, 1)
		defer close(channel)
		converter.On("ToGQL", clusterAsset).Return(gqlClusterAsset, nil).Once()
		defer converter.AssertExpectations(t)
		clusterAssetListener := listener.NewClusterAsset(channel, filterClusterAssetTrue, converter)

		// when
		clusterAssetListener.OnUpdate(clusterAsset, clusterAsset)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeUpdate, result.Type)
		assert.Equal(t, *gqlClusterAsset, result.ClusterAsset)

	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		clusterAssetListener := listener.NewClusterAsset(nil, filterClusterAssetFalse, nil)

		// when
		clusterAssetListener.OnUpdate(new(v1beta1.ClusterAsset), new(v1beta1.ClusterAsset))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		clusterAssetListener := listener.NewClusterAsset(nil, filterClusterAssetTrue, nil)

		// when
		clusterAssetListener.OnUpdate(nil, nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		clusterAsset := new(v1beta1.ClusterAsset)
		converter := automock.NewGqlClusterAssetConverter()

		converter.On("ToGQL", clusterAsset).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		clusterAssetListener := listener.NewClusterAsset(nil, filterClusterAssetTrue, converter)

		// when
		clusterAssetListener.OnUpdate(nil, clusterAsset)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		clusterAssetListener := listener.NewClusterAsset(nil, filterClusterAssetTrue, nil)

		// when
		clusterAssetListener.OnUpdate(new(struct{}), new(struct{}))
	})
}

func filterClusterAssetTrue(o *v1beta1.ClusterAsset) bool {
	return true
}

func filterClusterAssetFalse(o *v1beta1.ClusterAsset) bool {
	return false
}
