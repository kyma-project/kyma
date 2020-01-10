package rafter_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAssetConverter_ToGQL(t *testing.T) {
	converter := rafter.NewAssetConverter()

	t.Run("All properties are given", func(t *testing.T) {
		item := fixAsset()
		expected := gqlschema.Asset{
			Name:      "ExampleName",
			Namespace: "ExampleNamespace",
			Type:      "ExampleType",
			Status: gqlschema.AssetStatus{
				Phase:   gqlschema.AssetPhaseTypeReady,
				Reason:  "ExampleReason",
				Message: "ExampleMessage",
			},
			Parameters: gqlschema.JSON{"complex": map[string]interface{}{"data": "true"}, "json": "true"},
		}

		result, err := converter.ToGQL(item)
		require.NoError(t, err)
		assert.Equal(t, &expected, result)
	})

	t.Run("Empty", func(t *testing.T) {
		_, err := converter.ToGQL(&v1beta1.Asset{})
		require.NoError(t, err)
	})

	t.Run("Nil", func(t *testing.T) {
		item, err := converter.ToGQL(nil)

		require.NoError(t, err)
		assert.Nil(t, item)
	})
}

func TestAssetConverter_ToGQLs(t *testing.T) {
	converter := rafter.NewAssetConverter()

	t.Run("Success", func(t *testing.T) {
		assets := []*v1beta1.Asset{
			fixAsset(),
			fixAsset(),
		}

		result, err := converter.ToGQLs(assets)

		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "ExampleName", result[0].Name)
	})

	t.Run("Empty", func(t *testing.T) {
		var assets []*v1beta1.Asset

		result, err := converter.ToGQLs(assets)

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("With nil", func(t *testing.T) {
		assets := []*v1beta1.Asset{
			nil,
			fixAsset(),
			nil,
		}

		result, err := converter.ToGQLs(assets)

		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "ExampleName", result[0].Name)
	})
}

func fixAsset() *v1beta1.Asset {
	return &v1beta1.Asset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ExampleName",
			Namespace: "ExampleNamespace",
			Labels: map[string]string{
				rafter.TypeLabel: "ExampleType",
			},
		},
		Spec: v1beta1.AssetSpec{
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
		Status: v1beta1.AssetStatus{
			CommonAssetStatus: v1beta1.CommonAssetStatus{
				Phase:   v1beta1.AssetReady,
				Reason:  "ExampleReason",
				Message: "ExampleMessage",
			},
		},
	}
}
