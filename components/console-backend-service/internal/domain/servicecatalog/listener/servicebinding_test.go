package listener_test

import (
	"testing"

	api "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalog/listener"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalog/listener/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
)

func TestBinding_OnAdd(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlBinding := new(gqlschema.ServiceBinding)
		binding := new(api.ServiceBinding)
		converter := automock.NewGQLBindingConverter()

		channel := make(chan gqlschema.ServiceBindingEvent, 1)
		defer close(channel)
		converter.On("ToGQL", binding).Return(gqlBinding, nil).Once()
		defer converter.AssertExpectations(t)
		bindingListener := listener.NewBinding(channel, filterBindingTrue, converter)

		// when
		bindingListener.OnAdd(binding)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeAdd, result.Type)
		assert.Equal(t, *gqlBinding, result.ServiceBinding)
	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		bindingListener := listener.NewBinding(nil, filterBindingFalse, nil)

		// when
		bindingListener.OnAdd(new(api.ServiceBinding))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		bindingListener := listener.NewBinding(nil, filterBindingTrue, nil)

		// when
		bindingListener.OnAdd(nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		binding := new(api.ServiceBinding)
		converter := automock.NewGQLBindingConverter()

		converter.On("ToGQL", binding).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		bindingListener := listener.NewBinding(nil, filterBindingTrue, converter)

		// when
		bindingListener.OnAdd(binding)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		bindingListener := listener.NewBinding(nil, filterBindingTrue, nil)

		// when
		bindingListener.OnAdd(new(struct{}))
	})
}

func TestBinding_OnDelete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlBinding := new(gqlschema.ServiceBinding)
		binding := new(api.ServiceBinding)
		converter := automock.NewGQLBindingConverter()

		channel := make(chan gqlschema.ServiceBindingEvent, 1)
		defer close(channel)
		converter.On("ToGQL", binding).Return(gqlBinding, nil).Once()
		defer converter.AssertExpectations(t)
		bindingListener := listener.NewBinding(channel, filterBindingTrue, converter)

		// when
		bindingListener.OnDelete(binding)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeDelete, result.Type)
		assert.Equal(t, *gqlBinding, result.ServiceBinding)

	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		bindingListener := listener.NewBinding(nil, filterBindingFalse, nil)

		// when
		bindingListener.OnDelete(new(api.ServiceBinding))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		bindingListener := listener.NewBinding(nil, filterBindingTrue, nil)

		// when
		bindingListener.OnDelete(nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		binding := new(api.ServiceBinding)
		converter := automock.NewGQLBindingConverter()

		converter.On("ToGQL", binding).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		bindingListener := listener.NewBinding(nil, filterBindingTrue, converter)

		// when
		bindingListener.OnDelete(binding)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		bindingListener := listener.NewBinding(nil, filterBindingTrue, nil)

		// when
		bindingListener.OnDelete(new(struct{}))
	})
}

func TestBinding_OnUpdate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlBinding := new(gqlschema.ServiceBinding)
		binding := new(api.ServiceBinding)
		converter := automock.NewGQLBindingConverter()

		channel := make(chan gqlschema.ServiceBindingEvent, 1)
		defer close(channel)
		converter.On("ToGQL", binding).Return(gqlBinding, nil).Once()
		defer converter.AssertExpectations(t)
		bindingListener := listener.NewBinding(channel, filterBindingTrue, converter)

		// when
		bindingListener.OnUpdate(binding, binding)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeUpdate, result.Type)
		assert.Equal(t, *gqlBinding, result.ServiceBinding)

	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		bindingListener := listener.NewBinding(nil, filterBindingFalse, nil)

		// when
		bindingListener.OnUpdate(new(api.ServiceBinding), new(api.ServiceBinding))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		bindingListener := listener.NewBinding(nil, filterBindingTrue, nil)

		// when
		bindingListener.OnUpdate(nil, nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		binding := new(api.ServiceBinding)
		converter := automock.NewGQLBindingConverter()

		converter.On("ToGQL", binding).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		bindingListener := listener.NewBinding(nil, filterBindingTrue, converter)

		// when
		bindingListener.OnUpdate(nil, binding)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		bindingListener := listener.NewBinding(nil, filterBindingTrue, nil)

		// when
		bindingListener.OnUpdate(new(struct{}), new(struct{}))
	})
}

func filterBindingTrue(o *api.ServiceBinding) bool {
	return true
}

func filterBindingFalse(o *api.ServiceBinding) bool {
	return false
}
