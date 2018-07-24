package content

import (
	"testing"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content/storage"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/stretchr/testify/assert"
)

func TestContentConverter_ToGQL(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		content := &storage.Content{
			Raw: map[string]interface{}{
				"test": "data",
				"tree": map[string]interface{}{
					"treeTest": "treeData",
				},
			},
		}

		converter := &contentConverter{}

		result := converter.ToGQL(content)
		assert.Equal(t, &gqlschema.JSON{
			"test": "data",
			"tree": map[string]interface{}{
				"treeTest": "treeData",
			},
		}, result)
	})

	t.Run("Empty", func(t *testing.T) {
		converter := &contentConverter{}
		converter.ToGQL(&storage.Content{})
	})

	t.Run("Nil", func(t *testing.T) {
		converter := &contentConverter{}

		result := converter.ToGQL(nil)
		assert.Nil(t, result)
	})
}
