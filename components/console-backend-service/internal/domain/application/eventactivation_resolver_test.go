package application_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/application/automock"
	contentMock "github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared/automock"

	"github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/application"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/content/storage"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
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

		resolver := application.NewEventActivationResolver(svc, nil, nil)
		result, err := resolver.EventActivationsQuery(nil, "test")

		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Contains(t, result, *fixGQLEventActivation("event1"))
		assert.Contains(t, result, *fixGQLEventActivation("event2"))
	})

	t.Run("Not found", func(t *testing.T) {
		svc := automock.NewEventActivationLister()
		svc.On("List", "test").Return([]*v1alpha1.EventActivation{}, nil)
		defer svc.AssertExpectations(t)

		resolver := application.NewEventActivationResolver(svc, nil, nil)
		result, err := resolver.EventActivationsQuery(nil, "test")

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		svc := automock.NewEventActivationLister()
		svc.On("List", "test").Return(nil, errors.New("trol"))
		defer svc.AssertExpectations(t)

		resolver := application.NewEventActivationResolver(svc, nil, nil)
		_, err := resolver.EventActivationsQuery(nil, "test")

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
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

		getter := new(contentMock.AsyncApiSpecGetter)
		getter.On("Find", "service-class", "test").Return(asyncApiSpec, nil)
		defer getter.AssertExpectations(t)

		retriever := new(contentMock.ContentRetriever)
		retriever.On("AsyncApiSpec").Return(getter)

		resolver := application.NewEventActivationResolver(nil, retriever, nil)
		result, err := resolver.EventActivationEventsField(nil, fixGQLEventActivation("test"))

		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Contains(t, result, *fixGQLEventActivationEvent("sell", "v1", "desc"))
		assert.Contains(t, result, *fixGQLEventActivationEvent("sell", "v2", "desc"))
	})

	t.Run("Not found", func(t *testing.T) {
		types := []string{"asyncapi", "asyncApi", "asyncapispec", "asyncApiSpec", "events"}

		asyncApiSpecGetter := new(contentMock.AsyncApiSpecGetter)
		asyncApiSpecGetter.On("Find", "service-class", "test").Return(nil, nil)
		defer asyncApiSpecGetter.AssertExpectations(t)

		contentRetriever := new(contentMock.ContentRetriever)
		contentRetriever.On("AsyncApiSpec").Return(asyncApiSpecGetter)

		clusterAssetGetter := new(contentMock.ClusterAssetGetter)
		clusterAssetGetter.On("ListForDocsTopicByType", "test", types).Return(nil, nil)
		defer asyncApiSpecGetter.AssertExpectations(t)

		assetStoreRetriever := new(contentMock.AssetStoreRetriever)
		assetStoreRetriever.On("ClusterAsset").Return(clusterAssetGetter)

		resolver := application.NewEventActivationResolver(nil, contentRetriever, assetStoreRetriever)
		result, err := resolver.EventActivationEventsField(nil, fixGQLEventActivation("test"))

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("Invalid version", func(t *testing.T) {
		asyncApiSpec := &storage.AsyncApiSpec{
			Data: storage.AsyncApiSpecData{
				AsyncAPI: "1.0.1",
			},
		}

		getter := new(contentMock.AsyncApiSpecGetter)
		getter.On("Find", "service-class", "test").Return(asyncApiSpec, nil)
		defer getter.AssertExpectations(t)

		retriever := new(contentMock.ContentRetriever)
		retriever.On("AsyncApiSpec").Return(getter)

		resolver := application.NewEventActivationResolver(nil, retriever, nil)
		_, err := resolver.EventActivationEventsField(nil, fixGQLEventActivation("test"))

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})

	t.Run("Nil", func(t *testing.T) {
		getter := new(contentMock.AsyncApiSpecGetter)

		retriever := new(contentMock.ContentRetriever)
		retriever.On("AsyncApiSpec").Return(getter)

		resolver := application.NewEventActivationResolver(nil, retriever, nil)
		_, err := resolver.EventActivationEventsField(nil, nil)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})

	t.Run("Error", func(t *testing.T) {
		getter := new(contentMock.AsyncApiSpecGetter)
		getter.On("Find", "service-class", "test").Return(nil, errors.New("nope"))
		defer getter.AssertExpectations(t)

		retriever := new(contentMock.ContentRetriever)
		retriever.On("AsyncApiSpec").Return(getter)

		resolver := application.NewEventActivationResolver(nil, retriever, nil)
		_, err := resolver.EventActivationEventsField(nil, fixGQLEventActivation("test"))

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})
}

func fixGQLEventActivation(name string) *gqlschema.EventActivation {
	return &gqlschema.EventActivation{
		Name:        name,
		DisplayName: "aha!",
		SourceID:    "picco-bello",
	}
}

func fixGQLEventActivationEvent(eventType, version, desc string) *gqlschema.EventActivationEvent {
	return &gqlschema.EventActivationEvent{
		EventType:   eventType,
		Version:     version,
		Description: desc,
	}
}
