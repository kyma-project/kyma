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

func TestAssetResolver_AssetFilesField(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		assetName := "exampleAsset"
		namespace := "exampleNamespace"
		rawMap := fixRawMap(t)

		assetResource := &v1alpha2.Asset{
			ObjectMeta: metav1.ObjectMeta{
				Name:      assetName,
				Namespace: namespace,
			},
			Status: v1alpha2.AssetStatus{
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

		assetSvc := automock.NewAssetService()
		assetSvc.On("Find", namespace, assetName).Return(assetResource, nil).Once()
		defer assetSvc.AssertExpectations(t)

		fileSvc := automock.NewFileService()
		fileSvc.On("Extract", &assetResource.Status.AssetRef).Return(filesResource, nil).Once()
		defer fileSvc.AssertExpectations(t)

		converter := automock.NewGQLFileConverter()
		converter.On("ToGQLs", filesResource).Return(expected, nil)
		defer converter.AssertExpectations(t)

		parentObj := gqlschema.Asset{
			Name:      assetName,
			Namespace: namespace,
		}

		resolver := assetstore.NewAssetResolver(assetSvc)
		resolver.SetFileService(fileSvc)
		resolver.SetFileConverter(converter)

		result, err := resolver.AssetFilesField(nil, &parentObj, []string{})

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		assetName := "exampleClusterAsset"
		namespace := "exampleNamespace"

		assetSvc := automock.NewAssetService()
		assetSvc.On("Find", namespace, assetName).Return(nil, nil).Once()
		defer assetSvc.AssertExpectations(t)

		parentObj := gqlschema.Asset{
			Name:      assetName,
			Namespace: namespace,
		}

		resolver := assetstore.NewAssetResolver(assetSvc)

		result, err := resolver.AssetFilesField(nil, &parentObj, []string{})

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		expectedErr := errors.New("Test")
		assetName := "exampleClusterAsset"
		namespace := "exampleNamespace"

		assetSvc := automock.NewAssetService()
		assetSvc.On("Find", namespace, assetName).Return(nil, expectedErr).Once()
		defer assetSvc.AssertExpectations(t)

		parentObj := gqlschema.Asset{
			Name:      assetName,
			Namespace: namespace,
		}

		resolver := assetstore.NewAssetResolver(assetSvc)

		result, err := resolver.AssetFilesField(nil, &parentObj, []string{})

		assert.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

func TestAssetResolver_AssetEventSubscription(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), (-24 * time.Hour))
		cancel()

		svc := automock.NewAssetService()
		svc.On("Subscribe", mock.Anything).Once()
		svc.On("Unsubscribe", mock.Anything).Once()
		resolver := assetstore.NewAssetResolver(svc)

		_, err := resolver.AssetEventSubscription(ctx, "test")

		require.NoError(t, err)
		svc.AssertCalled(t, "Subscribe", mock.Anything)
	})

	t.Run("Unsubscribe after connection close", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), (-24 * time.Hour))
		cancel()

		svc := automock.NewAssetService()
		svc.On("Subscribe", mock.Anything).Once()
		svc.On("Unsubscribe", mock.Anything).Once()
		resolver := assetstore.NewAssetResolver(svc)

		channel, err := resolver.AssetEventSubscription(ctx, "test")
		<-channel

		require.NoError(t, err)
		svc.AssertCalled(t, "Unsubscribe", mock.Anything)
	})
}
