package rafter_test

import (
	"context"
	"errors"
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter/automock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

func TestAssetResolver_AssetFilesField(t *testing.T) {
	assetName := "exampleAsset"
	namespace := "exampleNamespace"

	t.Run("Success", func(t *testing.T) {
		rawMap := fixRawMap(t)

		assetResource := &v1beta1.Asset{
			ObjectMeta: metav1.ObjectMeta{
				Name:      assetName,
				Namespace: namespace,
			},
			Status: v1beta1.AssetStatus{
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

		assetSvc := automock.NewAssetService()
		assetSvc.On("Find", namespace, assetName).Return(assetResource, nil).Once()
		defer assetSvc.AssertExpectations(t)

		fileSvc := automock.NewFileService()
		fileSvc.On("Extract", &assetResource.Status.AssetRef).Return(filesResource, nil).Once()
		defer fileSvc.AssertExpectations(t)

		fileConverter := automock.NewGQLFileConverter()
		fileConverter.On("ToGQLs", filesResource).Return(expected, nil)
		defer fileConverter.AssertExpectations(t)

		parentObj := gqlschema.Asset{
			Name:      assetName,
			Namespace: namespace,
		}

		resolver := rafter.NewAssetResolver(assetSvc, automock.NewGQLAssetConverter(), fileSvc, fileConverter)

		result, err := resolver.AssetFilesField(nil, &parentObj, []string{})

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		assetSvc := automock.NewAssetService()
		assetSvc.On("Find", namespace, assetName).Return(nil, nil).Once()
		defer assetSvc.AssertExpectations(t)

		parentObj := gqlschema.Asset{
			Name:      assetName,
			Namespace: namespace,
		}

		resolver := rafter.NewAssetResolver(assetSvc, automock.NewGQLAssetConverter(), automock.NewFileService(), automock.NewGQLFileConverter())

		result, err := resolver.AssetFilesField(nil, &parentObj, []string{})

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		expectedErr := errors.New("Test")

		assetSvc := automock.NewAssetService()
		assetSvc.On("Find", namespace, assetName).Return(nil, expectedErr).Once()
		defer assetSvc.AssertExpectations(t)

		parentObj := gqlschema.Asset{
			Name:      assetName,
			Namespace: namespace,
		}

		resolver := rafter.NewAssetResolver(assetSvc, automock.NewGQLAssetConverter(), automock.NewFileService(), automock.NewGQLFileConverter())

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
		resolver := rafter.NewAssetResolver(svc, automock.NewGQLAssetConverter(), automock.NewFileService(), automock.NewGQLFileConverter())

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
		resolver := rafter.NewAssetResolver(svc, automock.NewGQLAssetConverter(), automock.NewFileService(), automock.NewGQLFileConverter())

		channel, err := resolver.AssetEventSubscription(ctx, "test")
		<-channel

		require.NoError(t, err)
		svc.AssertCalled(t, "Unsubscribe", mock.Anything)
	})
}
