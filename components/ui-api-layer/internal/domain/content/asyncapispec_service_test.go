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

func TestAsyncApiSpecService_Find(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		getter := automock.NewMinioAsyncApiSpecGetter()
		getter.On("AsyncApiSpec", "test/id").Return(fixAsyncApiSpec(), true, nil)
		defer getter.AssertExpectations(t)

		svc := content.NewAsyncApiSpecService(getter)

		result, err := svc.Find("test", "id")

		require.NoError(t, err)
		assert.Equal(t, fixAsyncApiSpec(), result)
	})

	t.Run("Not found", func(t *testing.T) {
		getter := automock.NewMinioAsyncApiSpecGetter()
		getter.On("AsyncApiSpec", "test/id").Return(nil, false, nil)
		defer getter.AssertExpectations(t)

		svc := content.NewAsyncApiSpecService(getter)

		result, err := svc.Find("test", "id")

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		getter := automock.NewMinioAsyncApiSpecGetter()
		getter.On("AsyncApiSpec", "test/id").Return(nil, false, errors.New("nope"))
		defer getter.AssertExpectations(t)

		svc := content.NewAsyncApiSpecService(getter)

		_, err := svc.Find("test", "id")

		require.Error(t, err)
	})
}

func fixAsyncApiSpec() *storage.AsyncApiSpec {
	return &storage.AsyncApiSpec{
		Raw: map[string]interface{}{
			"asyncapi": "0.0.1",
		},
		Data: storage.AsyncApiSpecData{
			AsyncAPI: "0.0.1",
		},
	}
}
