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

func TestInstanceListener_OnAdd(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlInstance := new(gqlschema.ServiceInstance)
		instance := new(v1beta1.ServiceInstance)
		converter := automock.NewGQLInstanceConverter()

		channel := make(chan *gqlschema.ServiceInstanceEvent, 1)
		defer close(channel)
		converter.On("ToGQL", instance).Return(gqlInstance, nil).Once()
		defer converter.AssertExpectations(t)
		instanceListener := listener.NewInstance(channel, filterInstanceTrue, converter)

		// when
		instanceListener.OnAdd(instance)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeAdd, result.Type)
		assert.Equal(t, gqlInstance, result.ServiceInstance)
	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		instanceListener := listener.NewInstance(nil, filterInstanceFalse, nil)

		// when
		instanceListener.OnAdd(new(v1beta1.ServiceInstance))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		instanceListener := listener.NewInstance(nil, filterInstanceTrue, nil)

		// when
		instanceListener.OnAdd(nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		instance := new(v1beta1.ServiceInstance)
		converter := automock.NewGQLInstanceConverter()

		converter.On("ToGQL", instance).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		instanceListener := listener.NewInstance(nil, filterInstanceTrue, converter)

		// when
		instanceListener.OnAdd(instance)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		instanceListener := listener.NewInstance(nil, filterInstanceTrue, nil)

		// when
		instanceListener.OnAdd(new(struct{}))
	})

	t.Run("Conversion error", func(t *testing.T) {
		// given
		instance := new(v1beta1.ServiceInstance)
		converter := automock.NewGQLInstanceConverter()

		converter.On("ToGQL", instance).Return(nil, errors.New("Conversion error")).Once()
		defer converter.AssertExpectations(t)
		instanceListener := listener.NewInstance(nil, filterInstanceTrue, converter)

		// when
		instanceListener.OnAdd(instance)
	})
}

func TestInstanceListener_OnDelete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlInstance := new(gqlschema.ServiceInstance)
		instance := new(v1beta1.ServiceInstance)
		converter := automock.NewGQLInstanceConverter()

		channel := make(chan *gqlschema.ServiceInstanceEvent, 1)
		defer close(channel)
		converter.On("ToGQL", instance).Return(gqlInstance, nil).Once()
		defer converter.AssertExpectations(t)
		instanceListener := listener.NewInstance(channel, filterInstanceTrue, converter)

		// when
		instanceListener.OnDelete(instance)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeDelete, result.Type)
		assert.Equal(t, gqlInstance, result.ServiceInstance)

	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		instanceListener := listener.NewInstance(nil, filterInstanceFalse, nil)

		// when
		instanceListener.OnDelete(new(v1beta1.ServiceInstance))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		instanceListener := listener.NewInstance(nil, filterInstanceTrue, nil)

		// when
		instanceListener.OnDelete(nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		instance := new(v1beta1.ServiceInstance)
		converter := automock.NewGQLInstanceConverter()

		converter.On("ToGQL", instance).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		instanceListener := listener.NewInstance(nil, filterInstanceTrue, converter)

		// when
		instanceListener.OnDelete(instance)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		instanceListener := listener.NewInstance(nil, filterInstanceTrue, nil)

		// when
		instanceListener.OnDelete(new(struct{}))
	})

	t.Run("Conversion error", func(t *testing.T) {
		// given
		instance := new(v1beta1.ServiceInstance)
		converter := automock.NewGQLInstanceConverter()

		converter.On("ToGQL", instance).Return(nil, errors.New("Conversion error")).Once()
		defer converter.AssertExpectations(t)
		instanceListener := listener.NewInstance(nil, filterInstanceTrue, converter)

		// when
		instanceListener.OnDelete(instance)
	})
}

func TestInstanceListener_OnUpdate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlInstance := new(gqlschema.ServiceInstance)
		instance := new(v1beta1.ServiceInstance)
		converter := automock.NewGQLInstanceConverter()

		channel := make(chan *gqlschema.ServiceInstanceEvent, 1)
		defer close(channel)
		converter.On("ToGQL", instance).Return(gqlInstance, nil).Once()
		defer converter.AssertExpectations(t)
		instanceListener := listener.NewInstance(channel, filterInstanceTrue, converter)

		// when
		instanceListener.OnUpdate(instance, instance)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeUpdate, result.Type)
		assert.Equal(t, gqlInstance, result.ServiceInstance)

	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		instanceListener := listener.NewInstance(nil, filterInstanceFalse, nil)

		// when
		instanceListener.OnUpdate(new(v1beta1.ServiceInstance), new(v1beta1.ServiceInstance))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		instanceListener := listener.NewInstance(nil, filterInstanceTrue, nil)

		// when
		instanceListener.OnUpdate(nil, nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		instance := new(v1beta1.ServiceInstance)
		converter := automock.NewGQLInstanceConverter()

		converter.On("ToGQL", instance).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		instanceListener := listener.NewInstance(nil, filterInstanceTrue, converter)

		// when
		instanceListener.OnUpdate(nil, instance)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		instanceListener := listener.NewInstance(nil, filterInstanceTrue, nil)

		// when
		instanceListener.OnUpdate(new(struct{}), new(struct{}))
	})

	t.Run("Conversion error", func(t *testing.T) {
		// given
		instance := new(v1beta1.ServiceInstance)
		converter := automock.NewGQLInstanceConverter()

		converter.On("ToGQL", instance).Return(nil, errors.New("Conversion error")).Once()
		defer converter.AssertExpectations(t)
		instanceListener := listener.NewInstance(nil, filterInstanceTrue, converter)

		// when
		instanceListener.OnUpdate(nil, instance)
	})
}

func filterInstanceTrue(o *v1beta1.ServiceInstance) bool {
	return true
}

func filterInstanceFalse(o *v1beta1.ServiceInstance) bool {
	return false
}
