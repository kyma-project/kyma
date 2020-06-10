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

func TestServiceBrokerListener_OnAdd(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlServiceBroker := new(gqlschema.ServiceBroker)
		serviceBroker := new(v1beta1.ServiceBroker)
		converter := automock.NewGQLServiceBrokerConverter()

		channel := make(chan *gqlschema.ServiceBrokerEvent, 1)
		defer close(channel)
		converter.On("ToGQL", serviceBroker).Return(gqlServiceBroker, nil).Once()
		defer converter.AssertExpectations(t)
		serviceBrokerListener := listener.NewServiceBroker(channel, filterServiceBrokerTrue, converter)

		// when
		serviceBrokerListener.OnAdd(serviceBroker)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeAdd, result.Type)
		assert.Equal(t, gqlServiceBroker, result.ServiceBroker)
	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		serviceBrokerListener := listener.NewServiceBroker(nil, filterServiceBrokerFalse, nil)

		// when
		serviceBrokerListener.OnAdd(new(v1beta1.ServiceBroker))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		serviceBrokerListener := listener.NewServiceBroker(nil, filterServiceBrokerTrue, nil)

		// when
		serviceBrokerListener.OnAdd(nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		serviceBroker := new(v1beta1.ServiceBroker)
		converter := automock.NewGQLServiceBrokerConverter()

		converter.On("ToGQL", serviceBroker).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		serviceBrokerListener := listener.NewServiceBroker(nil, filterServiceBrokerTrue, converter)

		// when
		serviceBrokerListener.OnAdd(serviceBroker)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		serviceBrokerListener := listener.NewServiceBroker(nil, filterServiceBrokerTrue, nil)

		// when
		serviceBrokerListener.OnAdd(new(struct{}))
	})

	t.Run("Conversion error", func(t *testing.T) {
		// given
		serviceBroker := new(v1beta1.ServiceBroker)
		converter := automock.NewGQLServiceBrokerConverter()

		converter.On("ToGQL", serviceBroker).Return(nil, errors.New("Conversion error")).Once()
		defer converter.AssertExpectations(t)
		serviceBrokerListener := listener.NewServiceBroker(nil, filterServiceBrokerTrue, converter)

		// when
		serviceBrokerListener.OnAdd(serviceBroker)
	})
}

func TestServiceBrokerListener_OnDelete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlServiceBroker := new(gqlschema.ServiceBroker)
		serviceBroker := new(v1beta1.ServiceBroker)
		converter := automock.NewGQLServiceBrokerConverter()

		channel := make(chan *gqlschema.ServiceBrokerEvent, 1)
		defer close(channel)
		converter.On("ToGQL", serviceBroker).Return(gqlServiceBroker, nil).Once()
		defer converter.AssertExpectations(t)
		serviceBrokerListener := listener.NewServiceBroker(channel, filterServiceBrokerTrue, converter)

		// when
		serviceBrokerListener.OnDelete(serviceBroker)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeDelete, result.Type)
		assert.Equal(t, gqlServiceBroker, result.ServiceBroker)

	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		serviceBrokerListener := listener.NewServiceBroker(nil, filterServiceBrokerFalse, nil)

		// when
		serviceBrokerListener.OnDelete(new(v1beta1.ServiceBroker))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		serviceBrokerListener := listener.NewServiceBroker(nil, filterServiceBrokerTrue, nil)

		// when
		serviceBrokerListener.OnDelete(nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		serviceBroker := new(v1beta1.ServiceBroker)
		converter := automock.NewGQLServiceBrokerConverter()

		converter.On("ToGQL", serviceBroker).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		serviceBrokerListener := listener.NewServiceBroker(nil, filterServiceBrokerTrue, converter)

		// when
		serviceBrokerListener.OnDelete(serviceBroker)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		serviceBrokerListener := listener.NewServiceBroker(nil, filterServiceBrokerTrue, nil)

		// when
		serviceBrokerListener.OnDelete(new(struct{}))
	})

	t.Run("Conversion error", func(t *testing.T) {
		// given
		serviceBroker := new(v1beta1.ServiceBroker)
		converter := automock.NewGQLServiceBrokerConverter()

		converter.On("ToGQL", serviceBroker).Return(nil, errors.New("Conversion error")).Once()
		defer converter.AssertExpectations(t)
		serviceBrokerListener := listener.NewServiceBroker(nil, filterServiceBrokerTrue, converter)

		// when
		serviceBrokerListener.OnDelete(serviceBroker)
	})
}

func TestServiceBrokerListener_OnUpdate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlServiceBroker := new(gqlschema.ServiceBroker)
		serviceBroker := new(v1beta1.ServiceBroker)
		converter := automock.NewGQLServiceBrokerConverter()

		channel := make(chan *gqlschema.ServiceBrokerEvent, 1)
		defer close(channel)
		converter.On("ToGQL", serviceBroker).Return(gqlServiceBroker, nil).Once()
		defer converter.AssertExpectations(t)
		serviceBrokerListener := listener.NewServiceBroker(channel, filterServiceBrokerTrue, converter)

		// when
		serviceBrokerListener.OnUpdate(serviceBroker, serviceBroker)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeUpdate, result.Type)
		assert.Equal(t, gqlServiceBroker, result.ServiceBroker)

	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		serviceBrokerListener := listener.NewServiceBroker(nil, filterServiceBrokerFalse, nil)

		// when
		serviceBrokerListener.OnUpdate(new(v1beta1.ServiceBroker), new(v1beta1.ServiceBroker))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		serviceBrokerListener := listener.NewServiceBroker(nil, filterServiceBrokerTrue, nil)

		// when
		serviceBrokerListener.OnUpdate(nil, nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		serviceBroker := new(v1beta1.ServiceBroker)
		converter := automock.NewGQLServiceBrokerConverter()

		converter.On("ToGQL", serviceBroker).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		serviceBrokerListener := listener.NewServiceBroker(nil, filterServiceBrokerTrue, converter)

		// when
		serviceBrokerListener.OnUpdate(nil, serviceBroker)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		serviceBrokerListener := listener.NewServiceBroker(nil, filterServiceBrokerTrue, nil)

		// when
		serviceBrokerListener.OnUpdate(new(struct{}), new(struct{}))
	})

	t.Run("Conversion error", func(t *testing.T) {
		// given
		serviceBroker := new(v1beta1.ServiceBroker)
		converter := automock.NewGQLServiceBrokerConverter()

		converter.On("ToGQL", serviceBroker).Return(nil, errors.New("Conversion error")).Once()
		defer converter.AssertExpectations(t)
		serviceBrokerListener := listener.NewServiceBroker(nil, filterServiceBrokerTrue, converter)

		// when
		serviceBrokerListener.OnUpdate(nil, serviceBroker)
	})
}

func filterServiceBrokerTrue(o *v1beta1.ServiceBroker) bool {
	return true
}

func filterServiceBrokerFalse(o *v1beta1.ServiceBroker) bool {
	return false
}
