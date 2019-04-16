package cms_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/cms"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/cms/automock"
	assetstoreMock "github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDocsTopicResolver_DocsTopicAssetsField(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		docsTopicName := "exampleDocsTopic"
		namespace := "exampleNamespace"
		resources := []*v1alpha2.Asset{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ExampleAssetA",
					Namespace: namespace,
					Labels: map[string]string{
						cms.DocsTopicLabel: docsTopicName,
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ExampleAssetB",
					Namespace: namespace,
					Labels: map[string]string{
						cms.DocsTopicLabel: docsTopicName,
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ExampleAssetC",
					Namespace: namespace,
					Labels: map[string]string{
						cms.DocsTopicLabel: docsTopicName,
					},
				},
			},
		}
		expected := []gqlschema.Asset{
			{
				Name:      "ExampleAssetA",
				Namespace: namespace,
			},
			{
				Name:      "ExampleAssetB",
				Namespace: namespace,
			},
			{
				Name:      "ExampleAssetC",
				Namespace: namespace,
			},
		}

		resourceGetter := new(assetstoreMock.AssetGetter)
		resourceGetter.On("ListForDocsTopicByType", namespace, docsTopicName, []string{}).Return(resources, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		resourceConverter := new(assetstoreMock.GqlAssetConverter)
		resourceConverter.On("ToGQLs", resources).Return(expected, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		retriever := new(assetstoreMock.AssetStoreRetriever)
		retriever.On("Asset").Return(resourceGetter)
		retriever.On("AssetConverter").Return(resourceConverter)

		parentObj := gqlschema.DocsTopic{
			Name:      docsTopicName,
			Namespace: namespace,
		}

		resolver := cms.NewDocsTopicResolver(nil, retriever)

		result, err := resolver.DocsTopicAssetsField(nil, &parentObj, []string{})

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		docsTopicName := "exampleDocsTopic"
		namespace := "exampleNamespace"

		resourceGetter := new(assetstoreMock.AssetGetter)
		resourceGetter.On("ListForDocsTopicByType", namespace, docsTopicName, []string{}).Return(nil, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		resourceConverter := new(assetstoreMock.GqlAssetConverter)
		resourceConverter.On("ToGQLs", ([]*v1alpha2.Asset)(nil)).Return(nil, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		retriever := new(assetstoreMock.AssetStoreRetriever)
		retriever.On("Asset").Return(resourceGetter)
		retriever.On("AssetConverter").Return(resourceConverter)

		parentObj := gqlschema.DocsTopic{
			Name:      docsTopicName,
			Namespace: namespace,
		}

		resolver := cms.NewDocsTopicResolver(nil, retriever)

		result, err := resolver.DocsTopicAssetsField(nil, &parentObj, []string{})

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		expectedErr := errors.New("Test")
		docsTopicName := "exampleDocsTopic"
		namespace := "exampleNamespace"

		resourceGetter := new(assetstoreMock.AssetGetter)
		resourceGetter.On("ListForDocsTopicByType", namespace, docsTopicName, []string{}).Return(nil, expectedErr).Once()
		defer resourceGetter.AssertExpectations(t)

		retriever := new(assetstoreMock.AssetStoreRetriever)
		retriever.On("Asset").Return(resourceGetter)

		parentObj := gqlschema.DocsTopic{
			Name:      docsTopicName,
			Namespace: namespace,
		}

		resolver := cms.NewDocsTopicResolver(nil, retriever)

		result, err := resolver.DocsTopicAssetsField(nil, &parentObj, []string{})

		assert.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

func TestDocsTopicResolver_DocsTopicEventSubscription(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), (-24 * time.Hour))
		cancel()

		svc := automock.NewDocsTopicService()
		svc.On("Subscribe", mock.Anything).Once()
		svc.On("Unsubscribe", mock.Anything).Once()
		resolver := cms.NewDocsTopicResolver(svc, nil)

		_, err := resolver.DocsTopicEventSubscription(ctx, "test")

		require.NoError(t, err)
		svc.AssertCalled(t, "Subscribe", mock.Anything)
	})

	t.Run("Unsubscribe after connection close", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), (-24 * time.Hour))
		cancel()

		svc := automock.NewDocsTopicService()
		svc.On("Subscribe", mock.Anything).Once()
		svc.On("Unsubscribe", mock.Anything).Once()
		resolver := cms.NewDocsTopicResolver(svc, nil)

		channel, err := resolver.DocsTopicEventSubscription(ctx, "test")
		<-channel

		require.NoError(t, err)
		svc.AssertCalled(t, "Unsubscribe", mock.Anything)
	})
}
