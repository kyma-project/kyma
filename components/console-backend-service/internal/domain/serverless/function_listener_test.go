package serverless

import (
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/serverless/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestFunctionListener_OnAdd(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlFunction := new(gqlschema.Function)
		function := new(v1alpha1.Function)
		converter := automock.NewGQLFunctionConverter()

		channel := make(chan gqlschema.FunctionEvent, 1)
		defer close(channel)
		converter.On("ToGQL", function).Return(gqlFunction, nil).Once()
		defer converter.AssertExpectations(t)
		functionListener := newFunctionListener(channel, filterFunctionEventTrue, converter)

		// when
		functionListener.OnAdd(function)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeAdd, result.Type)
		assert.Equal(t, *gqlFunction, result.Function)
	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		functionListener := newFunctionListener(nil, filterFunctionEventFalse, nil)

		// when
		functionListener.OnAdd(new(v1alpha1.Function))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		functionListener := newFunctionListener(nil, filterFunctionEventTrue, nil)

		// when
		functionListener.OnAdd(nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		function := new(v1alpha1.Function)
		converter := automock.NewGQLFunctionConverter()

		converter.On("ToGQL", function).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		functionListener := newFunctionListener(nil, filterFunctionEventTrue, converter)

		// when
		functionListener.OnAdd(function)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		functionListener := newFunctionListener(nil, filterFunctionEventTrue, nil)

		// when
		functionListener.OnAdd(new(struct{}))
	})
}

func TestFunctionListener_OnDelete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlFunction := new(gqlschema.Function)
		function := new(v1alpha1.Function)
		converter := automock.NewGQLFunctionConverter()

		channel := make(chan gqlschema.FunctionEvent, 1)
		defer close(channel)
		converter.On("ToGQL", function).Return(gqlFunction, nil).Once()
		defer converter.AssertExpectations(t)
		functionListener := newFunctionListener(channel, filterFunctionEventTrue, converter)

		// when
		functionListener.OnDelete(function)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeDelete, result.Type)
		assert.Equal(t, *gqlFunction, result.Function)
	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		functionListener := newFunctionListener(nil, filterFunctionEventFalse, nil)

		// when
		functionListener.OnDelete(new(v1alpha1.Function))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		functionListener := newFunctionListener(nil, filterFunctionEventTrue, nil)

		// when
		functionListener.OnDelete(nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		function := new(v1alpha1.Function)
		converter := automock.NewGQLFunctionConverter()

		converter.On("ToGQL", function).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		functionListener := newFunctionListener(nil, filterFunctionEventTrue, converter)

		// when
		functionListener.OnDelete(function)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		functionListener := newFunctionListener(nil, filterFunctionEventTrue, nil)

		// when
		functionListener.OnDelete(new(struct{}))
	})
}

func TestFunctionListener_OnUpdate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlFunction := new(gqlschema.Function)
		function := new(v1alpha1.Function)
		converter := automock.NewGQLFunctionConverter()

		channel := make(chan gqlschema.FunctionEvent, 1)
		defer close(channel)
		converter.On("ToGQL", function).Return(gqlFunction, nil).Once()
		defer converter.AssertExpectations(t)
		functionListener := newFunctionListener(channel, filterFunctionEventTrue, converter)

		// when
		functionListener.OnUpdate(function, function)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeUpdate, result.Type)
		assert.Equal(t, *gqlFunction, result.Function)
	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		functionListener := newFunctionListener(nil, filterFunctionEventFalse, nil)

		// when
		functionListener.OnUpdate(new(v1alpha1.Function), new(v1alpha1.Function))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		functionListener := newFunctionListener(nil, filterFunctionEventTrue, nil)

		// when
		functionListener.OnUpdate(nil, nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		function := new(v1alpha1.Function)
		converter := automock.NewGQLFunctionConverter()

		converter.On("ToGQL", function).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		functionListener := newFunctionListener(nil, filterFunctionEventTrue, converter)

		// when
		functionListener.OnUpdate(function, function)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		functionListener := newFunctionListener(nil, filterFunctionEventTrue, nil)

		// when
		functionListener.OnUpdate(new(struct{}), new(struct{}))
	})
}

func filterFunctionEventTrue(o *v1alpha1.Function) bool {
	return true
}

func filterFunctionEventFalse(o *v1alpha1.Function) bool {
	return false
}
