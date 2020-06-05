package rafter_test

import (
	"testing"

	"encoding/json"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestFileConverter_ToGQL(t *testing.T) {
	t.Run("All properties are given", func(t *testing.T) {
		converter := rafter.NewFileConverter()

		item := fixFile(t)
		expected := gqlschema.File{
			URL: "ExampleUrl",
			Metadata: gqlschema.JSON{
				"labels": []interface{}{"test1", "test2"},
			},
		}

		result, err := converter.ToGQL(item)
		require.NoError(t, err)
		assert.Equal(t, &expected, result)
	})

	t.Run("Empty", func(t *testing.T) {
		converter := rafter.NewFileConverter()
		_, err := converter.ToGQL(&rafter.File{})
		require.NoError(t, err)
	})

	t.Run("Nil", func(t *testing.T) {
		converter := rafter.NewFileConverter()
		item, err := converter.ToGQL(nil)

		require.NoError(t, err)
		assert.Nil(t, item)
	})
}

func TestFileConverter_ToGQLs(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		files := []*rafter.File{
			fixFile(t),
			fixFile(t),
		}

		converter := rafter.NewFileConverter()
		result, err := converter.ToGQLs(files)

		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "ExampleUrl", result[0].URL)
	})

	t.Run("Empty", func(t *testing.T) {
		var files []*rafter.File

		converter := rafter.NewFileConverter()
		result, err := converter.ToGQLs(files)

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("With nil", func(t *testing.T) {
		files := []*rafter.File{
			nil,
			fixFile(t),
			nil,
		}

		converter := rafter.NewFileConverter()
		result, err := converter.ToGQLs(files)

		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "ExampleUrl", result[0].URL)
	})
}

func fixFile(t *testing.T) *rafter.File {
	rawMap := map[string]interface{}{
		"labels": []string{"test1", "test2"},
	}
	raw, err := json.Marshal(rawMap)
	require.NoError(t, err)

	return &rafter.File{
		URL: "ExampleUrl",
		Metadata: &runtime.RawExtension{
			Raw: raw,
		},
	}
}
