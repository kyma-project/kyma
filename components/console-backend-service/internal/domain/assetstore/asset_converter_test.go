package assetstore

import (
	"testing"

	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAssetConverter_ToGQL(t *testing.T) {
	t.Run("All properties are given", func(t *testing.T) {
		converter := assetConverter{}

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
		}

		result, err := converter.ToGQL(item)
		require.NoError(t, err)
		assert.Equal(t, &expected, result)
	})

	t.Run("Empty", func(t *testing.T) {
		converter := &assetConverter{}
		_, err := converter.ToGQL(&v1alpha2.Asset{})
		require.NoError(t, err)
	})

	t.Run("Nil", func(t *testing.T) {
		converter := &assetConverter{}
		item, err := converter.ToGQL(nil)

		require.NoError(t, err)
		assert.Nil(t, item)
	})
}

func TestAssetConverter_ToGQLs(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		assets := []*v1alpha2.Asset{
			fixAsset(),
			fixAsset(),
		}

		converter := assetConverter{}
		result, err := converter.ToGQLs(assets)

		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "ExampleName", result[0].Name)
	})

	t.Run("Empty", func(t *testing.T) {
		var assets []*v1alpha2.Asset

		converter := assetConverter{}
		result, err := converter.ToGQLs(assets)

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("With nil", func(t *testing.T) {
		assets := []*v1alpha2.Asset{
			nil,
			fixAsset(),
			nil,
		}

		converter := assetConverter{}
		result, err := converter.ToGQLs(assets)

		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "ExampleName", result[0].Name)
	})
}

func fixAsset() *v1alpha2.Asset {
	return &v1alpha2.Asset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ExampleName",
			Namespace: "ExampleNamespace",
			Labels: map[string]string{
				CmsTypeLabel: "ExampleType",
			},
		},
		Spec: v1alpha2.AssetSpec{
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
		Status: v1alpha2.AssetStatus{
			CommonAssetStatus: v1alpha2.CommonAssetStatus{
				Phase:   v1alpha2.AssetReady,
				Reason:  "ExampleReason",
				Message: "ExampleMessage",
			},
		},
	}
}
