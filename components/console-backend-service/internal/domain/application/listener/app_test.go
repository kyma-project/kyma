package listener_test

import (
	"testing"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	api "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/application/listener"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/application/listener/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
)

func TestApplicationListener_OnAdd(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlApplication := new(gqlschema.Application)
		extractor := automock.NewExtractor()
		unstructuredApi := new(unstructured.Unstructured)

		application := new(api.Application)
		extractor.On("Do", unstructuredApi).Return(application, nil).Once()

		converter := automock.NewGQLApplicationConverter()

		channel := make(chan *gqlschema.ApplicationEvent, 1)
		defer close(channel)
		converter.On("ToGQL", application).Return(gqlApplication, nil).Once()
		defer converter.AssertExpectations(t)
		defer extractor.AssertExpectations(t)
		applicationListener := listener.NewApplication(channel, converter, extractor)

		// when
		applicationListener.OnAdd(unstructuredApi)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeAdd, result.Type)
		assert.Equal(t, gqlApplication, result.Application)
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		extractor := automock.NewExtractor()

		extractor.On("Do", nil).Return(nil, nil).Once()

		defer extractor.AssertExpectations(t)
		applicationListener := listener.NewApplication(nil, nil, extractor)

		// when
		applicationListener.OnAdd(nil)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		extractor := automock.NewExtractor()

		invalid := new(struct{})
		extractor.On("Do", invalid).Return(nil, errors.New("Error")).Once()

		defer extractor.AssertExpectations(t)
		applicationListener := listener.NewApplication(nil, nil, extractor)

		// when
		applicationListener.OnAdd(invalid)
	})
}

func TestApplicationListener_OnDelete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlApplication := new(gqlschema.Application)
		extractor := automock.NewExtractor()
		unstructuredApi := new(unstructured.Unstructured)

		application := new(api.Application)
		extractor.On("Do", unstructuredApi).Return(application, nil).Once()

		converter := automock.NewGQLApplicationConverter()

		channel := make(chan *gqlschema.ApplicationEvent, 1)
		defer close(channel)
		converter.On("ToGQL", application).Return(gqlApplication, nil).Once()
		defer converter.AssertExpectations(t)
		defer extractor.AssertExpectations(t)
		applicationListener := listener.NewApplication(channel, converter, extractor)

		// when
		applicationListener.OnDelete(unstructuredApi)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeDelete, result.Type)
		assert.Equal(t, gqlApplication, result.Application)

	})

	t.Run("Nil", func(t *testing.T) {
		// given
		extractor := automock.NewExtractor()
		extractor.On("Do", nil).Return(nil, nil).Once()

		defer extractor.AssertExpectations(t)
		applicationListener := listener.NewApplication(nil, nil, extractor)

		// when
		applicationListener.OnDelete(nil)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		extractor := automock.NewExtractor()
		invalid := new(struct{})
		extractor.On("Do", invalid).Return(nil, errors.New("Error")).Once()

		defer extractor.AssertExpectations(t)

		applicationListener := listener.NewApplication(nil, nil, extractor)

		// when
		applicationListener.OnDelete(invalid)
	})
}

func TestApplicationListener_OnUpdate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlApplication := new(gqlschema.Application)
		extractor := automock.NewExtractor()
		unstructuredApi := new(unstructured.Unstructured)

		application := new(api.Application)
		extractor.On("Do", unstructuredApi).Return(application, nil).Once()

		converter := automock.NewGQLApplicationConverter()

		channel := make(chan *gqlschema.ApplicationEvent, 1)
		defer close(channel)
		converter.On("ToGQL", application).Return(gqlApplication, nil).Once()
		defer converter.AssertExpectations(t)
		defer extractor.AssertExpectations(t)
		applicationListener := listener.NewApplication(channel, converter, extractor)

		// when
		applicationListener.OnUpdate(unstructuredApi, unstructuredApi)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeUpdate, result.Type)
		assert.Equal(t, gqlApplication, result.Application)

	})

	t.Run("Nil", func(t *testing.T) {
		// given
		extractor := automock.NewExtractor()

		extractor.On("Do", nil).Return(nil, nil).Once()

		defer extractor.AssertExpectations(t)
		applicationListener := listener.NewApplication(nil, nil, extractor)

		// when
		applicationListener.OnUpdate(nil, nil)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		extractor := automock.NewExtractor()
		invalid := new(struct{})

		extractor.On("Do", invalid).Return(nil, errors.New("Error")).Once()

		defer extractor.AssertExpectations(t)
		applicationListener := listener.NewApplication(nil, nil, extractor)

		// when
		applicationListener.OnUpdate(invalid, invalid)
	})
}
