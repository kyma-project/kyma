package rafter_test

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestClusterAssetConverter_ToGQL(t *testing.T) {
	converter := rafter.NewClusterAssetConverter()

	t.Run("All properties are given", func(t *testing.T) {
		item := fixClusterAsset()
		expected := gqlschema.ClusterAsset{
			Name: "ExampleName",
			Type: "ExampleType",
			Status: gqlschema.AssetStatus{
				Phase:   gqlschema.AssetPhaseTypeReady,
				Reason:  "ExampleReason",
				Message: "ExampleMessage",
			},
			Metadata:   gqlschema.JSON{"complex": map[string]interface{}{"data": "true"}, "json": "true"},
			Parameters: gqlschema.JSON{"complex": map[string]interface{}{"data": "true"}, "json": "true"},
		}

		result, err := converter.ToGQL(item)
		require.NoError(t, err)
		assert.Equal(t, &expected, result)
	})

	t.Run("Empty", func(t *testing.T) {
		_, err := converter.ToGQL(&v1beta1.ClusterAsset{})
		require.NoError(t, err)
	})

	t.Run("Nil", func(t *testing.T) {
		item, err := converter.ToGQL(nil)

		require.NoError(t, err)
		assert.Nil(t, item)
	})
}

func TestClusterAssetConverter_ToGQLs(t *testing.T) {
	converter := rafter.NewClusterAssetConverter()

	t.Run("Success", func(t *testing.T) {
		clusterAssets := []*v1beta1.ClusterAsset{
			fixClusterAsset(),
			fixClusterAsset(),
		}

		result, err := converter.ToGQLs(clusterAssets)

		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "ExampleName", result[0].Name)
	})

	t.Run("Empty", func(t *testing.T) {
		var clusterAssets []*v1beta1.ClusterAsset

		result, err := converter.ToGQLs(clusterAssets)

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("With nil", func(t *testing.T) {
		clusterAssets := []*v1beta1.ClusterAsset{
			nil,
			fixClusterAsset(),
			nil,
		}

		result, err := converter.ToGQLs(clusterAssets)

		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "ExampleName", result[0].Name)
	})
}

func fixClusterAsset() *v1beta1.ClusterAsset {
	return &v1beta1.ClusterAsset{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ExampleName",
			Labels: map[string]string{
				rafter.TypeLabel: "ExampleType",
			},
		},
		Spec: v1beta1.ClusterAssetSpec{
			CommonAssetSpec: v1beta1.CommonAssetSpec{
				Source: v1beta1.AssetSource{
					Mode: v1beta1.AssetSingle,
					URL:  "ExampleUrl",
				},
				BucketRef: v1beta1.AssetBucketRef{
					Name: "ExampleBucketRef",
				},
				Parameters: &runtime.RawExtension{Raw: []byte(`{"json":"true","complex":{"data":"true"}}`)},
			},
		},
		Status: v1beta1.ClusterAssetStatus{
			CommonAssetStatus: v1beta1.CommonAssetStatus{
				Phase:   v1beta1.AssetReady,
				Reason:  "ExampleReason",
				Message: "ExampleMessage",
			},
		},
	}
}
