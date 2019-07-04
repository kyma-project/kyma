package listener

import (
	"testing"

	"github.com/kyma-project/kyma/components/helm-broker/pkg/apis/addons/v1alpha1"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons/listener/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
)

func TestClusterAddonsConfiguration_OnAdd(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlAddonsConfiguration := new(gqlschema.AddonsConfiguration)
		cfg := new(v1alpha1.ClusterAddonsConfiguration)
		converter := automock.NewGQLClusterAddonsConfigurationConverter()

		channel := make(chan gqlschema.AddonsConfigurationEvent, 1)
		defer close(channel)
		converter.On("ToGQL", cfg).Return(gqlAddonsConfiguration, nil).Once()
		defer converter.AssertExpectations(t)
		serviceBrokerListener := NewClusterAddonsConfiguration(channel, filterClusterAddonsConfigurationTrue, converter)

		// when
		serviceBrokerListener.OnAdd(cfg)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeAdd, result.Type)
		assert.Equal(t, *gqlAddonsConfiguration, result.AddonsConfiguration)
	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		serviceBrokerListener := NewClusterAddonsConfiguration(nil, filterClusterAddonsConfigurationFalse, nil)

		// when
		serviceBrokerListener.OnAdd(new(v1alpha1.ClusterAddonsConfiguration))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		serviceBrokerListener := NewClusterAddonsConfiguration(nil, filterClusterAddonsConfigurationTrue, nil)

		// when
		serviceBrokerListener.OnAdd(nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		serviceBroker := new(v1alpha1.ClusterAddonsConfiguration)
		converter := automock.NewGQLClusterAddonsConfigurationConverter()

		converter.On("ToGQL", serviceBroker).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		serviceBrokerListener := NewClusterAddonsConfiguration(nil, filterClusterAddonsConfigurationTrue, converter)

		// when
		serviceBrokerListener.OnAdd(serviceBroker)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		serviceBrokerListener := NewClusterAddonsConfiguration(nil, filterClusterAddonsConfigurationTrue, nil)

		// when
		serviceBrokerListener.OnAdd(new(struct{}))
	})

}

func TestClusterAddonsConfiguration_OnDelete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlClusterServiceBroker := new(gqlschema.AddonsConfiguration)
		serviceBroker := new(v1alpha1.ClusterAddonsConfiguration)
		converter := automock.NewGQLClusterAddonsConfigurationConverter()

		channel := make(chan gqlschema.AddonsConfigurationEvent, 1)
		defer close(channel)
		converter.On("ToGQL", serviceBroker).Return(gqlClusterServiceBroker, nil).Once()
		defer converter.AssertExpectations(t)
		serviceBrokerListener := NewClusterAddonsConfiguration(channel, filterClusterAddonsConfigurationTrue, converter)

		// when
		serviceBrokerListener.OnDelete(serviceBroker)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeDelete, result.Type)
		assert.Equal(t, *gqlClusterServiceBroker, result.AddonsConfiguration)

	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		serviceBrokerListener := NewClusterAddonsConfiguration(nil, filterClusterAddonsConfigurationFalse, nil)

		// when
		serviceBrokerListener.OnDelete(new(v1alpha1.ClusterAddonsConfiguration))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		serviceBrokerListener := NewClusterAddonsConfiguration(nil, filterClusterAddonsConfigurationTrue, nil)

		// when
		serviceBrokerListener.OnDelete(nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		serviceBroker := new(v1alpha1.ClusterAddonsConfiguration)
		converter := automock.NewGQLClusterAddonsConfigurationConverter()

		converter.On("ToGQL", serviceBroker).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		serviceBrokerListener := NewClusterAddonsConfiguration(nil, filterClusterAddonsConfigurationTrue, converter)

		// when
		serviceBrokerListener.OnDelete(serviceBroker)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		serviceBrokerListener := NewClusterAddonsConfiguration(nil, filterClusterAddonsConfigurationTrue, nil)

		// when
		serviceBrokerListener.OnDelete(new(struct{}))
	})

}

func TestClusterAddonsConfiguration_OnUpdate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlClusterServiceBroker := new(gqlschema.AddonsConfiguration)
		configMap := new(v1alpha1.ClusterAddonsConfiguration)
		converter := automock.NewGQLClusterAddonsConfigurationConverter()

		channel := make(chan gqlschema.AddonsConfigurationEvent, 1)
		defer close(channel)
		converter.On("ToGQL", configMap).Return(gqlClusterServiceBroker, nil).Once()
		defer converter.AssertExpectations(t)
		serviceBrokerListener := NewClusterAddonsConfiguration(channel, filterClusterAddonsConfigurationTrue, converter)

		// when
		serviceBrokerListener.OnUpdate(configMap, configMap)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeUpdate, result.Type)
		assert.Equal(t, *gqlClusterServiceBroker, result.AddonsConfiguration)

	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		serviceBrokerListener := NewClusterAddonsConfiguration(nil, filterClusterAddonsConfigurationFalse, nil)

		// when
		serviceBrokerListener.OnUpdate(new(v1alpha1.ClusterAddonsConfiguration), new(v1alpha1.ClusterAddonsConfiguration))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		serviceBrokerListener := NewClusterAddonsConfiguration(nil, filterClusterAddonsConfigurationTrue, nil)

		// when
		serviceBrokerListener.OnUpdate(nil, nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		serviceBroker := new(v1alpha1.ClusterAddonsConfiguration)
		converter := automock.NewGQLClusterAddonsConfigurationConverter()

		converter.On("ToGQL", serviceBroker).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		serviceBrokerListener := NewClusterAddonsConfiguration(nil, filterClusterAddonsConfigurationTrue, converter)

		// when
		serviceBrokerListener.OnUpdate(nil, serviceBroker)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		serviceBrokerListener := NewClusterAddonsConfiguration(nil, filterClusterAddonsConfigurationTrue, nil)

		// when
		serviceBrokerListener.OnUpdate(new(struct{}), new(struct{}))
	})

}

func filterClusterAddonsConfigurationTrue(o *v1alpha1.ClusterAddonsConfiguration) bool {
	return true
}

func filterClusterAddonsConfigurationFalse(o *v1alpha1.ClusterAddonsConfiguration) bool {
	return false
}
