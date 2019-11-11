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

func TestClusterAssetGroupResolver_ClusterAssetGroupsQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		resource :=
			&v1beta1.ClusterAssetGroup{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
			}
		resources := []*v1beta1.ClusterAssetGroup{
			resource, resource,
		}
		expected := []gqlschema.ClusterAssetGroup{
			{
				Name: "Test",
			}, {
				Name: "Test",
			},
		}

		svc := automock.NewClusterAssetGroupService()
		svc.On("List", (*string)(nil), (*string)(nil)).Return(resources, nil).Once()
		defer svc.AssertExpectations(t)

		converter := automock.NewGQLClusterAssetGroupConverter()
		converter.On("ToGQLs", resources).Return(expected, nil)
		defer converter.AssertExpectations(t)

		resolver := rafter.NewClusterAssetGroupResolver(svc, converter, nil, nil)

		result, err := resolver.ClusterAssetGroupsQuery(nil, nil, nil)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		var resources []*v1beta1.ClusterAssetGroup
		var expected []gqlschema.ClusterAssetGroup

		svc := automock.NewClusterAssetGroupService()
		svc.On("List", (*string)(nil), (*string)(nil)).Return(resources, nil).Once()
		defer svc.AssertExpectations(t)

		converter := automock.NewGQLClusterAssetGroupConverter()
		converter.On("ToGQLs", resources).Return(expected, nil)
		defer converter.AssertExpectations(t)

		resolver := rafter.NewClusterAssetGroupResolver(svc, converter, nil, nil)

		result, err := resolver.ClusterAssetGroupsQuery(nil, nil, nil)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Error", func(t *testing.T) {
		expected := errors.New("Test")
		var resources []*v1beta1.ClusterAssetGroup

		svc := automock.NewClusterAssetGroupService()
		svc.On("List", (*string)(nil), (*string)(nil)).Return(resources, expected).Once()
		defer svc.AssertExpectations(t)

		resolver := rafter.NewClusterAssetGroupResolver(svc, nil, nil, nil)

		_, err := resolver.ClusterAssetGroupsQuery(nil, nil, nil)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})
}

func TestClusterAssetGroupResolver_ClusterAssetGroupAssetsField(t *testing.T) {
	assetGroupName := "exampleAssetGroup"

	t.Run("Success", func(t *testing.T) {
		resources := []*v1beta1.ClusterAsset{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "ExampleClusterAssetA",
					Labels: map[string]string{
						rafter.AssetGroupLabel: assetGroupName,
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "ExampleClusterAssetB",
					Labels: map[string]string{
						rafter.AssetGroupLabel: assetGroupName,
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "ExampleClusterAssetC",
					Labels: map[string]string{
						rafter.AssetGroupLabel: assetGroupName,
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

		clusterAssetSvc := automock.NewClusterAssetService()
		clusterAssetSvc.On("ListForClusterAssetGroupByType", assetGroupName, []string{}).Return(resources, nil).Once()
		defer clusterAssetSvc.AssertExpectations(t)

		clusterAssetConverter := automock.NewGQLClusterAssetConverter()
		clusterAssetConverter.On("ToGQLs", resources).Return(expected, nil).Once()
		defer clusterAssetConverter.AssertExpectations(t)

		parentObj := gqlschema.ClusterAssetGroup{
			Name: assetGroupName,
		}

		resolver := rafter.NewClusterAssetGroupResolver(nil, nil, clusterAssetSvc, clusterAssetConverter)

		result, err := resolver.ClusterAssetGroupAssetsField(nil, &parentObj, []string{})

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		clusterAssetSvc := automock.NewClusterAssetService()
		clusterAssetSvc.On("ListForClusterAssetGroupByType", assetGroupName, []string{}).Return(nil, nil).Once()
		defer clusterAssetSvc.AssertExpectations(t)

		clusterAssetConverter := automock.NewGQLClusterAssetConverter()
		clusterAssetConverter.On("ToGQLs", ([]*v1beta1.ClusterAsset)(nil)).Return(nil, nil).Once()
		defer clusterAssetConverter.AssertExpectations(t)

		parentObj := gqlschema.ClusterAssetGroup{
			Name: assetGroupName,
		}

		resolver := rafter.NewClusterAssetGroupResolver(nil, nil, clusterAssetSvc, clusterAssetConverter)

		result, err := resolver.ClusterAssetGroupAssetsField(nil, &parentObj, []string{})

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		expectedErr := errors.New("Test")

		clusterAssetSvc := automock.NewClusterAssetService()
		clusterAssetSvc.On("ListForClusterAssetGroupByType", assetGroupName, []string{}).Return(nil, expectedErr).Once()
		defer clusterAssetSvc.AssertExpectations(t)

		parentObj := gqlschema.ClusterAssetGroup{
			Name: assetGroupName,
		}

		resolver := rafter.NewClusterAssetGroupResolver(nil, nil, clusterAssetSvc, nil)

		result, err := resolver.ClusterAssetGroupAssetsField(nil, &parentObj, []string{})

		assert.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

func TestClusterAssetGroupResolver_ClusterAssetGroupEventSubscription(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), (-24 * time.Hour))
		cancel()

		svc := automock.NewClusterAssetGroupService()
		svc.On("Subscribe", mock.Anything).Once()
		svc.On("Unsubscribe", mock.Anything).Once()
		resolver := rafter.NewClusterAssetGroupResolver(svc, nil, nil, nil)

		_, err := resolver.ClusterAssetGroupEventSubscription(ctx)

		require.NoError(t, err)
		svc.AssertCalled(t, "Subscribe", mock.Anything)
	})

	t.Run("Unsubscribe after connection close", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), (-24 * time.Hour))
		cancel()

		svc := automock.NewClusterAssetGroupService()
		svc.On("Subscribe", mock.Anything).Once()
		svc.On("Unsubscribe", mock.Anything).Once()
		resolver := rafter.NewClusterAssetGroupResolver(svc, nil, nil, nil)

		channel, err := resolver.ClusterAssetGroupEventSubscription(ctx)
		<-channel

		require.NoError(t, err)
		svc.AssertCalled(t, "Unsubscribe", mock.Anything)
	})
}

