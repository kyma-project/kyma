package listener_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons/listener"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons/listener/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	api "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestBindingUsage_OnAdd(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlBindingUsage := new(gqlschema.ServiceBindingUsage)
		bindingUsage := new(api.ServiceBindingUsage)
		converter := automock.NewGQLBindingUsageConverter()

		channel := make(chan *gqlschema.ServiceBindingUsageEvent, 1)
		defer close(channel)
		converter.On("ToGQL", bindingUsage).Return(gqlBindingUsage, nil).Once()
		defer converter.AssertExpectations(t)
		bindingUsageListener := listener.NewBindingUsage(channel, filterBindingUsageTrue, converter)

		// when
		bindingUsageListener.OnAdd(bindingUsage)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeAdd, result.Type)
		assert.Equal(t, gqlBindingUsage, result.ServiceBindingUsage)
	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		bindingUsageListener := listener.NewBindingUsage(nil, filterBindingUsageFalse, nil)

		// when
		bindingUsageListener.OnAdd(new(api.ServiceBindingUsage))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		bindingUsageListener := listener.NewBindingUsage(nil, filterBindingUsageTrue, nil)

		// when
		bindingUsageListener.OnAdd(nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		bindingUsage := new(api.ServiceBindingUsage)
		converter := automock.NewGQLBindingUsageConverter()

		converter.On("ToGQL", bindingUsage).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		bindingUsageListener := listener.NewBindingUsage(nil, filterBindingUsageTrue, converter)

		// when
		bindingUsageListener.OnAdd(bindingUsage)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		bindingUsageListener := listener.NewBindingUsage(nil, filterBindingUsageTrue, nil)

		// when
		bindingUsageListener.OnAdd(new(struct{}))
	})

	t.Run("Conversion error", func(t *testing.T) {
		// given
		bindingUsage := new(api.ServiceBindingUsage)
		converter := automock.NewGQLBindingUsageConverter()

		converter.On("ToGQL", bindingUsage).Return(nil, errors.New("random error")).Once()
		defer converter.AssertExpectations(t)
		bindingUsageListener := listener.NewBindingUsage(nil, filterBindingUsageTrue, converter)

		// when
		bindingUsageListener.OnAdd(bindingUsage)
	})
}

func TestBindingUsage_OnDelete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlBindingUsage := new(gqlschema.ServiceBindingUsage)
		bindingUsage := new(api.ServiceBindingUsage)
		converter := automock.NewGQLBindingUsageConverter()

		channel := make(chan *gqlschema.ServiceBindingUsageEvent, 1)
		defer close(channel)
		converter.On("ToGQL", bindingUsage).Return(gqlBindingUsage, nil).Once()
		defer converter.AssertExpectations(t)
		bindingUsageListener := listener.NewBindingUsage(channel, filterBindingUsageTrue, converter)

		// when
		bindingUsageListener.OnDelete(bindingUsage)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeDelete, result.Type)
		assert.Equal(t, gqlBindingUsage, result.ServiceBindingUsage)

	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		bindingUsageListener := listener.NewBindingUsage(nil, filterBindingUsageFalse, nil)

		// when
		bindingUsageListener.OnDelete(new(api.ServiceBindingUsage))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		bindingUsageListener := listener.NewBindingUsage(nil, filterBindingUsageTrue, nil)

		// when
		bindingUsageListener.OnDelete(nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		bindingUsage := new(api.ServiceBindingUsage)
		converter := automock.NewGQLBindingUsageConverter()

		converter.On("ToGQL", bindingUsage).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		bindingUsageListener := listener.NewBindingUsage(nil, filterBindingUsageTrue, converter)

		// when
		bindingUsageListener.OnDelete(bindingUsage)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		bindingUsageListener := listener.NewBindingUsage(nil, filterBindingUsageTrue, nil)

		// when
		bindingUsageListener.OnDelete(new(struct{}))
	})

	t.Run("Conversion error", func(t *testing.T) {
		// given
		bindingUsage := new(api.ServiceBindingUsage)
		converter := automock.NewGQLBindingUsageConverter()

		converter.On("ToGQL", bindingUsage).Return(nil, errors.New("random error")).Once()
		defer converter.AssertExpectations(t)
		bindingUsageListener := listener.NewBindingUsage(nil, filterBindingUsageTrue, converter)

		// when
		bindingUsageListener.OnDelete(bindingUsage)
	})
}

func TestBindingUsage_OnUpdate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlBindingUsage := new(gqlschema.ServiceBindingUsage)
		bindingUsage := new(api.ServiceBindingUsage)
		converter := automock.NewGQLBindingUsageConverter()

		channel := make(chan *gqlschema.ServiceBindingUsageEvent, 1)
		defer close(channel)
		converter.On("ToGQL", bindingUsage).Return(gqlBindingUsage, nil).Once()
		defer converter.AssertExpectations(t)
		bindingUsageListener := listener.NewBindingUsage(channel, filterBindingUsageTrue, converter)

		// when
		bindingUsageListener.OnUpdate(bindingUsage, bindingUsage)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeUpdate, result.Type)
		assert.Equal(t, gqlBindingUsage, result.ServiceBindingUsage)

	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		bindingUsageListener := listener.NewBindingUsage(nil, filterBindingUsageFalse, nil)

		// when
		bindingUsageListener.OnUpdate(new(api.ServiceBindingUsage), new(api.ServiceBindingUsage))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		bindingUsageListener := listener.NewBindingUsage(nil, filterBindingUsageTrue, nil)

		// when
		bindingUsageListener.OnUpdate(nil, nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		bindingUsage := new(api.ServiceBindingUsage)
		converter := automock.NewGQLBindingUsageConverter()

		converter.On("ToGQL", bindingUsage).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		bindingUsageListener := listener.NewBindingUsage(nil, filterBindingUsageTrue, converter)

		// when
		bindingUsageListener.OnUpdate(nil, bindingUsage)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		bindingUsageListener := listener.NewBindingUsage(nil, filterBindingUsageTrue, nil)

		// when
		bindingUsageListener.OnUpdate(new(struct{}), new(struct{}))
	})

	t.Run("Conversion error", func(t *testing.T) {
		// given
		bindingUsage := new(api.ServiceBindingUsage)
		converter := automock.NewGQLBindingUsageConverter()

		converter.On("ToGQL", bindingUsage).Return(nil, errors.New("random error")).Once()
		defer converter.AssertExpectations(t)
		bindingUsageListener := listener.NewBindingUsage(nil, filterBindingUsageTrue, converter)

		// when
		bindingUsageListener.OnUpdate(bindingUsage, bindingUsage)
	})
}

func filterBindingUsageTrue(o *api.ServiceBindingUsage) bool {
	return true
}

func filterBindingUsageFalse(o *api.ServiceBindingUsage) bool {
	return false
}
