package listener

import (
	"testing"

	"github.com/kyma-project/helm-broker/pkg/apis/addons/v1alpha1"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons/listener/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
)

func TestAddonsConfiguration_OnAdd(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlAddonsConfiguration := new(gqlschema.AddonsConfiguration)
		cfg := new(v1alpha1.AddonsConfiguration)
		converter := automock.NewGQLAddonsConfigurationConverter()

		channel := make(chan gqlschema.AddonsConfigurationEvent, 1)
		defer close(channel)
		converter.On("ToGQL", cfg).Return(gqlAddonsConfiguration, nil).Once()
		defer converter.AssertExpectations(t)
		serviceBrokerListener := NewAddonsConfiguration(channel, filterAddonsConfigurationTrue, converter)

		// when
		serviceBrokerListener.OnAdd(cfg)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeAdd, result.Type)
		assert.Equal(t, *gqlAddonsConfiguration, result.AddonsConfiguration)
	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		serviceBrokerListener := NewAddonsConfiguration(nil, filterAddonsConfigurationFalse, nil)

		// when
		serviceBrokerListener.OnAdd(new(v1alpha1.AddonsConfiguration))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		serviceBrokerListener := NewAddonsConfiguration(nil, filterAddonsConfigurationTrue, nil)

		// when
		serviceBrokerListener.OnAdd(nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		serviceBroker := new(v1alpha1.AddonsConfiguration)
		converter := automock.NewGQLAddonsConfigurationConverter()

		converter.On("ToGQL", serviceBroker).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		serviceBrokerListener := NewAddonsConfiguration(nil, filterAddonsConfigurationTrue, converter)

		// when
		serviceBrokerListener.OnAdd(serviceBroker)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		serviceBrokerListener := NewAddonsConfiguration(nil, filterAddonsConfigurationTrue, nil)

		// when
		serviceBrokerListener.OnAdd(new(struct{}))
	})

}

func TestAddonsConfiguration_OnDelete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlClusterServiceBroker := new(gqlschema.AddonsConfiguration)
		serviceBroker := new(v1alpha1.AddonsConfiguration)
		converter := automock.NewGQLAddonsConfigurationConverter()

		channel := make(chan gqlschema.AddonsConfigurationEvent, 1)
		defer close(channel)
		converter.On("ToGQL", serviceBroker).Return(gqlClusterServiceBroker, nil).Once()
		defer converter.AssertExpectations(t)
		serviceBrokerListener := NewAddonsConfiguration(channel, filterAddonsConfigurationTrue, converter)

		// when
		serviceBrokerListener.OnDelete(serviceBroker)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeDelete, result.Type)
		assert.Equal(t, *gqlClusterServiceBroker, result.AddonsConfiguration)

	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		serviceBrokerListener := NewAddonsConfiguration(nil, filterAddonsConfigurationFalse, nil)

		// when
		serviceBrokerListener.OnDelete(new(v1alpha1.AddonsConfiguration))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		serviceBrokerListener := NewAddonsConfiguration(nil, filterAddonsConfigurationTrue, nil)

		// when
		serviceBrokerListener.OnDelete(nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		serviceBroker := new(v1alpha1.AddonsConfiguration)
		converter := automock.NewGQLAddonsConfigurationConverter()

		converter.On("ToGQL", serviceBroker).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		serviceBrokerListener := NewAddonsConfiguration(nil, filterAddonsConfigurationTrue, converter)

		// when
		serviceBrokerListener.OnDelete(serviceBroker)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		serviceBrokerListener := NewAddonsConfiguration(nil, filterAddonsConfigurationTrue, nil)

		// when
		serviceBrokerListener.OnDelete(new(struct{}))
	})

}

func TestAddonsConfiguration_OnUpdate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlClusterServiceBroker := new(gqlschema.AddonsConfiguration)
		cfg := new(v1alpha1.AddonsConfiguration)
		converter := automock.NewGQLAddonsConfigurationConverter()

		channel := make(chan gqlschema.AddonsConfigurationEvent, 1)
		defer close(channel)
		converter.On("ToGQL", cfg).Return(gqlClusterServiceBroker, nil).Once()
		defer converter.AssertExpectations(t)
		serviceBrokerListener := NewAddonsConfiguration(channel, filterAddonsConfigurationTrue, converter)

		// when
		serviceBrokerListener.OnUpdate(cfg, cfg)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeUpdate, result.Type)
		assert.Equal(t, *gqlClusterServiceBroker, result.AddonsConfiguration)

	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		serviceBrokerListener := NewAddonsConfiguration(nil, filterAddonsConfigurationFalse, nil)

		// when
		serviceBrokerListener.OnUpdate(new(v1alpha1.AddonsConfiguration), new(v1alpha1.AddonsConfiguration))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		serviceBrokerListener := NewAddonsConfiguration(nil, filterAddonsConfigurationTrue, nil)

		// when
		serviceBrokerListener.OnUpdate(nil, nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		serviceBroker := new(v1alpha1.AddonsConfiguration)
		converter := automock.NewGQLAddonsConfigurationConverter()

		converter.On("ToGQL", serviceBroker).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		serviceBrokerListener := NewAddonsConfiguration(nil, filterAddonsConfigurationTrue, converter)

		// when
		serviceBrokerListener.OnUpdate(nil, serviceBroker)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		serviceBrokerListener := NewAddonsConfiguration(nil, filterAddonsConfigurationTrue, nil)

		// when
		serviceBrokerListener.OnUpdate(new(struct{}), new(struct{}))
	})

}

func filterAddonsConfigurationTrue(o *v1alpha1.AddonsConfiguration) bool {
	return true
}

func filterAddonsConfigurationFalse(o *v1alpha1.AddonsConfiguration) bool {
	return false
}
