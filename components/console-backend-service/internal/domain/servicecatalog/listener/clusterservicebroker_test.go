package listener_test

import (
	"testing"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalog/listener"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalog/listener/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestClusterServiceBrokerListener_OnAdd(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlClusterServiceBroker := new(gqlschema.ClusterServiceBroker)
		serviceBroker := new(v1beta1.ClusterServiceBroker)
		converter := automock.NewGQLClusterServiceBrokerConverter()

		channel := make(chan gqlschema.ClusterServiceBrokerEvent, 1)
		defer close(channel)
		converter.On("ToGQL", serviceBroker).Return(gqlClusterServiceBroker, nil).Once()
		defer converter.AssertExpectations(t)
		serviceBrokerListener := listener.NewClusterServiceBroker(channel, filterClusterServiceBrokerTrue, converter)

		// when
		serviceBrokerListener.OnAdd(serviceBroker)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeAdd, result.Type)
		assert.Equal(t, *gqlClusterServiceBroker, result.ClusterServiceBroker)
	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		serviceBrokerListener := listener.NewClusterServiceBroker(nil, filterClusterServiceBrokerFalse, nil)

		// when
		serviceBrokerListener.OnAdd(new(v1beta1.ClusterServiceBroker))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		serviceBrokerListener := listener.NewClusterServiceBroker(nil, filterClusterServiceBrokerTrue, nil)

		// when
		serviceBrokerListener.OnAdd(nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		serviceBroker := new(v1beta1.ClusterServiceBroker)
		converter := automock.NewGQLClusterServiceBrokerConverter()

		converter.On("ToGQL", serviceBroker).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		serviceBrokerListener := listener.NewClusterServiceBroker(nil, filterClusterServiceBrokerTrue, converter)

		// when
		serviceBrokerListener.OnAdd(serviceBroker)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		serviceBrokerListener := listener.NewClusterServiceBroker(nil, filterClusterServiceBrokerTrue, nil)

		// when
		serviceBrokerListener.OnAdd(new(struct{}))
	})

	t.Run("Conversion error", func(t *testing.T) {
		// given
		serviceBroker := new(v1beta1.ClusterServiceBroker)
		converter := automock.NewGQLClusterServiceBrokerConverter()

		converter.On("ToGQL", serviceBroker).Return(nil, errors.New("Conversion error")).Once()
		defer converter.AssertExpectations(t)
		serviceBrokerListener := listener.NewClusterServiceBroker(nil, filterClusterServiceBrokerTrue, converter)

		// when
		serviceBrokerListener.OnAdd(serviceBroker)
	})
}

func TestClusterServiceBrokerListener_OnDelete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlClusterServiceBroker := new(gqlschema.ClusterServiceBroker)
		serviceBroker := new(v1beta1.ClusterServiceBroker)
		converter := automock.NewGQLClusterServiceBrokerConverter()

		channel := make(chan gqlschema.ClusterServiceBrokerEvent, 1)
		defer close(channel)
		converter.On("ToGQL", serviceBroker).Return(gqlClusterServiceBroker, nil).Once()
		defer converter.AssertExpectations(t)
		serviceBrokerListener := listener.NewClusterServiceBroker(channel, filterClusterServiceBrokerTrue, converter)

		// when
		serviceBrokerListener.OnDelete(serviceBroker)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeDelete, result.Type)
		assert.Equal(t, *gqlClusterServiceBroker, result.ClusterServiceBroker)

	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		serviceBrokerListener := listener.NewClusterServiceBroker(nil, filterClusterServiceBrokerFalse, nil)

		// when
		serviceBrokerListener.OnDelete(new(v1beta1.ClusterServiceBroker))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		serviceBrokerListener := listener.NewClusterServiceBroker(nil, filterClusterServiceBrokerTrue, nil)

		// when
		serviceBrokerListener.OnDelete(nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		serviceBroker := new(v1beta1.ClusterServiceBroker)
		converter := automock.NewGQLClusterServiceBrokerConverter()

		converter.On("ToGQL", serviceBroker).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		serviceBrokerListener := listener.NewClusterServiceBroker(nil, filterClusterServiceBrokerTrue, converter)

		// when
		serviceBrokerListener.OnDelete(serviceBroker)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		serviceBrokerListener := listener.NewClusterServiceBroker(nil, filterClusterServiceBrokerTrue, nil)

		// when
		serviceBrokerListener.OnDelete(new(struct{}))
	})

	t.Run("Conversion error", func(t *testing.T) {
		// given
		serviceBroker := new(v1beta1.ClusterServiceBroker)
		converter := automock.NewGQLClusterServiceBrokerConverter()

		converter.On("ToGQL", serviceBroker).Return(nil, errors.New("Conversion error")).Once()
		defer converter.AssertExpectations(t)
		serviceBrokerListener := listener.NewClusterServiceBroker(nil, filterClusterServiceBrokerTrue, converter)

		// when
		serviceBrokerListener.OnDelete(serviceBroker)
	})
}

func TestClusterServiceBrokerListener_OnUpdate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlClusterServiceBroker := new(gqlschema.ClusterServiceBroker)
		serviceBroker := new(v1beta1.ClusterServiceBroker)
		converter := automock.NewGQLClusterServiceBrokerConverter()

		channel := make(chan gqlschema.ClusterServiceBrokerEvent, 1)
		defer close(channel)
		converter.On("ToGQL", serviceBroker).Return(gqlClusterServiceBroker, nil).Once()
		defer converter.AssertExpectations(t)
		serviceBrokerListener := listener.NewClusterServiceBroker(channel, filterClusterServiceBrokerTrue, converter)

		// when
		serviceBrokerListener.OnUpdate(serviceBroker, serviceBroker)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeUpdate, result.Type)
		assert.Equal(t, *gqlClusterServiceBroker, result.ClusterServiceBroker)

	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		serviceBrokerListener := listener.NewClusterServiceBroker(nil, filterClusterServiceBrokerFalse, nil)

		// when
		serviceBrokerListener.OnUpdate(new(v1beta1.ClusterServiceBroker), new(v1beta1.ClusterServiceBroker))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		serviceBrokerListener := listener.NewClusterServiceBroker(nil, filterClusterServiceBrokerTrue, nil)

		// when
		serviceBrokerListener.OnUpdate(nil, nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		serviceBroker := new(v1beta1.ClusterServiceBroker)
		converter := automock.NewGQLClusterServiceBrokerConverter()

		converter.On("ToGQL", serviceBroker).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		serviceBrokerListener := listener.NewClusterServiceBroker(nil, filterClusterServiceBrokerTrue, converter)

		// when
		serviceBrokerListener.OnUpdate(nil, serviceBroker)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		serviceBrokerListener := listener.NewClusterServiceBroker(nil, filterClusterServiceBrokerTrue, nil)

		// when
		serviceBrokerListener.OnUpdate(new(struct{}), new(struct{}))
	})

	t.Run("Conversion error", func(t *testing.T) {
		// given
		serviceBroker := new(v1beta1.ClusterServiceBroker)
		converter := automock.NewGQLClusterServiceBrokerConverter()

		converter.On("ToGQL", serviceBroker).Return(nil, errors.New("Conversion error")).Once()
		defer converter.AssertExpectations(t)
		serviceBrokerListener := listener.NewClusterServiceBroker(nil, filterClusterServiceBrokerTrue, converter)

		// when
		serviceBrokerListener.OnUpdate(nil, serviceBroker)
	})
}

func filterClusterServiceBrokerTrue(o *v1beta1.ClusterServiceBroker) bool {
	return true
}

func filterClusterServiceBrokerFalse(o *v1beta1.ClusterServiceBroker) bool {
	return false
}
