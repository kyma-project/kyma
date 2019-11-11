package rafter_test

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter"
	"testing"

	"encoding/json"

	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestFileService_ToGQL(t *testing.T) {
	rawMap := fixRawMap(t)

	t.Run("Success without filter", func(t *testing.T) {
		assetRef := fixAssetStatusRef(rawMap)
		expected := []*rafter.File{
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

		svc := rafter.NewFileService()

		result, err := svc.Extract(assetRef)
		require.NoError(t, err)

		assert.Equal(t, expected, result)
	})

	t.Run("Success with filter", func(t *testing.T) {
		assetRef := fixAssetStatusRef(rawMap)
		expected := []*rafter.File{
			{
				URL:      "https://example.com/markdown.md",
				Metadata: rawMap,
			},
		}

		svc := rafter.NewFileService()

		result, err := svc.FilterByExtensionsAndExtract(assetRef, []string{"md"})
		require.NoError(t, err)

		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		svc := rafter.NewFileService()

		result, err := svc.Extract(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func fixAssetStatusRef(rawMap *runtime.RawExtension) *v1beta1.AssetStatusRef {
	return &v1beta1.AssetStatusRef{
		BaseURL: "https://example.com",
		Files: []v1beta1.AssetFile{
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
	}
}

func fixRawMap(t *testing.T) *runtime.RawExtension {
	rawMap := map[string]interface{}{
		"labels": []string{"test1", "test2"},
	}
	raw, err := json.Marshal(rawMap)
	require.NoError(t, err)

	return &runtime.RawExtension{
		Raw: raw,
	}
}
