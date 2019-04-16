package cms_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
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

func TestClusterDocsTopicResolver_ClusterDocsTopicsQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		resource :=
			&v1alpha1.ClusterDocsTopic{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
			}
		resources := []*v1alpha1.ClusterDocsTopic{
			resource, resource,
		}
		expected := []gqlschema.ClusterDocsTopic{
			{
				Name: "Test",
			}, {
				Name: "Test",
			},
		}

		svc := automock.NewClusterDocsTopicService()
		svc.On("List", (*string)(nil), (*string)(nil)).Return(resources, nil).Once()
		defer svc.AssertExpectations(t)

		converter := automock.NewGQLClusterDocsTopicConverter()
		converter.On("ToGQLs", resources).Return(expected, nil)
		defer converter.AssertExpectations(t)

		resolver := cms.NewClusterDocsTopicResolver(svc, nil)
		resolver.SetDocsTopicConverter(converter)

		result, err := resolver.ClusterDocsTopicsQuery(nil, nil, nil)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		var resources []*v1alpha1.ClusterDocsTopic

		svc := automock.NewClusterDocsTopicService()
		svc.On("List", (*string)(nil), (*string)(nil)).Return(resources, nil).Once()
		defer svc.AssertExpectations(t)
		resolver := cms.NewClusterDocsTopicResolver(svc, nil)
		var expected []gqlschema.ClusterDocsTopic

		result, err := resolver.ClusterDocsTopicsQuery(nil, nil, nil)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Error", func(t *testing.T) {
		expected := errors.New("Test")

		var resources []*v1alpha1.ClusterDocsTopic

		svc := automock.NewClusterDocsTopicService()
		svc.On("List", (*string)(nil), (*string)(nil)).Return(resources, expected).Once()
		defer svc.AssertExpectations(t)
		resolver := cms.NewClusterDocsTopicResolver(svc, nil)

		_, err := resolver.ClusterDocsTopicsQuery(nil, nil, nil)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})
}

func TestClusterDocsTopicResolver_ClusterDocsTopicAssetsField(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		docsTopicName := "exampleDocsTopic"
		resources := []*v1alpha2.ClusterAsset{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "ExampleClusterAssetA",
					Labels: map[string]string{
						cms.DocsTopicLabel: docsTopicName,
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "ExampleClusterAssetB",
					Labels: map[string]string{
						cms.DocsTopicLabel: docsTopicName,
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "ExampleClusterAssetC",
					Labels: map[string]string{
						cms.DocsTopicLabel: docsTopicName,
					},
				},
			},
		}
		expected := []gqlschema.ClusterAsset{
			{
				Name: "ExampleClusterAssetA",
			},
			{
				Name: "ExampleClusterAssetB",
			},
			{
				Name: "ExampleClusterAssetC",
			},
		}

		resourceGetter := new(assetstoreMock.ClusterAssetGetter)
		resourceGetter.On("ListForDocsTopicByType", docsTopicName, []string{}).Return(resources, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		resourceConverter := new(assetstoreMock.GqlClusterAssetConverter)
		resourceConverter.On("ToGQLs", resources).Return(expected, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		retriever := new(assetstoreMock.AssetStoreRetriever)
		retriever.On("ClusterAsset").Return(resourceGetter)
		retriever.On("ClusterAssetConverter").Return(resourceConverter)

		parentObj := gqlschema.ClusterDocsTopic{
			Name: docsTopicName,
		}

		resolver := cms.NewClusterDocsTopicResolver(nil, retriever)

		result, err := resolver.ClusterDocsTopicAssetsField(nil, &parentObj, []string{})

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		docsTopicName := "exampleDocsTopic"

		resourceGetter := new(assetstoreMock.ClusterAssetGetter)
		resourceGetter.On("ListForDocsTopicByType", docsTopicName, []string{}).Return(nil, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		resourceConverter := new(assetstoreMock.GqlClusterAssetConverter)
		resourceConverter.On("ToGQLs", ([]*v1alpha2.ClusterAsset)(nil)).Return(nil, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		retriever := new(assetstoreMock.AssetStoreRetriever)
		retriever.On("ClusterAsset").Return(resourceGetter)
		retriever.On("ClusterAssetConverter").Return(resourceConverter)

		parentObj := gqlschema.ClusterDocsTopic{
			Name: docsTopicName,
		}

		resolver := cms.NewClusterDocsTopicResolver(nil, retriever)

		result, err := resolver.ClusterDocsTopicAssetsField(nil, &parentObj, []string{})

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		expectedErr := errors.New("Test")
		docsTopicName := "exampleDocsTopic"

		resourceGetter := new(assetstoreMock.ClusterAssetGetter)
		resourceGetter.On("ListForDocsTopicByType", docsTopicName, []string{}).Return(nil, expectedErr).Once()
		defer resourceGetter.AssertExpectations(t)

		retriever := new(assetstoreMock.AssetStoreRetriever)
		retriever.On("ClusterAsset").Return(resourceGetter)

		parentObj := gqlschema.ClusterDocsTopic{
			Name: docsTopicName,
		}

		resolver := cms.NewClusterDocsTopicResolver(nil, retriever)

		result, err := resolver.ClusterDocsTopicAssetsField(nil, &parentObj, []string{})

		assert.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

func TestClusterDocsTopicResolver_ClusterDocsTopicEventSubscription(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), (-24 * time.Hour))
		cancel()

		svc := automock.NewClusterDocsTopicService()
		svc.On("Subscribe", mock.Anything).Once()
		svc.On("Unsubscribe", mock.Anything).Once()
		resolver := cms.NewClusterDocsTopicResolver(svc, nil)

		_, err := resolver.ClusterDocsTopicEventSubscription(ctx)

		require.NoError(t, err)
		svc.AssertCalled(t, "Subscribe", mock.Anything)
	})

	t.Run("Unsubscribe after connection close", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), (-24 * time.Hour))
		cancel()

		svc := automock.NewClusterDocsTopicService()
		svc.On("Subscribe", mock.Anything).Once()
		svc.On("Unsubscribe", mock.Anything).Once()
		resolver := cms.NewClusterDocsTopicResolver(svc, nil)

		channel, err := resolver.ClusterDocsTopicEventSubscription(ctx)
		<-channel

		require.NoError(t, err)
		svc.AssertCalled(t, "Unsubscribe", mock.Anything)
	})
}
