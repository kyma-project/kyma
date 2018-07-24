package content_test

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content/automock"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content/storage"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
)

func TestContentResolver_ContentQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		cnt := &storage.Content{
			Raw: map[string]interface{}{
				"test": "data",
				"tree": map[string]interface{}{
					"testTree": "dataTree",
				},
			},
		}

		getter := automock.NewContentGetter()
		getter.On("Find", "test", "test").Return(cnt, nil)

		resolver := content.NewContentResolver(getter)

		result, err := resolver.ContentQuery(nil, "test", "test")

		require.NoError(t, err)
		assert.Equal(t, &gqlschema.JSON{
			"test": "data",
			"tree": map[string]interface{}{
				"testTree": "dataTree",
			},
		}, result)
	})

	t.Run("Not found", func(t *testing.T) {
		getter := automock.NewContentGetter()
		getter.On("Find", "test", "test").Return(nil, nil)

		resolver := content.NewContentResolver(getter)

		result, err := resolver.ContentQuery(nil, "test", "test")

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Error while gathering content", func(t *testing.T) {
		getter := automock.NewContentGetter()
		getter.On("Find", "test", "test").Return(nil, errors.New("trolololo"))

		resolver := content.NewContentResolver(getter)

		_, err := resolver.ContentQuery(nil, "test", "test")

		require.Error(t, err)
	})
}
