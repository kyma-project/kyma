package content_test

import (
	"testing"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content/automock"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/assert"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content/storage"
	"github.com/pkg/errors"
)

func TestODataSpecService_Find(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		getter := automock.NewMinioODataSpecGetter()
		getter.On("ODataSpec", "test/id").Return(fixODataSpec(), true, nil)
		defer getter.AssertExpectations(t)

		svc := content.NewODataSpecService(getter)

		result, err := svc.Find("test", "id")

		require.NoError(t, err)
		assert.Equal(t, fixODataSpec(), result)
	})

	t.Run("Not found", func(t *testing.T) {
		getter := automock.NewMinioODataSpecGetter()
		getter.On("ODataSpec", "test/id").Return(nil, false, nil)
		defer getter.AssertExpectations(t)

		svc := content.NewODataSpecService(getter)

		result, err := svc.Find("test", "id")

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		getter := automock.NewMinioODataSpecGetter()
		getter.On("ODataSpec", "test/id").Return(nil, false, errors.New("nope"))
		defer getter.AssertExpectations(t)

		svc := content.NewODataSpecService(getter)

		_, err := svc.Find("test", "id")

		require.Error(t, err)
	})
}

func fixODataSpec() *storage.ODataSpec {
	return &storage.ODataSpec{
		Raw: "example",
	}
}
