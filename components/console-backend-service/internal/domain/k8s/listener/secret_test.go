package listener_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/listener"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/listener/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func TestSecretListener_OnAdd(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlSecret := new(gqlschema.Secret)
		pod := new(v1.Secret)
		converter := automock.NewGQLSecretConverter()

		channel := make(chan gqlschema.SecretEvent, 1)
		defer close(channel)
		converter.On("ToGQL", pod).Return(gqlSecret, nil).Once()
		defer converter.AssertExpectations(t)
		podListener := listener.NewSecret(channel, filterSecretTrue, converter)

		// when
		podListener.OnAdd(pod)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeAdd, result.Type)
		assert.Equal(t, *gqlSecret, result.Secret)
	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		podListener := listener.NewSecret(nil, filterSecretFalse, nil)

		// when
		podListener.OnAdd(new(v1.Secret))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		podListener := listener.NewSecret(nil, filterSecretTrue, nil)

		// when
		podListener.OnAdd(nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		pod := new(v1.Secret)
		converter := automock.NewGQLSecretConverter()

		converter.On("ToGQL", pod).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		podListener := listener.NewSecret(nil, filterSecretTrue, converter)

		// when
		podListener.OnAdd(pod)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		podListener := listener.NewSecret(nil, filterSecretTrue, nil)

		// when
		podListener.OnAdd(new(struct{}))
	})

	t.Run("Conversion error", func(t *testing.T) {
		// given
		pod := new(v1.Secret)
		converter := automock.NewGQLSecretConverter()

		converter.On("ToGQL", pod).Return(nil, errors.New("Conversion error")).Once()
		defer converter.AssertExpectations(t)
		podListener := listener.NewSecret(nil, filterSecretTrue, converter)

		// when
		podListener.OnAdd(pod)
	})
}

func TestSecretListener_OnDelete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlSecret := new(gqlschema.Secret)
		pod := new(v1.Secret)
		converter := automock.NewGQLSecretConverter()

		channel := make(chan gqlschema.SecretEvent, 1)
		defer close(channel)
		converter.On("ToGQL", pod).Return(gqlSecret, nil).Once()
		defer converter.AssertExpectations(t)
		podListener := listener.NewSecret(channel, filterSecretTrue, converter)

		// when
		podListener.OnDelete(pod)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeDelete, result.Type)
		assert.Equal(t, *gqlSecret, result.Secret)

	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		podListener := listener.NewSecret(nil, filterSecretFalse, nil)

		// when
		podListener.OnDelete(new(v1.Secret))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		podListener := listener.NewSecret(nil, filterSecretTrue, nil)

		// when
		podListener.OnDelete(nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		pod := new(v1.Secret)
		converter := automock.NewGQLSecretConverter()

		converter.On("ToGQL", pod).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		podListener := listener.NewSecret(nil, filterSecretTrue, converter)

		// when
		podListener.OnDelete(pod)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		podListener := listener.NewSecret(nil, filterSecretTrue, nil)

		// when
		podListener.OnDelete(new(struct{}))
	})

	t.Run("Conversion error", func(t *testing.T) {
		// given
		pod := new(v1.Secret)
		converter := automock.NewGQLSecretConverter()

		converter.On("ToGQL", pod).Return(nil, errors.New("Conversion error")).Once()
		defer converter.AssertExpectations(t)
		podListener := listener.NewSecret(nil, filterSecretTrue, converter)

		// when
		podListener.OnDelete(pod)
	})
}

func TestSecretListener_OnUpdate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlSecret := new(gqlschema.Secret)
		pod := new(v1.Secret)
		converter := automock.NewGQLSecretConverter()

		channel := make(chan gqlschema.SecretEvent, 1)
		defer close(channel)
		converter.On("ToGQL", pod).Return(gqlSecret, nil).Once()
		defer converter.AssertExpectations(t)
		podListener := listener.NewSecret(channel, filterSecretTrue, converter)

		// when
		podListener.OnUpdate(pod, pod)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeUpdate, result.Type)
		assert.Equal(t, *gqlSecret, result.Secret)

	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		podListener := listener.NewSecret(nil, filterSecretFalse, nil)

		// when
		podListener.OnUpdate(new(v1.Secret), new(v1.Secret))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		podListener := listener.NewSecret(nil, filterSecretTrue, nil)

		// when
		podListener.OnUpdate(nil, nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		pod := new(v1.Secret)
		converter := automock.NewGQLSecretConverter()

		converter.On("ToGQL", pod).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		podListener := listener.NewSecret(nil, filterSecretTrue, converter)

		// when
		podListener.OnUpdate(nil, pod)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		podListener := listener.NewSecret(nil, filterSecretTrue, nil)

		// when
		podListener.OnUpdate(new(struct{}), new(struct{}))
	})

	t.Run("Conversion error", func(t *testing.T) {
		// given
		pod := new(v1.Secret)
		converter := automock.NewGQLSecretConverter()

		converter.On("ToGQL", pod).Return(nil, errors.New("Conversion error")).Once()
		defer converter.AssertExpectations(t)
		podListener := listener.NewSecret(nil, filterSecretTrue, converter)

		// when
		podListener.OnUpdate(nil, pod)
	})
}

func filterSecretTrue(o *v1.Secret) bool {
	return true
}

func filterSecretFalse(o *v1.Secret) bool {
	return false
}
