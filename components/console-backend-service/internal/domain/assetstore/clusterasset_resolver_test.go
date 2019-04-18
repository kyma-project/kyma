package assetstore_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/assetstore"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/assetstore/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestClusterAssetResolver_ClusterAssetFilesField(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		assetName := "exampleClusterAsset"
		rawMap := fixRawMap(t)

		clusterAssetResource := &v1alpha2.ClusterAsset{
			ObjectMeta: metav1.ObjectMeta{
				Name: assetName,
			},
			Status: v1alpha2.ClusterAssetStatus{
				CommonAssetStatus: v1alpha2.CommonAssetStatus{
					AssetRef: v1alpha2.AssetStatusRef{
						BaseURL: "https://example.com",
						Files: []v1alpha2.AssetFile{
							{
								Name:     "markdown.md",
								Metadata: rawMap,
							},
							{
								Name:     "apiSpec.json",
								Metadata: rawMap,
							},
							{
								Name:     "odata.xml",
								Metadata: rawMap,
							},
						},
					},
				},
			},
		}
		filesResource := []*assetstore.File{
			{
				URL:      "https://example.com/markdown.md",
				Metadata: rawMap,
			},
			{
				URL:      "https://example.com/apiSpec.json",
				Metadata: rawMap,
			},
			{
				URL:      "https://example.com/odata.xml",
				Metadata: rawMap,
			},
		}
		expected := []gqlschema.File{
			{
				URL: "https://example.com/markdown.md",
				Metadata: gqlschema.JSON{
					"labels": []interface{}{"test1", "test2"},
				},
			},
			{
				URL: "https://example.com/apiSpec.json",
				Metadata: gqlschema.JSON{
					"labels": []interface{}{"test1", "test2"},
				},
			},
			{
				URL: "https://example.com/odata.xml",
				Metadata: gqlschema.JSON{
					"labels": []interface{}{"test1", "test2"},
				},
			},
		}

		assetSvc := automock.NewClusterAssetService()
		assetSvc.On("Find", assetName).Return(clusterAssetResource, nil).Once()
		defer assetSvc.AssertExpectations(t)

		fileSvc := automock.NewFileService()
		fileSvc.On("Extract", &clusterAssetResource.Status.AssetRef).Return(filesResource, nil).Once()
		defer fileSvc.AssertExpectations(t)

		converter := automock.NewGQLFileConverter()
		converter.On("ToGQLs", filesResource).Return(expected, nil)
		defer converter.AssertExpectations(t)

		parentObj := gqlschema.ClusterAsset{
			Name: assetName,
		}

		resolver := assetstore.NewClusterAssetResolver(assetSvc)
		resolver.SetFileService(fileSvc)
		resolver.SetFileConverter(converter)

		result, err := resolver.ClusterAssetFilesField(nil, &parentObj, []string{})

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		assetName := "exampleClusterAsset"

		assetSvc := automock.NewClusterAssetService()
		assetSvc.On("Find", assetName).Return(nil, nil).Once()
		defer assetSvc.AssertExpectations(t)

		parentObj := gqlschema.ClusterAsset{
			Name: assetName,
		}

		resolver := assetstore.NewClusterAssetResolver(assetSvc)

		result, err := resolver.ClusterAssetFilesField(nil, &parentObj, []string{})

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		expectedErr := errors.New("Test")
		assetName := "exampleClusterAsset"

		assetSvc := automock.NewClusterAssetService()
		assetSvc.On("Find", assetName).Return(nil, expectedErr).Once()
		defer assetSvc.AssertExpectations(t)

		parentObj := gqlschema.ClusterAsset{
			Name: assetName,
		}

		resolver := assetstore.NewClusterAssetResolver(assetSvc)

		result, err := resolver.ClusterAssetFilesField(nil, &parentObj, []string{})

		assert.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

func TestClusterAssetResolver_ClusterAssetEventSubscription(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), (-24 * time.Hour))
		cancel()

		svc := automock.NewClusterAssetService()
		svc.On("Subscribe", mock.Anything).Once()
		svc.On("Unsubscribe", mock.Anything).Once()
		resolver := assetstore.NewClusterAssetResolver(svc)

		_, err := resolver.ClusterAssetEventSubscription(ctx)

		require.NoError(t, err)
		svc.AssertCalled(t, "Subscribe", mock.Anything)
	})

	t.Run("Unsubscribe after connection close", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), (-24 * time.Hour))
		cancel()

		svc := automock.NewClusterAssetService()
		svc.On("Subscribe", mock.Anything).Once()
		svc.On("Unsubscribe", mock.Anything).Once()
		resolver := assetstore.NewClusterAssetResolver(svc)

		channel, err := resolver.ClusterAssetEventSubscription(ctx)
		<-channel

		require.NoError(t, err)
		svc.AssertCalled(t, "Unsubscribe", mock.Anything)
	})
}
