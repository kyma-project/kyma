package listener_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/apigateway/listener"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/apigateway/listener/automock"

	"github.com/kyma-incubator/api-gateway/api/v1alpha1"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
)

func filterAll(_ *v1alpha1.APIRule) bool {
	return true
}
func filterNone(_ *v1alpha1.APIRule) bool {
	return false
}
func TestApiRuleListener_OnAdd(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlApiRule := new(gqlschema.APIRule)
		extractor := automock.NewExtractor()
		unstructuredApiRule := new(unstructured.Unstructured)

		apiRule := new(v1alpha1.APIRule)
		extractor.On("Do", unstructuredApiRule).Return(apiRule, nil).Once()

		converter := automock.NewGqlApiRuleConverter()

		channel := make(chan gqlschema.ApiRuleEvent, 1)
		defer close(channel)
		converter.On("ToGQL", apiRule).Return(gqlApiRule, nil).Once()
		defer converter.AssertExpectations(t)
		defer extractor.AssertExpectations(t)
		apiRuleListener := listener.NewApiRule(channel, filterAll, converter, extractor)

		// when
		apiRuleListener.OnAdd(unstructuredApiRule)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeAdd, result.Type)
		assert.Equal(t, *gqlApiRule, result.APIRule)
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		extractor := automock.NewExtractor()

		extractor.On("Do", nil).Return(nil, nil).Once()

		defer extractor.AssertExpectations(t)
		apiRuleListener := listener.NewApiRule(nil, filterAll, nil, extractor)

		// when
		apiRuleListener.OnAdd(nil)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		extractor := automock.NewExtractor()

		invalid := new(struct{})
		extractor.On("Do", invalid).Return(nil, errors.New("Error")).Once()

		defer extractor.AssertExpectations(t)
		apiRuleListener := listener.NewApiRule(nil, filterAll, nil, extractor)

		// when
		apiRuleListener.OnAdd(invalid)
	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		extractor := automock.NewExtractor()
		unstructuredApiRule := new(unstructured.Unstructured)

		apiRule := new(v1alpha1.APIRule)
		extractor.On("Do", unstructuredApiRule).Return(apiRule, nil).Once()

		defer extractor.AssertExpectations(t)
		apiRuleListener := listener.NewApiRule(nil, filterNone, nil, extractor)

		// when
		apiRuleListener.OnAdd(unstructuredApiRule)
	})
}

func TestApiRuleListener_OnDelete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlApiRule := new(gqlschema.APIRule)
		extractor := automock.NewExtractor()
		unstructuredApiRule := new(unstructured.Unstructured)

		apiRule := new(v1alpha1.APIRule)
		extractor.On("Do", unstructuredApiRule).Return(apiRule, nil).Once()

		converter := automock.NewGqlApiRuleConverter()

		channel := make(chan gqlschema.ApiRuleEvent, 1)
		defer close(channel)
		converter.On("ToGQL", apiRule).Return(gqlApiRule, nil).Once()
		defer converter.AssertExpectations(t)
		defer extractor.AssertExpectations(t)
		apiRuleListener := listener.NewApiRule(channel, filterAll, converter, extractor)

		// when
		apiRuleListener.OnDelete(unstructuredApiRule)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeDelete, result.Type)
		assert.Equal(t, *gqlApiRule, result.APIRule)

	})

	t.Run("Nil", func(t *testing.T) {
		// given
		extractor := automock.NewExtractor()
		extractor.On("Do", nil).Return(nil, nil).Once()

		defer extractor.AssertExpectations(t)
		apiRuleListener := listener.NewApiRule(nil, filterAll, nil, extractor)

		// when
		apiRuleListener.OnDelete(nil)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		extractor := automock.NewExtractor()
		invalid := new(struct{})
		extractor.On("Do", invalid).Return(nil, errors.New("Error")).Once()

		defer extractor.AssertExpectations(t)

		apiRuleListener := listener.NewApiRule(nil, filterAll, nil, extractor)

		// when
		apiRuleListener.OnDelete(invalid)
	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		extractor := automock.NewExtractor()
		unstructuredApiRule := new(unstructured.Unstructured)

		apiRule := new(v1alpha1.APIRule)
		extractor.On("Do", unstructuredApiRule).Return(apiRule, nil).Once()

		defer extractor.AssertExpectations(t)
		apiRuleListener := listener.NewApiRule(nil, filterNone, nil, extractor)

		// when
		apiRuleListener.OnDelete(unstructuredApiRule)
	})
}

func TestApiRuleListener_OnUpdate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlApiRule := new(gqlschema.APIRule)
		extractor := automock.NewExtractor()
		unstructuredApiRule := new(unstructured.Unstructured)

		apiRule := new(v1alpha1.APIRule)
		extractor.On("Do", unstructuredApiRule).Return(apiRule, nil).Once()

		converter := automock.NewGqlApiRuleConverter()

		channel := make(chan gqlschema.ApiRuleEvent, 1)
		defer close(channel)
		converter.On("ToGQL", apiRule).Return(gqlApiRule, nil).Once()
		defer converter.AssertExpectations(t)
		defer extractor.AssertExpectations(t)
		apiRuleListener := listener.NewApiRule(channel, filterAll, converter, extractor)

		// when
		apiRuleListener.OnUpdate(unstructuredApiRule, unstructuredApiRule)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeUpdate, result.Type)
		assert.Equal(t, *gqlApiRule, result.APIRule)

	})

	t.Run("Nil", func(t *testing.T) {
		// given
		extractor := automock.NewExtractor()

		extractor.On("Do", nil).Return(nil, nil).Once()

		defer extractor.AssertExpectations(t)
		apiRuleListener := listener.NewApiRule(nil, filterAll, nil, extractor)

		// when
		apiRuleListener.OnUpdate(nil, nil)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		extractor := automock.NewExtractor()
		invalid := new(struct{})

		extractor.On("Do", invalid).Return(nil, errors.New("Error")).Once()

		defer extractor.AssertExpectations(t)
		apiRuleListener := listener.NewApiRule(nil, filterAll, nil, extractor)

		// when
		apiRuleListener.OnUpdate(invalid, invalid)
	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		extractor := automock.NewExtractor()
		unstructuredApiRule := new(unstructured.Unstructured)

		apiRule := new(v1alpha1.APIRule)
		extractor.On("Do", unstructuredApiRule).Return(apiRule, nil).Once()

		defer extractor.AssertExpectations(t)
		apiRuleListener := listener.NewApiRule(nil, filterNone, nil, extractor)

		// when
		apiRuleListener.OnUpdate(unstructuredApiRule, unstructuredApiRule)
	})
}
