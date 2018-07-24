package remoteenvironment_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/remoteenvironment/v1alpha1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content/storage"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/remoteenvironment"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/remoteenvironment/automock"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventActivationResolver_EventActivationsQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		eventActivation1 := fixEventActivation("test", "event1")
		eventActivation2 := fixEventActivation("test", "event2")

		svc := automock.NewEventActivationLister()
		svc.On("List", "test").Return([]*v1alpha1.EventActivation{
			eventActivation1,
			eventActivation2,
		}, nil)
		defer svc.AssertExpectations(t)

		resolver := remoteenvironment.NewEventActivationResolver(svc, nil)
		result, err := resolver.EventActivationsQuery(nil, "test")

		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Contains(t, result, *fixGQLEventActivation("test", "event1"))
		assert.Contains(t, result, *fixGQLEventActivation("test", "event2"))
	})

	t.Run("Not found", func(t *testing.T) {
		svc := automock.NewEventActivationLister()
		svc.On("List", "test").Return([]*v1alpha1.EventActivation{}, nil)
		defer svc.AssertExpectations(t)

		resolver := remoteenvironment.NewEventActivationResolver(svc, nil)
		result, err := resolver.EventActivationsQuery(nil, "test")

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		svc := automock.NewEventActivationLister()
		svc.On("List", "test").Return(nil, errors.New("trol"))
		defer svc.AssertExpectations(t)

		resolver := remoteenvironment.NewEventActivationResolver(svc, nil)
		_, err := resolver.EventActivationsQuery(nil, "test")

		require.Error(t, err)
	})
}

func TestEventActivationResolver_EventActivationEventsField(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		asyncApiSpec := &storage.AsyncApiSpec{
			Data: storage.AsyncApiSpecData{
				AsyncAPI: "1.0.0",
				Topics: map[string]interface{}{
					"sell.v1": map[string]interface{}{
						"subscribe": map[string]interface{}{
							"summary": "desc",
						},
					},
					"sell.v2": map[string]interface{}{
						"subscribe": map[string]interface{}{
							"summary": "desc",
						},
					},
				},
			},
		}

		getter := new(automock.AsyncApiSpecGetter)
		getter.On("Find", "service-class", "test").Return(asyncApiSpec, nil)
		defer getter.AssertExpectations(t)

		resolver := remoteenvironment.NewEventActivationResolver(nil, getter)
		result, err := resolver.EventActivationEventsField(nil, fixGQLEventActivation("env", "test"))

		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Contains(t, result, *fixGQLEventActivationEvent("sell", "v1", "desc"))
		assert.Contains(t, result, *fixGQLEventActivationEvent("sell", "v2", "desc"))
	})

	t.Run("Not found", func(t *testing.T) {
		getter := new(automock.AsyncApiSpecGetter)
		getter.On("Find", "service-class", "test").Return(nil, nil)
		defer getter.AssertExpectations(t)

		resolver := remoteenvironment.NewEventActivationResolver(nil, getter)
		result, err := resolver.EventActivationEventsField(nil, fixGQLEventActivation("env", "test"))

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("Invalid version", func(t *testing.T) {
		asyncApiSpec := &storage.AsyncApiSpec{
			Data: storage.AsyncApiSpecData{
				AsyncAPI: "1.0.1",
			},
		}

		getter := new(automock.AsyncApiSpecGetter)
		getter.On("Find", "service-class", "test").Return(asyncApiSpec, nil)
		defer getter.AssertExpectations(t)

		resolver := remoteenvironment.NewEventActivationResolver(nil, getter)
		_, err := resolver.EventActivationEventsField(nil, fixGQLEventActivation("env", "test"))

		require.Error(t, err)
	})

	t.Run("Nil", func(t *testing.T) {
		getter := new(automock.AsyncApiSpecGetter)

		resolver := remoteenvironment.NewEventActivationResolver(nil, getter)
		_, err := resolver.EventActivationEventsField(nil, nil)

		require.Error(t, err)
	})

	t.Run("Error", func(t *testing.T) {
		getter := new(automock.AsyncApiSpecGetter)
		getter.On("Find", "service-class", "test").Return(nil, errors.New("nope"))
		defer getter.AssertExpectations(t)

		resolver := remoteenvironment.NewEventActivationResolver(nil, getter)
		_, err := resolver.EventActivationEventsField(nil, fixGQLEventActivation("env", "test"))

		require.Error(t, err)
	})
}

func fixGQLEventActivation(environment, name string) *gqlschema.EventActivation {
	return &gqlschema.EventActivation{
		Name:        name,
		DisplayName: "aha!",
		Source: gqlschema.EventActivationSource{
			Namespace:   "com.sap.test",
			Type:        "taaa",
			Environment: environment,
		},
	}
}

func fixGQLEventActivationEvent(eventType, version, desc string) *gqlschema.EventActivationEvent {
	return &gqlschema.EventActivationEvent{
		EventType:   eventType,
		Version:     version,
		Description: desc,
	}
}
