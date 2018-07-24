package content_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content/automock"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content/storage"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApiSpecService_Find(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		getter := automock.NewMinioApiSpecGetter()
		getter.On("ApiSpec", "test/id").Return(fixApiSpec(), true, nil)
		defer getter.AssertExpectations(t)

		svc := content.NewApiSpecService(getter)

		result, err := svc.Find("test", "id")

		require.NoError(t, err)
		assert.Equal(t, fixApiSpec(), result)
	})

	t.Run("Not found", func(t *testing.T) {
		getter := automock.NewMinioApiSpecGetter()
		getter.On("ApiSpec", "test/id").Return(nil, false, nil)
		defer getter.AssertExpectations(t)

		svc := content.NewApiSpecService(getter)

		result, err := svc.Find("test", "id")

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		getter := automock.NewMinioApiSpecGetter()
		getter.On("ApiSpec", "test/id").Return(nil, false, errors.New("nope"))
		defer getter.AssertExpectations(t)

		svc := content.NewApiSpecService(getter)

		_, err := svc.Find("test", "id")

		require.Error(t, err)
	})
}

func fixApiSpec() *storage.ApiSpec {
	return &storage.ApiSpec{
		Raw: map[string]interface{}{
			"kind": "trololo",
			"name": "nope",
		},
	}
}
