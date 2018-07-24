package servicecatalog_test

import (
	"testing"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/stretchr/testify/assert"
)

func TestInstanceListener_OnAdd(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		channel := make(chan gqlschema.ServiceInstanceEvent, 1)
		defer close(channel)
		gqlInstance := new(gqlschema.ServiceInstance)
		instance := new(v1beta1.ServiceInstance)
		converter := servicecatalog.NewMockInstanceConverter()
		converter.On("ToGQL", instance).Return(gqlInstance, nil).Once()
		listener := servicecatalog.NewInstanceListener(channel, filterTrue, converter)

		listener.OnAdd(instance)
		result := <-channel

		assert.Equal(t, gqlschema.ServiceInstanceEventTypeAdd, result.Type)
		assert.Equal(t, *gqlInstance, result.Instance)

	})

	t.Run("Filtered out", func(t *testing.T) {
		listener := servicecatalog.NewInstanceListener(nil, filterFalse, nil)

		listener.OnAdd(new(v1beta1.ServiceInstance))
	})

	t.Run("Nil", func(t *testing.T) {
		listener := servicecatalog.NewInstanceListener(nil, filterTrue, nil)

		listener.OnAdd(nil)
	})

	t.Run("Invalid type", func(t *testing.T) {
		listener := servicecatalog.NewInstanceListener(nil, filterTrue, nil)

		listener.OnAdd(new(struct{}))
	})
}

func TestInstanceListener_OnDelete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		channel := make(chan gqlschema.ServiceInstanceEvent, 1)
		defer close(channel)
		gqlInstance := new(gqlschema.ServiceInstance)
		instance := new(v1beta1.ServiceInstance)
		converter := servicecatalog.NewMockInstanceConverter()
		converter.On("ToGQL", instance).Return(gqlInstance, nil).Once()
		listener := servicecatalog.NewInstanceListener(channel, filterTrue, converter)

		listener.OnDelete(instance)
		result := <-channel

		assert.Equal(t, gqlschema.ServiceInstanceEventTypeDelete, result.Type)
		assert.Equal(t, *gqlInstance, result.Instance)
	})

	t.Run("Filtered out", func(t *testing.T) {
		listener := servicecatalog.NewInstanceListener(nil, filterFalse, nil)

		listener.OnDelete(new(v1beta1.ServiceInstance))
	})

	t.Run("Nil", func(t *testing.T) {
		listener := servicecatalog.NewInstanceListener(nil, filterTrue, nil)

		listener.OnDelete(nil)
	})

	t.Run("Invalid type", func(t *testing.T) {
		listener := servicecatalog.NewInstanceListener(nil, filterTrue, nil)

		listener.OnDelete(new(struct{}))
	})
}

func TestInstanceListener_OnUpdate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		channel := make(chan gqlschema.ServiceInstanceEvent, 1)
		defer close(channel)
		gqlInstance := new(gqlschema.ServiceInstance)
		instance := new(v1beta1.ServiceInstance)
		converter := servicecatalog.NewMockInstanceConverter()
		converter.On("ToGQL", instance).Return(gqlInstance, nil).Once()
		listener := servicecatalog.NewInstanceListener(channel, filterTrue, converter)

		listener.OnUpdate(instance, instance)
		result := <-channel

		assert.Equal(t, gqlschema.ServiceInstanceEventTypeUpdate, result.Type)
		assert.Equal(t, *gqlInstance, result.Instance)
	})

	t.Run("Filtered out", func(t *testing.T) {
		listener := servicecatalog.NewInstanceListener(nil, filterFalse, nil)

		listener.OnUpdate(new(v1beta1.ServiceInstance), new(v1beta1.ServiceInstance))
	})

	t.Run("Nil", func(t *testing.T) {
		listener := servicecatalog.NewInstanceListener(nil, filterTrue, nil)

		listener.OnUpdate(nil, nil)
	})

	t.Run("Invalid type", func(t *testing.T) {
		listener := servicecatalog.NewInstanceListener(nil, filterTrue, nil)

		listener.OnUpdate(new(struct{}), new(struct{}))
	})
}

func filterTrue(object interface{}) bool {
	return true
}

func filterFalse(object interface{}) bool {
	return false
}
