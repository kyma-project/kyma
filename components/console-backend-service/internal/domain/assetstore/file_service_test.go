package assetstore_test

import (
	"testing"

	"encoding/json"

	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/assetstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestFileConverter_ToGQL(t *testing.T) {
	t.Run("Success without filter", func(t *testing.T) {
		rawMap := fixRawMap(t)

		assetRef := &v1alpha2.AssetStatusRef{
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
		}
		expected := []*assetstore.File{
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

		svc := assetstore.NewFileService()

		result, err := svc.Extract(assetRef)
		require.NoError(t, err)

		assert.Equal(t, expected, result)
	})

	t.Run("Success with filter", func(t *testing.T) {
		rawMap := fixRawMap(t)

		assetRef := &v1alpha2.AssetStatusRef{
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
		}
		expected := []*assetstore.File{
			{
				URL:      "https://example.com/markdown.md",
				Metadata: rawMap,
			},
		}

		svc := assetstore.NewFileService()

		result, err := svc.FilterByExtensionsAndExtract(assetRef, []string{"md"})
		require.NoError(t, err)

		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		svc := assetstore.NewFileService()

		result, err := svc.Extract(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})
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
