package listener_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/cms/listener"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/cms/listener/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
)

func TestClusterDocsTopic_OnAdd(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlClusterDocsTopic := new(gqlschema.ClusterDocsTopic)
		clusterDocsTopic := new(v1alpha1.ClusterDocsTopic)
		converter := automock.NewGQLClusterDocsTopicConverter()

		channel := make(chan gqlschema.ClusterDocsTopicEvent, 1)
		defer close(channel)
		converter.On("ToGQL", clusterDocsTopic).Return(gqlClusterDocsTopic, nil).Once()
		defer converter.AssertExpectations(t)
		clusterDocsTopicListener := listener.NewClusterDocsTopic(channel, filterClusterDocsTopicTrue, converter)

		// when
		clusterDocsTopicListener.OnAdd(clusterDocsTopic)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeAdd, result.Type)
		assert.Equal(t, *gqlClusterDocsTopic, result.ClusterDocsTopic)
	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		clusterDocsTopicListener := listener.NewClusterDocsTopic(nil, filterClusterDocsTopicFalse, nil)

		// when
		clusterDocsTopicListener.OnAdd(new(v1alpha1.ClusterDocsTopic))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		clusterDocsTopicListener := listener.NewClusterDocsTopic(nil, filterClusterDocsTopicTrue, nil)

		// when
		clusterDocsTopicListener.OnAdd(nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		clusterDocsTopic := new(v1alpha1.ClusterDocsTopic)
		converter := automock.NewGQLClusterDocsTopicConverter()

		converter.On("ToGQL", clusterDocsTopic).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		clusterDocsTopicListener := listener.NewClusterDocsTopic(nil, filterClusterDocsTopicTrue, converter)

		// when
		clusterDocsTopicListener.OnAdd(clusterDocsTopic)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		clusterDocsTopicListener := listener.NewClusterDocsTopic(nil, filterClusterDocsTopicTrue, nil)

		// when
		clusterDocsTopicListener.OnAdd(new(struct{}))
	})
}

func TestClusterDocsTopic_OnDelete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlClusterDocsTopic := new(gqlschema.ClusterDocsTopic)
		clusterDocsTopic := new(v1alpha1.ClusterDocsTopic)
		converter := automock.NewGQLClusterDocsTopicConverter()

		channel := make(chan gqlschema.ClusterDocsTopicEvent, 1)
		defer close(channel)
		converter.On("ToGQL", clusterDocsTopic).Return(gqlClusterDocsTopic, nil).Once()
		defer converter.AssertExpectations(t)
		clusterDocsTopicListener := listener.NewClusterDocsTopic(channel, filterClusterDocsTopicTrue, converter)

		// when
		clusterDocsTopicListener.OnDelete(clusterDocsTopic)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeDelete, result.Type)
		assert.Equal(t, *gqlClusterDocsTopic, result.ClusterDocsTopic)

	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		clusterDocsTopicListener := listener.NewClusterDocsTopic(nil, filterClusterDocsTopicFalse, nil)

		// when
		clusterDocsTopicListener.OnDelete(new(v1alpha1.ClusterDocsTopic))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		clusterDocsTopicListener := listener.NewClusterDocsTopic(nil, filterClusterDocsTopicTrue, nil)

		// when
		clusterDocsTopicListener.OnDelete(nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		clusterDocsTopic := new(v1alpha1.ClusterDocsTopic)
		converter := automock.NewGQLClusterDocsTopicConverter()

		converter.On("ToGQL", clusterDocsTopic).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		clusterDocsTopicListener := listener.NewClusterDocsTopic(nil, filterClusterDocsTopicTrue, converter)

		// when
		clusterDocsTopicListener.OnDelete(clusterDocsTopic)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		clusterDocsTopicListener := listener.NewClusterDocsTopic(nil, filterClusterDocsTopicTrue, nil)

		// when
		clusterDocsTopicListener.OnDelete(new(struct{}))
	})
}

func TestClusterDocsTopic_OnUpdate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlClusterDocsTopic := new(gqlschema.ClusterDocsTopic)
		clusterDocsTopic := new(v1alpha1.ClusterDocsTopic)
		converter := automock.NewGQLClusterDocsTopicConverter()

		channel := make(chan gqlschema.ClusterDocsTopicEvent, 1)
		defer close(channel)
		converter.On("ToGQL", clusterDocsTopic).Return(gqlClusterDocsTopic, nil).Once()
		defer converter.AssertExpectations(t)
		clusterDocsTopicListener := listener.NewClusterDocsTopic(channel, filterClusterDocsTopicTrue, converter)

		// when
		clusterDocsTopicListener.OnUpdate(clusterDocsTopic, clusterDocsTopic)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeUpdate, result.Type)
		assert.Equal(t, *gqlClusterDocsTopic, result.ClusterDocsTopic)

	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		clusterDocsTopicListener := listener.NewClusterDocsTopic(nil, filterClusterDocsTopicFalse, nil)

		// when
		clusterDocsTopicListener.OnUpdate(new(v1alpha1.ClusterDocsTopic), new(v1alpha1.ClusterDocsTopic))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		clusterDocsTopicListener := listener.NewClusterDocsTopic(nil, filterClusterDocsTopicTrue, nil)

		// when
		clusterDocsTopicListener.OnUpdate(nil, nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		clusterDocsTopic := new(v1alpha1.ClusterDocsTopic)
		converter := automock.NewGQLClusterDocsTopicConverter()

		converter.On("ToGQL", clusterDocsTopic).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		clusterDocsTopicListener := listener.NewClusterDocsTopic(nil, filterClusterDocsTopicTrue, converter)

		// when
		clusterDocsTopicListener.OnUpdate(nil, clusterDocsTopic)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		clusterDocsTopicListener := listener.NewClusterDocsTopic(nil, filterClusterDocsTopicTrue, nil)

		// when
		clusterDocsTopicListener.OnUpdate(new(struct{}), new(struct{}))
	})
}

func filterClusterDocsTopicTrue(o *v1alpha1.ClusterDocsTopic) bool {
	return true
}

func filterClusterDocsTopicFalse(o *v1alpha1.ClusterDocsTopic) bool {
	return false
}
