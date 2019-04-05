package assetstore

import (
	"testing"

	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestClusterAssetConverter_ToGQL(t *testing.T) {
	t.Run("All properties are given", func(t *testing.T) {
		converter := clusterAssetConverter{}

		item := fixClusterAsset()
		expected := gqlschema.ClusterAsset{
			Name: "ExampleName",
			Type: "ExampleType",
			Status: gqlschema.AssetStatus{
				Phase:   gqlschema.AssetPhaseTypeReady,
				Reason:  "ExampleReason",
				Message: "ExampleMessage",
			},
		}

		result, err := converter.ToGQL(item)
		require.NoError(t, err)
		assert.Equal(t, &expected, result)
	})

	t.Run("Empty", func(t *testing.T) {
		converter := &clusterAssetConverter{}
		_, err := converter.ToGQL(&v1alpha2.ClusterAsset{})
		require.NoError(t, err)
	})

	t.Run("Nil", func(t *testing.T) {
		converter := &clusterAssetConverter{}
		item, err := converter.ToGQL(nil)

		require.NoError(t, err)
		assert.Nil(t, item)
	})
}

func TestClusterAssetConverter_ToGQLs(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		clusterAssets := []*v1alpha2.ClusterAsset{
			fixClusterAsset(),
			fixClusterAsset(),
		}

		converter := clusterAssetConverter{}
		result, err := converter.ToGQLs(clusterAssets)

		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "ExampleName", result[0].Name)
	})

	t.Run("Empty", func(t *testing.T) {
		var clusterAssets []*v1alpha2.ClusterAsset

		converter := clusterAssetConverter{}
		result, err := converter.ToGQLs(clusterAssets)

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("With nil", func(t *testing.T) {
		clusterAssets := []*v1alpha2.ClusterAsset{
			nil,
			fixClusterAsset(),
			nil,
		}

		converter := clusterAssetConverter{}
		result, err := converter.ToGQLs(clusterAssets)

		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "ExampleName", result[0].Name)
	})
}

func fixClusterAsset() *v1alpha2.ClusterAsset {
	return &v1alpha2.ClusterAsset{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ExampleName",
			Labels: map[string]string{
				CmsTypeLabel: "ExampleType",
			},
		},
		Spec: v1alpha2.ClusterAssetSpec{
			CommonAssetSpec: v1alpha2.CommonAssetSpec{
				Source: v1alpha2.AssetSource{
					Mode: v1alpha2.AssetSingle,
					URL:  "ExampleUrl",
				},
				BucketRef: v1alpha2.AssetBucketRef{
					Name: "ExampleBucketRef",
				},
			},
		},
		Status: v1alpha2.ClusterAssetStatus{
			CommonAssetStatus: v1alpha2.CommonAssetStatus{
				Phase:   v1alpha2.AssetReady,
				Reason:  "ExampleReason",
				Message: "ExampleMessage",
			},
		},
	}
}
