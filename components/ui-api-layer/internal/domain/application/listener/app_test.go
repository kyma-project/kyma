package listener_test

import (
	"testing"

	api "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/application/listener"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/application/listener/automock"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/stretchr/testify/assert"
)

func TestApplicationListener_OnAdd(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlApplication := new(gqlschema.Application)
		application := new(api.Application)
		converter := automock.NewGQLApplicationConverter()

		channel := make(chan gqlschema.ApplicationEvent, 1)
		defer close(channel)
		converter.On("ToGQL", application).Return(*gqlApplication, nil).Once()
		defer converter.AssertExpectations(t)
		applicationListener := listener.NewApplication(channel, converter)

		// when
		applicationListener.OnAdd(application)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeAdd, result.Type)
		assert.Equal(t, *gqlApplication, result.Application)
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		applicationListener := listener.NewApplication(nil, nil)

		// when
		applicationListener.OnAdd(nil)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		applicationListener := listener.NewApplication(nil, nil)

		// when
		applicationListener.OnAdd(new(struct{}))
	})
}

func TestApplicationListener_OnDelete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlApplication := new(gqlschema.Application)
		application := new(api.Application)
		converter := automock.NewGQLApplicationConverter()

		channel := make(chan gqlschema.ApplicationEvent, 1)
		defer close(channel)
		converter.On("ToGQL", application).Return(*gqlApplication, nil).Once()
		defer converter.AssertExpectations(t)
		applicationListener := listener.NewApplication(channel, converter)

		// when
		applicationListener.OnDelete(application)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeDelete, result.Type)
		assert.Equal(t, *gqlApplication, result.Application)

	})

	t.Run("Nil", func(t *testing.T) {
		// given
		applicationListener := listener.NewApplication(nil, nil)

		// when
		applicationListener.OnDelete(nil)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		applicationListener := listener.NewApplication(nil, nil)

		// when
		applicationListener.OnDelete(new(struct{}))
	})
}

func TestApplicationListener_OnUpdate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlApplication := new(gqlschema.Application)
		application := new(api.Application)
		converter := automock.NewGQLApplicationConverter()

		channel := make(chan gqlschema.ApplicationEvent, 1)
		defer close(channel)
		converter.On("ToGQL", application).Return(*gqlApplication, nil).Once()
		defer converter.AssertExpectations(t)
		applicationListener := listener.NewApplication(channel, converter)

		// when
		applicationListener.OnUpdate(application, application)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeUpdate, result.Type)
		assert.Equal(t, *gqlApplication, result.Application)

	})

	t.Run("Nil", func(t *testing.T) {
		// given
		applicationListener := listener.NewApplication(nil, nil)

		// when
		applicationListener.OnUpdate(nil, nil)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		applicationListener := listener.NewApplication(nil, nil)

		// when
		applicationListener.OnUpdate(new(struct{}), new(struct{}))
	})
}
