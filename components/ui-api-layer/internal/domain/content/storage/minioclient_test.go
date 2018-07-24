package storage_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content/storage"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content/storage/automock"
	minio2 "github.com/minio/minio-go"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestMinioClient_IsNotExistsError(t *testing.T) {
	minio := new(automock.Minio)
	client := storage.NewMinioClient(minio)

	t.Run("Other error", func(t *testing.T) {
		ok := client.IsNotExistsError(errors.New("other error"))
		assert.False(t, ok)
	})

	t.Run("Nil error", func(t *testing.T) {
		ok := client.IsNotExistsError(nil)
		assert.False(t, ok)
	})

	t.Run("Not exists error", func(t *testing.T) {
		ok := client.IsNotExistsError(minio2.ErrorResponse{Code: "NoSuchKey"})
		assert.True(t, ok)
	})

	t.Run("Different code", func(t *testing.T) {
		ok := client.IsNotExistsError(minio2.ErrorResponse{Code: "Different Code"})
		assert.False(t, ok)
	})
}
func TestMinioClient_Object(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		minio := new(automock.Minio)
		client := storage.NewMinioClient(minio)

		minio.On("GetObject", "valid", "name", mock.Anything).
			Return(&minio2.Object{}, nil)
		obj, err := client.Object("valid", "name")
		require.NoError(t, err)
		assert.IsType(t, &minio2.Object{}, obj)
	})

	t.Run("Error while getting object", func(t *testing.T) {
		minio := new(automock.Minio)
		client := storage.NewMinioClient(minio)

		minio.On("GetObject", "invalid", "name", mock.Anything).
			Return(nil, errors.New("get-object"))
		_, err := client.Object("invalid", "name")
		require.Error(t, err)
	})
}
