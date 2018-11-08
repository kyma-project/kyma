package listener_test

import (
	"testing"

	api "github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/remoteenvironment/listener"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/remoteenvironment/listener/automock"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/stretchr/testify/assert"
)

func TestRemoteEnvironmentListener_OnAdd(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlRemoteEnvironment := new(gqlschema.RemoteEnvironment)
		remoteEnvironment := new(api.RemoteEnvironment)
		converter := automock.NewGQLRemoteEnvironmentConverter()

		channel := make(chan gqlschema.RemoteEnvironmentEvent, 1)
		defer close(channel)
		converter.On("ToGQL", remoteEnvironment).Return(*gqlRemoteEnvironment, nil).Once()
		defer converter.AssertExpectations(t)
		remoteEnvironmentListener := listener.NewRemoteEnvironment(channel, converter)

		// when
		remoteEnvironmentListener.OnAdd(remoteEnvironment)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeAdd, result.Type)
		assert.Equal(t, *gqlRemoteEnvironment, result.RemoteEnvironment)
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		remoteEnvironmentListener := listener.NewRemoteEnvironment(nil, nil)

		// when
		remoteEnvironmentListener.OnAdd(nil)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		remoteEnvironmentListener := listener.NewRemoteEnvironment(nil, nil)

		// when
		remoteEnvironmentListener.OnAdd(new(struct{}))
	})
}

func TestRemoteEnvironmentListener_OnDelete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlRemoteEnvironment := new(gqlschema.RemoteEnvironment)
		remoteEnvironment := new(api.RemoteEnvironment)
		converter := automock.NewGQLRemoteEnvironmentConverter()

		channel := make(chan gqlschema.RemoteEnvironmentEvent, 1)
		defer close(channel)
		converter.On("ToGQL", remoteEnvironment).Return(*gqlRemoteEnvironment, nil).Once()
		defer converter.AssertExpectations(t)
		remoteEnvironmentListener := listener.NewRemoteEnvironment(channel, converter)

		// when
		remoteEnvironmentListener.OnDelete(remoteEnvironment)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeDelete, result.Type)
		assert.Equal(t, *gqlRemoteEnvironment, result.RemoteEnvironment)

	})

	t.Run("Nil", func(t *testing.T) {
		// given
		remoteEnvironmentListener := listener.NewRemoteEnvironment(nil, nil)

		// when
		remoteEnvironmentListener.OnDelete(nil)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		remoteEnvironmentListener := listener.NewRemoteEnvironment(nil, nil)

		// when
		remoteEnvironmentListener.OnDelete(new(struct{}))
	})
}

func TestRemoteEnvironmentListener_OnUpdate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlRemoteEnvironment := new(gqlschema.RemoteEnvironment)
		remoteEnvironment := new(api.RemoteEnvironment)
		converter := automock.NewGQLRemoteEnvironmentConverter()

		channel := make(chan gqlschema.RemoteEnvironmentEvent, 1)
		defer close(channel)
		converter.On("ToGQL", remoteEnvironment).Return(*gqlRemoteEnvironment, nil).Once()
		defer converter.AssertExpectations(t)
		remoteEnvironmentListener := listener.NewRemoteEnvironment(channel, converter)

		// when
		remoteEnvironmentListener.OnUpdate(remoteEnvironment, remoteEnvironment)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeUpdate, result.Type)
		assert.Equal(t, *gqlRemoteEnvironment, result.RemoteEnvironment)

	})

	t.Run("Nil", func(t *testing.T) {
		// given
		remoteEnvironmentListener := listener.NewRemoteEnvironment(nil, nil)

		// when
		remoteEnvironmentListener.OnUpdate(nil, nil)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		remoteEnvironmentListener := listener.NewRemoteEnvironment(nil, nil)

		// when
		remoteEnvironmentListener.OnUpdate(new(struct{}), new(struct{}))
	})
}
