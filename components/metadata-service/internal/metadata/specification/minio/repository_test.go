package minio

import (
	"bytes"
	"testing"

	"github.com/kyma-project/kyma/components/metadata-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/metadata-service/internal/metadata/specification/minio/mocks"
	"github.com/minio/minio-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type errorReader struct {
}

func (errorReader) Read(p []byte) (n int, err error) {
	return 0, minio.ErrorResponse{}
}

func TestRepository_Create(t *testing.T) {
	t.Run("should create Minio resource", func(t *testing.T) {

		// given
		clientMock := &mocks.Client{}
		clientMock.On(
			"PutObjectWithContext",
			mock.AnythingOfType("*context.timerCtx"),
			"testBucket",
			"testRemotePath",
			mock.AnythingOfType("*bytes.Reader"),
			int64(0),
			minio.PutObjectOptions{}).
			Return(int64(0), nil)

		minioRepository := repository{clientMock, 5, "us-east-1"}

		// when
		err := minioRepository.Put("testBucket", "testRemotePath", make([]byte, 0))

		// then
		assert.Nil(t, err)

		clientMock.AssertExpectations(t)
	})

	t.Run("should return an error if creation failed", func(t *testing.T) {

		// given
		clientMock := &mocks.Client{}
		clientMock.On(
			"PutObjectWithContext",
			mock.AnythingOfType("*context.timerCtx"),
			"testBucket",
			"testRemotePath",
			mock.AnythingOfType("*bytes.Reader"),
			int64(0),
			minio.PutObjectOptions{}).
			Return(int64(0), apperrors.Internal("Error: %s", "error"))

		minioRepository := repository{clientMock, 5, "us-east-1"}

		// when
		err := minioRepository.Put("testBucket", "testRemotePath", make([]byte, 0))

		// then
		assert.NotNil(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())

		clientMock.AssertExpectations(t)
	})
}

func TestRepository_Remove(t *testing.T) {
	t.Run("should delete", func(t *testing.T) {
		// given
		clientMock := &mocks.Client{}
		clientMock.On("RemoveObject", "content", "service-class/1111-2222/content.json").Return(nil)

		minioRepository := repository{clientMock, 5, "us-east-1"}

		// when
		err := minioRepository.Remove("content", "service-class/1111-2222/content.json")

		// then
		assert.NoError(t, err)
	})

	t.Run("should handle errors when deleting", func(t *testing.T) {
		// given
		clientMock := &mocks.Client{}
		clientMock.On("RemoveObject", "content", "service-class/1111-2222/content.json").Return(apperrors.Internal("e"))

		minioRepository := repository{clientMock, 5, "us-east-1"}

		// when
		err := minioRepository.Remove("content", "service-class/1111-2222/content.json")

		// then
		assert.Error(t, err)
	})
}

func TestRepository_getObject(t *testing.T) {
	t.Run("should get Minio resource", func(t *testing.T) {

		// given
		minioObject := &minio.Object{}

		clientMock := &mocks.Client{}
		clientMock.On(
			"GetObjectWithContext",
			mock.AnythingOfType("*context.timerCtx"),
			"testBucket",
			"testRemotePath",
			minio.GetObjectOptions{}).
			Return(minioObject, nil)

		minioRepository := repository{clientMock, 5, "us-east-1"}

		// when
		object, err, cancel := minioRepository.getObject("testBucket", "testRemotePath")

		// then
		assert.Nil(t, err)
		assert.NotNil(t, object)

		assert.NotNil(t, cancel)

		clientMock.AssertExpectations(t)
	})

	t.Run("should return an error if minio fails", func(t *testing.T) {

		// given
		clientMock := &mocks.Client{}
		clientMock.On(
			"GetObjectWithContext",
			mock.AnythingOfType("*context.timerCtx"),
			"testBucket",
			"testRemotePath",
			minio.GetObjectOptions{}).
			Return(nil, apperrors.Internal("Error: %s", "an error"))

		minioRepository := repository{clientMock, 5, "us-east-1"}

		// when
		object, err, cancel := minioRepository.getObject("testBucket", "testRemotePath")

		// then
		assert.NotNil(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())

		assert.NotNil(t, cancel)

		assert.Nil(t, object)

		clientMock.AssertExpectations(t)
	})
}

func TestRepository_readBytes(t *testing.T) {
	t.Run("should return byte array from given io.Reader", func(t *testing.T) {

		// given
		data := []byte("testData")
		reader := bytes.NewReader(data)

		// when
		dataRead, err := readBytes(reader)

		// then
		assert.Equal(t, data, dataRead)
		assert.Nil(t, err)
	})

	t.Run("should return error in case reading fails", func(t *testing.T) {

		// given
		reader := errorReader{}

		// when
		_, err := readBytes(reader)

		// then
		assert.NotNil(t, err)
	})
}
