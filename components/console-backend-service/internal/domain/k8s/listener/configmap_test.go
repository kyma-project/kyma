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

func TestConfigMapListener_OnAdd(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlConfigMap := new(gqlschema.ConfigMap)
		configMap := new(v1.ConfigMap)
		converter := automock.NewGQLConfigMapConverter()

		channel := make(chan gqlschema.ConfigMapEvent, 1)
		defer close(channel)
		converter.On("ToGQL", configMap).Return(gqlConfigMap, nil).Once()
		defer converter.AssertExpectations(t)
		configMapListener := listener.NewConfigMap(channel, filterConfigMapTrue, converter)

		// when
		configMapListener.OnAdd(configMap)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeAdd, result.Type)
		assert.Equal(t, *gqlConfigMap, result.ConfigMap)
	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		configMapListener := listener.NewConfigMap(nil, filterConfigMapFalse, nil)

		// when
		configMapListener.OnAdd(new(v1.ConfigMap))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		configMapListener := listener.NewConfigMap(nil, filterConfigMapTrue, nil)

		// when
		configMapListener.OnAdd(nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		configMap := new(v1.ConfigMap)
		converter := automock.NewGQLConfigMapConverter()

		converter.On("ToGQL", configMap).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		configMapListener := listener.NewConfigMap(nil, filterConfigMapTrue, converter)

		// when
		configMapListener.OnAdd(configMap)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		configMapListener := listener.NewConfigMap(nil, filterConfigMapTrue, nil)

		// when
		configMapListener.OnAdd(new(struct{}))
	})

	t.Run("Conversion error", func(t *testing.T) {
		// given
		configMap := new(v1.ConfigMap)
		converter := automock.NewGQLConfigMapConverter()

		converter.On("ToGQL", configMap).Return(nil, errors.New("Conversion error")).Once()
		defer converter.AssertExpectations(t)
		configMapListener := listener.NewConfigMap(nil, filterConfigMapTrue, converter)

		// when
		configMapListener.OnAdd(configMap)
	})
}

func TestConfigMapListener_OnDelete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlConfigMap := new(gqlschema.ConfigMap)
		configMap := new(v1.ConfigMap)
		converter := automock.NewGQLConfigMapConverter()

		channel := make(chan gqlschema.ConfigMapEvent, 1)
		defer close(channel)
		converter.On("ToGQL", configMap).Return(gqlConfigMap, nil).Once()
		defer converter.AssertExpectations(t)
		configMapListener := listener.NewConfigMap(channel, filterConfigMapTrue, converter)

		// when
		configMapListener.OnDelete(configMap)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeDelete, result.Type)
		assert.Equal(t, *gqlConfigMap, result.ConfigMap)

	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		configMapListener := listener.NewConfigMap(nil, filterConfigMapFalse, nil)

		// when
		configMapListener.OnDelete(new(v1.ConfigMap))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		configMapListener := listener.NewConfigMap(nil, filterConfigMapTrue, nil)

		// when
		configMapListener.OnDelete(nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		configMap := new(v1.ConfigMap)
		converter := automock.NewGQLConfigMapConverter()

		converter.On("ToGQL", configMap).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		configMapListener := listener.NewConfigMap(nil, filterConfigMapTrue, converter)

		// when
		configMapListener.OnDelete(configMap)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		configMapListener := listener.NewConfigMap(nil, filterConfigMapTrue, nil)

		// when
		configMapListener.OnDelete(new(struct{}))
	})

	t.Run("Conversion error", func(t *testing.T) {
		// given
		configMap := new(v1.ConfigMap)
		converter := automock.NewGQLConfigMapConverter()

		converter.On("ToGQL", configMap).Return(nil, errors.New("Conversion error")).Once()
		defer converter.AssertExpectations(t)
		configMapListener := listener.NewConfigMap(nil, filterConfigMapTrue, converter)

		// when
		configMapListener.OnDelete(configMap)
	})
}

func TestConfigMapListener_OnUpdate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlConfigMap := new(gqlschema.ConfigMap)
		configMap := new(v1.ConfigMap)
		converter := automock.NewGQLConfigMapConverter()

		channel := make(chan gqlschema.ConfigMapEvent, 1)
		defer close(channel)
		converter.On("ToGQL", configMap).Return(gqlConfigMap, nil).Once()
		defer converter.AssertExpectations(t)
		configMapListener := listener.NewConfigMap(channel, filterConfigMapTrue, converter)

		// when
		configMapListener.OnUpdate(configMap, configMap)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeUpdate, result.Type)
		assert.Equal(t, *gqlConfigMap, result.ConfigMap)

	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		configMapListener := listener.NewConfigMap(nil, filterConfigMapFalse, nil)

		// when
		configMapListener.OnUpdate(new(v1.ConfigMap), new(v1.ConfigMap))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		configMapListener := listener.NewConfigMap(nil, filterConfigMapTrue, nil)

		// when
		configMapListener.OnUpdate(nil, nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		configMap := new(v1.ConfigMap)
		converter := automock.NewGQLConfigMapConverter()

		converter.On("ToGQL", configMap).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		configMapListener := listener.NewConfigMap(nil, filterConfigMapTrue, converter)

		// when
		configMapListener.OnUpdate(nil, configMap)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		configMapListener := listener.NewConfigMap(nil, filterConfigMapTrue, nil)

		// when
		configMapListener.OnUpdate(new(struct{}), new(struct{}))
	})

	t.Run("Conversion error", func(t *testing.T) {
		// given
		configMap := new(v1.ConfigMap)
		converter := automock.NewGQLConfigMapConverter()

		converter.On("ToGQL", configMap).Return(nil, errors.New("Conversion error")).Once()
		defer converter.AssertExpectations(t)
		configMapListener := listener.NewConfigMap(nil, filterConfigMapTrue, converter)

		// when
		configMapListener.OnUpdate(nil, configMap)
	})
}

func filterConfigMapTrue(o *v1.ConfigMap) bool {
	return true
}

func filterConfigMapFalse(o *v1.ConfigMap) bool {
	return false
}
