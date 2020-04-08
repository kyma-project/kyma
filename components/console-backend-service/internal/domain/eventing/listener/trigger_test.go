package listener_test

import (
	"testing"

	"github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/eventing/extractor"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/eventing/listener"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/eventing/listener/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"

	"github.com/stretchr/testify/assert"
)

func TestTrigger_OnAdd(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlTrigger := new(gqlschema.Trigger)
		trigger := new(v1alpha1.Trigger)
		converter := new(automock.Converter)

		channel := make(chan gqlschema.TriggerEvent, 1)
		defer close(channel)
		converter.On("ToGQL", trigger).Return(gqlTrigger, nil).Once()
		defer converter.AssertExpectations(t)
		extractor := extractor.TriggerUnstructuredExtractor{}
		triggerListener := listener.NewTrigger(extractor, channel, filterTriggerTrue, converter)

		// when
		triggerListener.OnAdd(trigger)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeAdd, result.Type)
		assert.Equal(t, *gqlTrigger, result.Trigger)
	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		extractor := extractor.TriggerUnstructuredExtractor{}
		triggerListener := listener.NewTrigger(extractor, nil, filterTriggerFalse, nil)

		// when
		triggerListener.OnAdd(new(v1alpha1.Trigger))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		extractor := extractor.TriggerUnstructuredExtractor{}
		triggerListener := listener.NewTrigger(extractor, nil, filterTriggerTrue, nil)

		// when
		triggerListener.OnAdd(nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		trigger := new(v1alpha1.Trigger)
		converter := new(automock.Converter)

		converter.On("ToGQL", trigger).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		extractor := extractor.TriggerUnstructuredExtractor{}
		triggerListener := listener.NewTrigger(extractor, nil, filterTriggerTrue, converter)

		// when
		triggerListener.OnAdd(trigger)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		extractor := extractor.TriggerUnstructuredExtractor{}
		triggerListener := listener.NewTrigger(extractor, nil, filterTriggerTrue, nil)

		// when
		triggerListener.OnAdd(new(struct{}))
	})
}

func TestTrigger_OnDelete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlTrigger := new(gqlschema.Trigger)
		trigger := new(v1alpha1.Trigger)
		converter := new(automock.Converter)

		channel := make(chan gqlschema.TriggerEvent, 1)
		defer close(channel)
		converter.On("ToGQL", trigger).Return(gqlTrigger, nil).Once()
		defer converter.AssertExpectations(t)
		extractor := extractor.TriggerUnstructuredExtractor{}
		triggerListener := listener.NewTrigger(extractor, channel, filterTriggerTrue, converter)

		// when
		triggerListener.OnDelete(trigger)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeDelete, result.Type)
		assert.Equal(t, *gqlTrigger, result.Trigger)

	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		extractor := extractor.TriggerUnstructuredExtractor{}
		triggerListener := listener.NewTrigger(extractor, nil, filterTriggerFalse, nil)

		// when
		triggerListener.OnDelete(new(v1alpha1.Trigger))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		extractor := extractor.TriggerUnstructuredExtractor{}
		triggerListener := listener.NewTrigger(extractor, nil, filterTriggerTrue, nil)

		// when
		triggerListener.OnDelete(nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		trigger := new(v1alpha1.Trigger)
		converter := new(automock.Converter)

		converter.On("ToGQL", trigger).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		extractor := extractor.TriggerUnstructuredExtractor{}
		triggerListener := listener.NewTrigger(extractor, nil, filterTriggerTrue, converter)

		// when
		triggerListener.OnDelete(trigger)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		extractor := extractor.TriggerUnstructuredExtractor{}
		triggerListener := listener.NewTrigger(extractor, nil, filterTriggerTrue, nil)

		// when
		triggerListener.OnDelete(new(struct{}))
	})
}

func TestTrigger_OnUpdate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlTrigger := new(gqlschema.Trigger)
		trigger := new(v1alpha1.Trigger)
		converter := new(automock.Converter)

		channel := make(chan gqlschema.TriggerEvent, 1)
		defer close(channel)
		converter.On("ToGQL", trigger).Return(gqlTrigger, nil).Once()
		defer converter.AssertExpectations(t)
		extractor := extractor.TriggerUnstructuredExtractor{}
		triggerListener := listener.NewTrigger(extractor, channel, filterTriggerTrue, converter)

		// when
		triggerListener.OnUpdate(trigger, trigger)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeUpdate, result.Type)
		assert.Equal(t, *gqlTrigger, result.Trigger)

	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		extractor := extractor.TriggerUnstructuredExtractor{}
		triggerListener := listener.NewTrigger(extractor, nil, filterTriggerFalse, nil)

		// when
		triggerListener.OnUpdate(new(v1alpha1.Trigger), new(v1alpha1.Trigger))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		extractor := extractor.TriggerUnstructuredExtractor{}
		triggerListener := listener.NewTrigger(extractor, nil, filterTriggerTrue, nil)

		// when
		triggerListener.OnUpdate(nil, nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		trigger := new(v1alpha1.Trigger)
		converter := new(automock.Converter)

		converter.On("ToGQL", trigger).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		extractor := extractor.TriggerUnstructuredExtractor{}
		triggerListener := listener.NewTrigger(extractor, nil, filterTriggerTrue, converter)

		// when
		triggerListener.OnUpdate(nil, trigger)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		extractor := extractor.TriggerUnstructuredExtractor{}
		triggerListener := listener.NewTrigger(extractor, nil, filterTriggerTrue, nil)

		// when
		triggerListener.OnUpdate(new(struct{}), new(struct{}))
	})
}

func filterTriggerTrue(o *v1alpha1.Trigger) bool {
	return true
}

func filterTriggerFalse(o *v1alpha1.Trigger) bool {
	return false
}
