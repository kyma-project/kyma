package listener_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma-project.io/v1alpha2"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/apicontroller/listener"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/apicontroller/listener/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
)

func filterAll(_ *v1alpha2.Api) bool {
	return true
}
func filterNone(_ *v1alpha2.Api) bool {
	return false
}
func TestApiListener_OnAdd(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlApi := new(gqlschema.API)
		extractor := automock.NewExtractor()
		unstructuredApi := new(unstructured.Unstructured)

		api := new(v1alpha2.Api)
		extractor.On("Do", unstructuredApi).Return(api, nil).Once()

		converter := automock.NewGqlApiConverter()

		channel := make(chan gqlschema.ApiEvent, 1)
		defer close(channel)
		converter.On("ToGQL", api).Return(gqlApi, nil).Once()
		defer converter.AssertExpectations(t)
		defer extractor.AssertExpectations(t)
		apiListener := listener.NewApi(channel, filterAll, converter, extractor)

		// when
		apiListener.OnAdd(unstructuredApi)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeAdd, result.Type)
		assert.Equal(t, *gqlApi, result.API)
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		extractor := automock.NewExtractor()

		extractor.On("Do", nil).Return(nil, nil).Once()

		defer extractor.AssertExpectations(t)
		apiListener := listener.NewApi(nil, filterAll, nil, extractor)

		// when
		apiListener.OnAdd(nil)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		extractor := automock.NewExtractor()

		invalid := new(struct{})
		extractor.On("Do", invalid).Return(nil, errors.New("Error")).Once()

		defer extractor.AssertExpectations(t)
		apiListener := listener.NewApi(nil, filterAll, nil, extractor)

		// when
		apiListener.OnAdd(invalid)
	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		extractor := automock.NewExtractor()
		unstructuredApi := new(unstructured.Unstructured)

		api := new(v1alpha2.Api)
		extractor.On("Do", unstructuredApi).Return(api, nil).Once()

		defer extractor.AssertExpectations(t)
		apiListener := listener.NewApi(nil, filterNone, nil, extractor)

		// when
		apiListener.OnAdd(unstructuredApi)
	})
}

func TestApplicationListener_OnDelete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlApi := new(gqlschema.API)
		extractor := automock.NewExtractor()
		unstructuredApi := new(unstructured.Unstructured)

		api := new(v1alpha2.Api)
		extractor.On("Do", unstructuredApi).Return(api, nil).Once()

		converter := automock.NewGqlApiConverter()

		channel := make(chan gqlschema.ApiEvent, 1)
		defer close(channel)
		converter.On("ToGQL", api).Return(gqlApi, nil).Once()
		defer converter.AssertExpectations(t)
		defer extractor.AssertExpectations(t)
		apiListener := listener.NewApi(channel, filterAll, converter, extractor)

		// when
		apiListener.OnDelete(unstructuredApi)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeDelete, result.Type)
		assert.Equal(t, *gqlApi, result.API)

	})

	t.Run("Nil", func(t *testing.T) {
		// given
		extractor := automock.NewExtractor()
		extractor.On("Do", nil).Return(nil, nil).Once()

		defer extractor.AssertExpectations(t)
		apiListener := listener.NewApi(nil, filterAll, nil, extractor)

		// when
		apiListener.OnDelete(nil)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		extractor := automock.NewExtractor()
		invalid := new(struct{})
		extractor.On("Do", invalid).Return(nil, errors.New("Error")).Once()

		defer extractor.AssertExpectations(t)

		apiListener := listener.NewApi(nil, filterAll, nil, extractor)

		// when
		apiListener.OnDelete(invalid)
	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		extractor := automock.NewExtractor()
		unstructuredApi := new(unstructured.Unstructured)

		api := new(v1alpha2.Api)
		extractor.On("Do", unstructuredApi).Return(api, nil).Once()

		defer extractor.AssertExpectations(t)
		apiListener := listener.NewApi(nil, filterNone, nil, extractor)

		// when
		apiListener.OnDelete(unstructuredApi)
	})
}

func TestApplicationListener_OnUpdate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlApi := new(gqlschema.API)
		extractor := automock.NewExtractor()
		unstructuredApi := new(unstructured.Unstructured)

		api := new(v1alpha2.Api)
		extractor.On("Do", unstructuredApi).Return(api, nil).Once()

		converter := automock.NewGqlApiConverter()

		channel := make(chan gqlschema.ApiEvent, 1)
		defer close(channel)
		converter.On("ToGQL", api).Return(gqlApi, nil).Once()
		defer converter.AssertExpectations(t)
		defer extractor.AssertExpectations(t)
		apiListener := listener.NewApi(channel, filterAll, converter, extractor)

		// when
		apiListener.OnUpdate(unstructuredApi, unstructuredApi)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeUpdate, result.Type)
		assert.Equal(t, *gqlApi, result.API)

	})

	t.Run("Nil", func(t *testing.T) {
		// given
		extractor := automock.NewExtractor()

		extractor.On("Do", nil).Return(nil, nil).Once()

		defer extractor.AssertExpectations(t)
		apiListener := listener.NewApi(nil, filterAll, nil, extractor)

		// when
		apiListener.OnUpdate(nil, nil)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		extractor := automock.NewExtractor()
		invalid := new(struct{})

		extractor.On("Do", invalid).Return(nil, errors.New("Error")).Once()

		defer extractor.AssertExpectations(t)
		apiListener := listener.NewApi(nil, filterAll, nil, extractor)

		// when
		apiListener.OnUpdate(invalid, invalid)
	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		extractor := automock.NewExtractor()
		unstructuredApi := new(unstructured.Unstructured)

		api := new(v1alpha2.Api)
		extractor.On("Do", unstructuredApi).Return(api, nil).Once()

		defer extractor.AssertExpectations(t)
		apiListener := listener.NewApi(nil, filterNone, nil, extractor)

		// when
		apiListener.OnUpdate(unstructuredApi, unstructuredApi)
	})
}
