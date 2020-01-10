package rafter_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestClusterAssetResolver_ClusterAssetFilesField(t *testing.T) {
	assetName := "exampleClusterAsset"

	t.Run("Success", func(t *testing.T) {
		rawMap := fixRawMap(t)

		clusterAssetResource := &v1beta1.ClusterAsset{
			ObjectMeta: metav1.ObjectMeta{
				Name: assetName,
			},
			Status: v1beta1.ClusterAssetStatus{
				CommonAssetStatus: v1beta1.CommonAssetStatus{
					AssetRef: *fixAssetStatusRef(rawMap),
				},
			},
		}
		filesResource := []*rafter.File{
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

		clusterAssetSvc := automock.NewClusterAssetService()
		clusterAssetSvc.On("Find", assetName).Return(clusterAssetResource, nil).Once()
		defer clusterAssetSvc.AssertExpectations(t)

		fileSvc := automock.NewFileService()
		fileSvc.On("Extract", &clusterAssetResource.Status.AssetRef).Return(filesResource, nil).Once()
		defer fileSvc.AssertExpectations(t)

		fileConverter := automock.NewGQLFileConverter()
		fileConverter.On("ToGQLs", filesResource).Return(expected, nil)
		defer fileConverter.AssertExpectations(t)

		parentObj := gqlschema.ClusterAsset{
			Name: assetName,
		}

		resolver := rafter.NewClusterAssetResolver(clusterAssetSvc, nil, fileSvc, fileConverter)

		result, err := resolver.ClusterAssetFilesField(nil, &parentObj, []string{})

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		clusterAssetSvc := automock.NewClusterAssetService()
		clusterAssetSvc.On("Find", assetName).Return(nil, nil).Once()
		defer clusterAssetSvc.AssertExpectations(t)

		parentObj := gqlschema.ClusterAsset{
			Name: assetName,
		}

		resolver := rafter.NewClusterAssetResolver(clusterAssetSvc, nil, nil, nil)

		result, err := resolver.ClusterAssetFilesField(nil, &parentObj, []string{})

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		expectedErr := errors.New("Test")

		clusterAssetSvc := automock.NewClusterAssetService()
		clusterAssetSvc.On("Find", assetName).Return(nil, expectedErr).Once()
		defer clusterAssetSvc.AssertExpectations(t)

		parentObj := gqlschema.ClusterAsset{
			Name: assetName,
		}

		resolver := rafter.NewClusterAssetResolver(clusterAssetSvc, nil, nil, nil)

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
		resolver := rafter.NewClusterAssetResolver(svc, nil, nil, nil)

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
		resolver := rafter.NewClusterAssetResolver(svc, nil, nil, nil)

		channel, err := resolver.ClusterAssetEventSubscription(ctx)
		<-channel

		require.NoError(t, err)
		svc.AssertCalled(t, "Unsubscribe", mock.Anything)
	})
}
