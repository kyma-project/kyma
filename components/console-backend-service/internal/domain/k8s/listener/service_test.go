package listener_test

import (
	"errors"
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/listener"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/listener/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	v1 "k8s.io/api/core/v1"

	"github.com/stretchr/testify/assert"
)

func TestServiceListener_OnAdd(t *testing.T) {
	assert := assert.New(t)

	t.Run("Success", func(t *testing.T) {
		// given
		gqlService := new(gqlschema.Service)
		service := new(v1.Service)
		converter := automock.NewGQLServiceConverter()

		channel := make(chan *gqlschema.ServiceEvent, 1)
		defer close(channel)
		converter.On("ToGQL", service).Return(gqlService, nil).Once()
		defer converter.AssertExpectations(t)
		serviceListener := listener.NewService(channel, filterServiceTrue, converter)

		// when
		serviceListener.OnAdd(service)
		result := <-channel

		// then
		assert.Equal(gqlschema.SubscriptionEventTypeAdd, result.Type)
		assert.Equal(gqlService, result.Service)
	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		serviceListener := listener.NewService(nil, filterServiceFalse, nil)

		// when
		serviceListener.OnAdd(new(v1.Service))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		serviceListener := listener.NewService(nil, filterServiceTrue, nil)

		// when
		serviceListener.OnAdd(nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		service := new(v1.Service)
		converter := automock.NewGQLServiceConverter()

		converter.On("ToGQL", service).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		serviceListener := listener.NewService(nil, filterServiceTrue, converter)

		// when
		serviceListener.OnAdd(service)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		serviceListener := listener.NewService(nil, filterServiceTrue, nil)

		// when
		serviceListener.OnAdd(new(struct{}))
	})

	t.Run("Conversion error", func(t *testing.T) {
		// given
		service := new(v1.Service)
		converter := automock.NewGQLServiceConverter()

		converter.On("ToGQL", service).Return(nil, errors.New("conversion error")).Once()
		defer converter.AssertExpectations(t)
		serviceListener := listener.NewService(nil, filterServiceTrue, converter)

		// when
		serviceListener.OnAdd(service)
	})
}

func TestServiceListener_OnDelete(t *testing.T) {
	assert := assert.New(t)

	t.Run("Success", func(t *testing.T) {
		// given
		gqlService := new(gqlschema.Service)
		service := new(v1.Service)
		converter := automock.NewGQLServiceConverter()

		channel := make(chan *gqlschema.ServiceEvent, 1)
		defer close(channel)
		converter.On("ToGQL", service).Return(gqlService, nil).Once()
		defer converter.AssertExpectations(t)
		serviceListener := listener.NewService(channel, filterServiceTrue, converter)

		// when
		serviceListener.OnDelete(service)
		result := <-channel

		// then
		assert.Equal(gqlschema.SubscriptionEventTypeDelete, result.Type)
		assert.Equal(gqlService, result.Service)

	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		serviceListener := listener.NewService(nil, filterServiceTrue, nil)

		// when
		serviceListener.OnDelete(new(v1.Pod))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		serviceListener := listener.NewService(nil, filterServiceTrue, nil)

		// when
		serviceListener.OnDelete(nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		service := new(v1.Service)
		converter := automock.NewGQLServiceConverter()

		converter.On("ToGQL", service).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		serviceListener := listener.NewService(nil, filterServiceTrue, converter)

		// when
		serviceListener.OnDelete(service)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		serviceListener := listener.NewService(nil, filterServiceTrue, nil)

		// when
		serviceListener.OnDelete(new(struct{}))
	})

	t.Run("Conversion error", func(t *testing.T) {
		// given
		service := new(v1.Service)
		converter := automock.NewGQLServiceConverter()

		converter.On("ToGQL", service).Return(nil, errors.New("conversion error")).Once()
		defer converter.AssertExpectations(t)
		serviceListener := listener.NewService(nil, filterServiceTrue, converter)

		// when
		serviceListener.OnDelete(service)
	})
}

func TestServiceListener_OnUpdate(t *testing.T) {
	assert := assert.New(t)

	t.Run("Success", func(t *testing.T) {
		// given
		gqlService := new(gqlschema.Service)
		service := new(v1.Service)
		converter := automock.NewGQLServiceConverter()

		channel := make(chan *gqlschema.ServiceEvent, 1)
		defer close(channel)
		converter.On("ToGQL", service).Return(gqlService, nil).Once()
		defer converter.AssertExpectations(t)
		serviceListener := listener.NewService(channel, filterServiceTrue, converter)

		// when
		serviceListener.OnUpdate(service, service)
		result := <-channel

		// then
		assert.Equal(gqlschema.SubscriptionEventTypeUpdate, result.Type)
		assert.Equal(gqlService, result.Service)

	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		serviceListener := listener.NewService(nil, filterServiceFalse, nil)

		// when
		serviceListener.OnUpdate(new(v1.Service), new(v1.Service))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		serviceListener := listener.NewService(nil, filterServiceTrue, nil)

		// when
		serviceListener.OnUpdate(nil, nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		service := new(v1.Service)
		converter := automock.NewGQLServiceConverter()

		converter.On("ToGQL", service).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		serviceListener := listener.NewService(nil, filterServiceTrue, converter)

		// when
		serviceListener.OnUpdate(nil, service)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		serviceListener := listener.NewService(nil, filterServiceTrue, nil)

		// when
		serviceListener.OnUpdate(new(struct{}), new(struct{}))
	})

	t.Run("Conversion error", func(t *testing.T) {
		// given
		service := new(v1.Service)
		converter := automock.NewGQLServiceConverter()

		converter.On("ToGQL", service).Return(nil, errors.New("conversion error")).Once()
		defer converter.AssertExpectations(t)
		serviceListener := listener.NewService(nil, filterServiceTrue, converter)

		// when
		serviceListener.OnUpdate(nil, service)
	})
}

func filterServiceTrue(_ *v1.Service) bool {
	return true
}

func filterServiceFalse(_ *v1.Service) bool {
	return false
}
