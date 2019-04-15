package listener_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/cms/listener"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/cms/listener/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
)

func TestAsset_OnAdd(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlDocsTopic := new(gqlschema.DocsTopic)
		docsTopic := new(v1alpha1.DocsTopic)
		converter := automock.NewGQLDocsTopicConverter()

		channel := make(chan gqlschema.DocsTopicEvent, 1)
		defer close(channel)
		converter.On("ToGQL", docsTopic).Return(gqlDocsTopic, nil).Once()
		defer converter.AssertExpectations(t)
		docsTopicListener := listener.NewDocsTopic(channel, filterDocsTopicTrue, converter)

		// when
		docsTopicListener.OnAdd(docsTopic)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeAdd, result.Type)
		assert.Equal(t, *gqlDocsTopic, result.DocsTopic)
	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		docsTopicListener := listener.NewDocsTopic(nil, filterDocsTopicFalse, nil)

		// when
		docsTopicListener.OnAdd(new(v1alpha1.DocsTopic))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		docsTopicListener := listener.NewDocsTopic(nil, filterDocsTopicTrue, nil)

		// when
		docsTopicListener.OnAdd(nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		docsTopic := new(v1alpha1.DocsTopic)
		converter := automock.NewGQLDocsTopicConverter()

		converter.On("ToGQL", docsTopic).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		docsTopicListener := listener.NewDocsTopic(nil, filterDocsTopicTrue, converter)

		// when
		docsTopicListener.OnAdd(docsTopic)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		docsTopicListener := listener.NewDocsTopic(nil, filterDocsTopicTrue, nil)

		// when
		docsTopicListener.OnAdd(new(struct{}))
	})
}

func TestAsset_OnDelete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlDocsTopic := new(gqlschema.DocsTopic)
		docsTopic := new(v1alpha1.DocsTopic)
		converter := automock.NewGQLDocsTopicConverter()

		channel := make(chan gqlschema.DocsTopicEvent, 1)
		defer close(channel)
		converter.On("ToGQL", docsTopic).Return(gqlDocsTopic, nil).Once()
		defer converter.AssertExpectations(t)
		docsTopicListener := listener.NewDocsTopic(channel, filterDocsTopicTrue, converter)

		// when
		docsTopicListener.OnDelete(docsTopic)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeDelete, result.Type)
		assert.Equal(t, *gqlDocsTopic, result.DocsTopic)

	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		docsTopicListener := listener.NewDocsTopic(nil, filterDocsTopicFalse, nil)

		// when
		docsTopicListener.OnDelete(new(v1alpha1.DocsTopic))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		docsTopicListener := listener.NewDocsTopic(nil, filterDocsTopicTrue, nil)

		// when
		docsTopicListener.OnDelete(nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		docsTopic := new(v1alpha1.DocsTopic)
		converter := automock.NewGQLDocsTopicConverter()

		converter.On("ToGQL", docsTopic).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		docsTopicListener := listener.NewDocsTopic(nil, filterDocsTopicTrue, converter)

		// when
		docsTopicListener.OnDelete(docsTopic)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		docsTopicListener := listener.NewDocsTopic(nil, filterDocsTopicTrue, nil)

		// when
		docsTopicListener.OnDelete(new(struct{}))
	})
}

func TestAsset_OnUpdate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		gqlDocsTopic := new(gqlschema.DocsTopic)
		docsTopic := new(v1alpha1.DocsTopic)
		converter := automock.NewGQLDocsTopicConverter()

		channel := make(chan gqlschema.DocsTopicEvent, 1)
		defer close(channel)
		converter.On("ToGQL", docsTopic).Return(gqlDocsTopic, nil).Once()
		defer converter.AssertExpectations(t)
		docsTopicListener := listener.NewDocsTopic(channel, filterDocsTopicTrue, converter)

		// when
		docsTopicListener.OnUpdate(docsTopic, docsTopic)
		result := <-channel

		// then
		assert.Equal(t, gqlschema.SubscriptionEventTypeUpdate, result.Type)
		assert.Equal(t, *gqlDocsTopic, result.DocsTopic)

	})

	t.Run("Filtered out", func(t *testing.T) {
		// given
		docsTopicListener := listener.NewDocsTopic(nil, filterDocsTopicFalse, nil)

		// when
		docsTopicListener.OnUpdate(new(v1alpha1.DocsTopic), new(v1alpha1.DocsTopic))
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		docsTopicListener := listener.NewDocsTopic(nil, filterDocsTopicTrue, nil)

		// when
		docsTopicListener.OnUpdate(nil, nil)
	})

	t.Run("Nil GQL Type", func(t *testing.T) {
		// given
		docsTopic := new(v1alpha1.DocsTopic)
		converter := automock.NewGQLDocsTopicConverter()

		converter.On("ToGQL", docsTopic).Return(nil, nil).Once()
		defer converter.AssertExpectations(t)
		docsTopicListener := listener.NewDocsTopic(nil, filterDocsTopicTrue, converter)

		// when
		docsTopicListener.OnUpdate(nil, docsTopic)
	})

	t.Run("Invalid type", func(t *testing.T) {
		// given
		docsTopicListener := listener.NewDocsTopic(nil, filterDocsTopicTrue, nil)

		// when
		docsTopicListener.OnUpdate(new(struct{}), new(struct{}))
	})
}

func filterDocsTopicTrue(o *v1alpha1.DocsTopic) bool {
	return true
}

func filterDocsTopicFalse(o *v1alpha1.DocsTopic) bool {
	return false
}
