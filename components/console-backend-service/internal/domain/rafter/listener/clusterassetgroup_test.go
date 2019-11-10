package listener_test

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter/listener"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter/listener/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestClusterAssetGroup_OnAdd(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlClusterAssetGroup := new(gqlschema.ClusterAssetGroup)
		clusterAssetGroup := new(v1beta1.ClusterAssetGroup)
		converter := automock.NewGqlClusterAssetGroupConverter()

		channel := make(chan gqlschema.ClusterAssetGroupEvent, 1)
		defer close(channel)
		converter.On("ToGQL", clusterAssetGroup).Return(gqlClusterAssetGroup, nil).Once()
		defer converter.AssertExpectations(t)
		clusterAssetGroupListener := listener.NewClusterAssetGroup(channel, filterClusterAssetGroupTrue, converter)

		// when
		clusterAssetGroupListener.OnAdd(clusterAssetGroup)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeAdd, result.Type)
		assert.Equal(t, *gqlClusterAssetGroup, result.ClusterAssetGroup)
	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		clusterAssetGroupListener := listener.NewClusterAssetGroup(nil, filterClusterAssetGroupFalse, nil)

		// when
		clusterAssetGroupListener.OnAdd(new(v1beta1.ClusterAssetGroup))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		clusterAssetGroupListener := listener.NewClusterAssetGroup(nil, filterClusterAssetGroupTrue, nil)

		// when
		clusterAssetGroupListener.OnAdd(nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		clusterAssetGroup := new(v1beta1.ClusterAssetGroup)
		converter := automock.NewGqlClusterAssetGroupConverter()

		converter.On("ToGQL", clusterAssetGroup).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		clusterAssetGroupListener := listener.NewClusterAssetGroup(nil, filterClusterAssetGroupTrue, converter)

		// when
		clusterAssetGroupListener.OnAdd(clusterAssetGroup)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		clusterAssetGroupListener := listener.NewClusterAssetGroup(nil, filterClusterAssetGroupTrue, nil)

		// when
		clusterAssetGroupListener.OnAdd(new(struct{}))
	})
}

func TestClusterAssetGroup_OnDelete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlClusterAssetGroup := new(gqlschema.ClusterAssetGroup)
		clusterAssetGroup := new(v1beta1.ClusterAssetGroup)
		converter := automock.NewGqlClusterAssetGroupConverter()

		channel := make(chan gqlschema.ClusterAssetGroupEvent, 1)
		defer close(channel)
		converter.On("ToGQL", clusterAssetGroup).Return(gqlClusterAssetGroup, nil).Once()
		defer converter.AssertExpectations(t)
		clusterAssetGroupListener := listener.NewClusterAssetGroup(channel, filterClusterAssetGroupTrue, converter)

		// when
		clusterAssetGroupListener.OnDelete(clusterAssetGroup)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeDelete, result.Type)
		assert.Equal(t, *gqlClusterAssetGroup, result.ClusterAssetGroup)

	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		clusterAssetGroupListener := listener.NewClusterAssetGroup(nil, filterClusterAssetGroupFalse, nil)

		// when
		clusterAssetGroupListener.OnDelete(new(v1beta1.ClusterAssetGroup))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		clusterAssetGroupListener := listener.NewClusterAssetGroup(nil, filterClusterAssetGroupTrue, nil)

		// when
		clusterAssetGroupListener.OnDelete(nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		clusterAssetGroup := new(v1beta1.ClusterAssetGroup)
		converter := automock.NewGqlClusterAssetGroupConverter()

		converter.On("ToGQL", clusterAssetGroup).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		clusterAssetGroupListener := listener.NewClusterAssetGroup(nil, filterClusterAssetGroupTrue, converter)

		// when
		clusterAssetGroupListener.OnDelete(clusterAssetGroup)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		clusterAssetGroupListener := listener.NewClusterAssetGroup(nil, filterClusterAssetGroupTrue, nil)

		// when
		clusterAssetGroupListener.OnDelete(new(struct{}))
	})
}

func TestClusterAssetGroup_OnUpdate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlAssetGroup := new(gqlschema.ClusterAssetGroup)
		clusterAssetGroup := new(v1beta1.ClusterAssetGroup)
		converter := automock.NewGqlClusterAssetGroupConverter()

		channel := make(chan gqlschema.ClusterAssetGroupEvent, 1)
		defer close(channel)
		converter.On("ToGQL", clusterAssetGroup).Return(gqlAssetGroup, nil).Once()
		defer converter.AssertExpectations(t)
		assetGroupListener := listener.NewClusterAssetGroup(channel, filterClusterAssetGroupTrue, converter)

		// when
		assetGroupListener.OnUpdate(clusterAssetGroup, clusterAssetGroup)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeUpdate, result.Type)
		assert.Equal(t, *gqlAssetGroup, result.ClusterAssetGroup)

	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		assetGroupListener := listener.NewClusterAssetGroup(nil, filterClusterAssetGroupFalse, nil)

		// when
		assetGroupListener.OnUpdate(new(v1beta1.ClusterAssetGroup), new(v1beta1.ClusterAssetGroup))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		assetGroupListener := listener.NewClusterAssetGroup(nil, filterClusterAssetGroupTrue, nil)

		// when
		assetGroupListener.OnUpdate(nil, nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		clusterAssetGroup := new(v1beta1.ClusterAssetGroup)
		converter := automock.NewGqlClusterAssetGroupConverter()

		converter.On("ToGQL", clusterAssetGroup).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		assetGroupListener := listener.NewClusterAssetGroup(nil, filterClusterAssetGroupTrue, converter)

		// when
		assetGroupListener.OnUpdate(nil, clusterAssetGroup)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		assetGroupListener := listener.NewClusterAssetGroup(nil, filterClusterAssetGroupTrue, nil)

		// when
		assetGroupListener.OnUpdate(new(struct{}), new(struct{}))
	})
}

func filterClusterAssetGroupTrue(o *v1beta1.ClusterAssetGroup) bool {
	return true
}

func filterClusterAssetGroupFalse(o *v1beta1.ClusterAssetGroup) bool {
	return false
}
