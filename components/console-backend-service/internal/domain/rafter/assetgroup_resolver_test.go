package rafter_test

import (
	"context"
	"errors"
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter/automock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

func TestAssetGroupResolver_AssetGroupAssetsField(t *testing.T) {
	assetGroupName := "exampleAssetGroup"
	namespace := "exampleNamespace"

	t.Run("Success", func(t *testing.T) {
		resources := []*v1beta1.Asset{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ExampleAssetA",
					Namespace: namespace,
					Labels: map[string]string{
						rafter.AssetGroupLabel: assetGroupName,
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ExampleAssetB",
					Namespace: namespace,
					Labels: map[string]string{
						rafter.AssetGroupLabel: assetGroupName,
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ExampleAssetC",
					Namespace: namespace,
					Labels: map[string]string{
						rafter.AssetGroupLabel: assetGroupName,
					},
				},
			},
		}
		expected := []gqlschema.RafterAsset{
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

		assetSvc := automock.NewAssetService()
		assetSvc.On("ListForAssetGroupByType", namespace, assetGroupName, []string{}).Return(resources, nil).Once()
		defer assetSvc.AssertExpectations(t)

		assetConverter := automock.NewGQLAssetConverter()
		assetConverter.On("ToGQLs", resources).Return(expected, nil).Once()
		defer assetConverter.AssertExpectations(t)

		parentObj := gqlschema.AssetGroup{
			Name:      assetGroupName,
			Namespace: namespace,
		}

		resolver := rafter.NewAssetGroupResolver(nil, nil, assetSvc, assetConverter)

		result, err := resolver.AssetGroupAssetsField(nil, &parentObj, []string{})

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		assetSvc := automock.NewAssetService()
		assetSvc.On("ListForAssetGroupByType", namespace, assetGroupName, []string{}).Return(nil, nil).Once()
		defer assetSvc.AssertExpectations(t)

		assetConverter := automock.NewGQLAssetConverter()
		assetConverter.On("ToGQLs", ([]*v1beta1.Asset)(nil)).Return(nil, nil).Once()
		defer assetConverter.AssertExpectations(t)

		parentObj := gqlschema.AssetGroup{
			Name:      assetGroupName,
			Namespace: namespace,
		}

		resolver := rafter.NewAssetGroupResolver(nil, nil, assetSvc, assetConverter)

		result, err := resolver.AssetGroupAssetsField(nil, &parentObj, []string{})

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		expectedErr := errors.New("Test")

		assetSvc := automock.NewAssetService()
		assetSvc.On("ListForAssetGroupByType", namespace, assetGroupName, []string{}).Return(nil, expectedErr).Once()
		defer assetSvc.AssertExpectations(t)

		parentObj := gqlschema.AssetGroup{
			Name:      assetGroupName,
			Namespace: namespace,
		}

		resolver := rafter.NewAssetGroupResolver(nil, nil, assetSvc, nil)

		result, err := resolver.AssetGroupAssetsField(nil, &parentObj, []string{})

		assert.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

func TestAssetGroupResolver_AssetGroupEventSubscription(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), (-24 * time.Hour))
		cancel()

		svc := automock.NewAssetGroupService()
		svc.On("Subscribe", mock.Anything).Once()
		svc.On("Unsubscribe", mock.Anything).Once()
		resolver := rafter.NewAssetGroupResolver(svc, automock.NewGQLAssetGroupConverter(), automock.NewAssetService(), automock.NewGQLAssetConverter())

		_, err := resolver.AssetGroupEventSubscription(ctx, "test")

		require.NoError(t, err)
		svc.AssertCalled(t, "Subscribe", mock.Anything)
	})

	t.Run("Unsubscribe after connection close", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), (-24 * time.Hour))
		cancel()

		svc := automock.NewAssetGroupService()
		svc.On("Subscribe", mock.Anything).Once()
		svc.On("Unsubscribe", mock.Anything).Once()
		resolver := rafter.NewAssetGroupResolver(svc, automock.NewGQLAssetGroupConverter(), automock.NewAssetService(), automock.NewGQLAssetConverter())

		channel, err := resolver.AssetGroupEventSubscription(ctx, "test")
		<-channel

		require.NoError(t, err)
		svc.AssertCalled(t, "Unsubscribe", mock.Anything)
	})
}
