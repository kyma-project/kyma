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
		secret := new(v1.Secret)
		converter := automock.NewGQLSecretConverter()

		channel := make(chan gqlschema.SecretEvent, 1)
		defer close(channel)
		converter.On("ToGQL", secret).Return(gqlSecret, nil).Once()
		defer converter.AssertExpectations(t)
		secretListener := listener.NewSecret(channel, filterSecretTrue, converter)

		// when
		secretListener.OnAdd(secret)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeAdd, result.Type)
		assert.Equal(t, *gqlSecret, result.Secret)
	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		secretListener := listener.NewSecret(nil, filterSecretFalse, nil)

		// when
		secretListener.OnAdd(new(v1.Secret))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		secretListener := listener.NewSecret(nil, filterSecretTrue, nil)

		// when
		secretListener.OnAdd(nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		secret := new(v1.Secret)
		converter := automock.NewGQLSecretConverter()

		converter.On("ToGQL", secret).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		secretListener := listener.NewSecret(nil, filterSecretTrue, converter)

		// when
		secretListener.OnAdd(secret)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		secretListener := listener.NewSecret(nil, filterSecretTrue, nil)

		// when
		secretListener.OnAdd(new(struct{}))
	})

	t.Run("Conversion error", func(t *testing.T) {
		// given
		secret := new(v1.Secret)
		converter := automock.NewGQLSecretConverter()

		converter.On("ToGQL", secret).Return(nil, errors.New("Conversion error")).Once()
		defer converter.AssertExpectations(t)
		secretListener := listener.NewSecret(nil, filterSecretTrue, converter)

		// when
		secretListener.OnAdd(secret)
	})
}

func TestSecretListener_OnDelete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlSecret := new(gqlschema.Secret)
		secret := new(v1.Secret)
		converter := automock.NewGQLSecretConverter()

		channel := make(chan gqlschema.SecretEvent, 1)
		defer close(channel)
		converter.On("ToGQL", secret).Return(gqlSecret, nil).Once()
		defer converter.AssertExpectations(t)
		secretListener := listener.NewSecret(channel, filterSecretTrue, converter)

		// when
		secretListener.OnDelete(secret)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeDelete, result.Type)
		assert.Equal(t, *gqlSecret, result.Secret)

	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		secretListener := listener.NewSecret(nil, filterSecretFalse, nil)

		// when
		secretListener.OnDelete(new(v1.Secret))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		secretListener := listener.NewSecret(nil, filterSecretTrue, nil)

		// when
		secretListener.OnDelete(nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		secret := new(v1.Secret)
		converter := automock.NewGQLSecretConverter()

		converter.On("ToGQL", secret).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		secretListener := listener.NewSecret(nil, filterSecretTrue, converter)

		// when
		secretListener.OnDelete(secret)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		secretListener := listener.NewSecret(nil, filterSecretTrue, nil)

		// when
		secretListener.OnDelete(new(struct{}))
	})

	t.Run("Conversion error", func(t *testing.T) {
		// given
		secret := new(v1.Secret)
		converter := automock.NewGQLSecretConverter()

		converter.On("ToGQL", secret).Return(nil, errors.New("Conversion error")).Once()
		defer converter.AssertExpectations(t)
		secretListener := listener.NewSecret(nil, filterSecretTrue, converter)

		// when
		secretListener.OnDelete(secret)
	})
}

func TestSecretListener_OnUpdate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlSecret := new(gqlschema.Secret)
		secret := new(v1.Secret)
		converter := automock.NewGQLSecretConverter()

		channel := make(chan gqlschema.SecretEvent, 1)
		defer close(channel)
		converter.On("ToGQL", secret).Return(gqlSecret, nil).Once()
		defer converter.AssertExpectations(t)
		secretListener := listener.NewSecret(channel, filterSecretTrue, converter)

		// when
		secretListener.OnUpdate(secret, secret)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeUpdate, result.Type)
		assert.Equal(t, *gqlSecret, result.Secret)

	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		secretListener := listener.NewSecret(nil, filterSecretFalse, nil)

		// when
		secretListener.OnUpdate(new(v1.Secret), new(v1.Secret))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		secretListener := listener.NewSecret(nil, filterSecretTrue, nil)

		// when
		secretListener.OnUpdate(nil, nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		secret := new(v1.Secret)
		converter := automock.NewGQLSecretConverter()

		converter.On("ToGQL", secret).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		secretListener := listener.NewSecret(nil, filterSecretTrue, converter)

		// when
		secretListener.OnUpdate(nil, secret)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		secretListener := listener.NewSecret(nil, filterSecretTrue, nil)

		// when
		secretListener.OnUpdate(new(struct{}), new(struct{}))
	})

	t.Run("Conversion error", func(t *testing.T) {
		// given
		secret := new(v1.Secret)
		converter := automock.NewGQLSecretConverter()

		converter.On("ToGQL", secret).Return(nil, errors.New("Conversion error")).Once()
		defer converter.AssertExpectations(t)
		secretListener := listener.NewSecret(nil, filterSecretTrue, converter)

		// when
		secretListener.OnUpdate(nil, secret)
	})
}

func filterSecretTrue(o *v1.Secret) bool {
	return true
}

func filterSecretFalse(o *v1.Secret) bool {
	return false
}
