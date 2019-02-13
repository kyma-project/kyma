package uploader_test

import (
"context"
"errors"
"fmt"
	"github.com/kyma-project/kyma/components/asset-upload-service/internal/uploader"
	"github.com/kyma-project/kyma/components/asset-upload-service/internal/uploader/automock"
	"github.com/stretchr/testify/require"
	"mime/multipart"
	"testing"
"time"

"github.com/minio/minio-go"
"github.com/stretchr/testify/assert"

)

func TestUploadFilesSuccess(t *testing.T) {
	timeout, err := time.ParseDuration("10h")
	assert.Nil(t, err)
	files := getTestFiles()

	clientMock := new(automock.MinioClient)
	clientMock.On("FPutObjectWithContext", nil, "testBucket", "path/first", "local/path/first", minio.PutObjectOptions{}).Return(1, nil).Once()
	clientMock.On("FPutObjectWithContext", nil, "testBucket", "path/second", "local/path/second", minio.PutObjectOptions{}).Return(1, nil).Once()
	defer clientMock.AssertExpectations(t)

	uploadClient := uploader.New(clientMock, timeout, 5)
	err := uploadClient.UploadFiles(context.TODO(), files, bucketName)

	require.NoError(t, err)
}

func TestUploadFilesFailure(t *testing.T) {
	bucketName := "testBucket"
	timeout, err := time.ParseDuration("10h")
	assert.Nil(t, err)
	files := getTestFiles()
	clientMock := new(automock.MinioClient)
	uploadError := errors.New("Test upload error")
	clientMock.On("FPutObjectWithContext", nil, "testBucket", "path/first", "local/path/first", minio.PutObjectOptions{}).Return(-1, uploadError).Once()
	clientMock.On("FPutObjectWithContext", nil, "testBucket", "path/second", "local/path/second", minio.PutObjectOptions{}).Return(-1, uploadError).Once()
	uploadClient := uploader.New(clientMock, timeout, 5)

	actualError := uploadClient.UploadFiles(context.TODO(), files, bucketName)

	for _, file := range files {
		assert.Contains(t, actualError.Error(), fmt.Sprintf("while uploading file `%s` into `%s`: %s", file.LocalPath, bucketName, uploadError))
	}
	clientMock.AssertExpectations(t)
}

func TestConsumeUploadErrors(t *testing.T) {
	t.Run("Returning no error in empty channel", func(t *testing.T) {
		noErrorChan := make(chan error, 2)
		close(noErrorChan)

		actualError := uploader.ConsumeUploadErrors(noErrorChan)
		assert.Nil(t, actualError)
	})

	t.Run("Consolidating errors from channel", func(t *testing.T) {
		errorChan := make(chan error, 2)
		errorChan <- errors.New("Test 1")
		errorChan <- errors.New("Test 2")
		close(errorChan)

		actualError := uploader.ConsumeUploadErrors(errorChan)
		assert.EqualError(t, actualError, "Test 1;\nTest 2")
	})
}

func getTestFiles() []uploader.FileUpload {
	multipart.File()
	return []uploader.FileUpload{
		{
			LocalPath:  "local/path/first",
			RemotePath: "path/first",
		},
		{
			LocalPath:  "local/path/second",
			RemotePath: "path/second",
		},
	}
}
