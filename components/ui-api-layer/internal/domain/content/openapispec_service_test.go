package content_test

import (
	"github.com/stretchr/testify/require"
	"testing"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content/automock"
	"github.com/stretchr/testify/assert"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content/storage"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content"
	"github.com/pkg/errors"
)

func TestOpenApiSpecService_Find(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		getter := automock.NewMinioOpenApiSpecGetter()
		getter.On("OpenApiSpec", "test/id").Return(fixOpenApiSpec(), true, nil)
		defer getter.AssertExpectations(t)

		svc := content.NewOpenApiSpecService(getter)

		result, err := svc.Find("test", "id")

		require.NoError(t, err)
		assert.Equal(t, fixOpenApiSpec(), result)
	})

	t.Run("Not found", func(t *testing.T) {
		getter := automock.NewMinioOpenApiSpecGetter()
		getter.On("OpenApiSpec", "test/id").Return(nil, false, nil)
		defer getter.AssertExpectations(t)

		svc := content.NewOpenApiSpecService(getter)

		result, err := svc.Find("test", "id")

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		getter := automock.NewMinioOpenApiSpecGetter()
		getter.On("OpenApiSpec", "test/id").Return(nil, false, errors.New("nope"))
		defer getter.AssertExpectations(t)

		svc := content.NewOpenApiSpecService(getter)

		_, err := svc.Find("test", "id")

		require.Error(t, err)
	})
}

func fixOpenApiSpec() *storage.OpenApiSpec {
	return &storage.OpenApiSpec{
		Raw: map[string]interface{}{
			"kind": "trololo",
			"name": "nope",
		},
	}
}
