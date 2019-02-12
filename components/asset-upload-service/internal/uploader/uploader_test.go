package uploader_test

import (
"context"
"errors"
"fmt"
	"github.com/kyma-project/kyma/components/asset-upload-service/internal/uploader"
	"testing"
"time"

"github.com/minio/minio-go"
"github.com/stretchr/testify/assert"

)

func TestCreateBucketIfDoesntExistSuccess(t *testing.T) {
	bucketRegion := "sampleRegion"
	objectPrefix := "*"

	t.Run("Succeeding when bucket already exists", func(t *testing.T) {
		existingBucketName := "existingBucket"
		timeout, err := time.ParseDuration("10h")
		assert.Nil(t, err)
		clientMock := new(MockClient)
		clientMock.On("BucketExists", existingBucketName).Return(true, nil).Once()
		uploadClient := uploader.New(clientMock, timeout, 5)

		uploadClient.CreateBucketIfDoesntExist(existingBucketName, bucketRegion)

		clientMock.AssertExpectations(t)
	})

	t.Run("Creating a new bucket when it doesn't exist", func(t *testing.T) {
		notExistingBucketName := "notExistingBucket"
		timeout, err := time.ParseDuration("10h")
		assert.Nil(t, err)
		clientMock := new(MockClient)
		clientMock.On("BucketExists", notExistingBucketName).Return(false, nil).Once()
		clientMock.On("MakeBucket", notExistingBucketName, bucketRegion).Return(nil).Once()
		clientMock.On("SetBucketPolicy", notExistingBucketName, objectPrefix, bucketPolicy).Return(nil).Once()

		uploadClient := uploader.New(clientMock, timeout, 5)

		uploadClient.CreateBucketIfDoesntExist(notExistingBucketName, bucketRegion)

		clientMock.AssertExpectations(t)
	})

}

func TestCreateBucketIfDoesntExistFailure(t *testing.T) {
	bucketRegion := "region"
	objectPrefix := "*"

	t.Run("Failing on checking if bucket exists", func(t *testing.T) {
		bucketName := "checkingExistingFailure"
		bucketExistError := errors.New("Bucket Exist Failure")
		timeout, err := time.ParseDuration("10h")
		assert.Nil(t, err)
		clientMock := new(MockClient)
		clientMock.On("BucketExists", bucketName).Return(false, bucketExistError).Once()
		uploadClient := uploader.New(clientMock, timeout, 5)

		actualError := uploadClient.CreateBucketIfDoesntExist(bucketName, bucketRegion)

		assert.EqualError(t, actualError, fmt.Sprintf("while checking if bucket `%s` exists: %s", bucketName, bucketExistError))
		clientMock.AssertExpectations(t)
	})

	t.Run("Failing on bucket creation", func(t *testing.T) {
		bucketName := "makingFailure"
		timeout, err := time.ParseDuration("10h")
		assert.Nil(t, err)
		clientMock := new(MockClient)
		bucketMakingError := errors.New("Bucket Making Failure")
		clientMock.On("BucketExists", bucketName).Return(false, nil).Once()
		clientMock.On("MakeBucket", bucketName, bucketRegion).Return(bucketMakingError).Once()
		uploadClient := uploader.New(clientMock, timeout, 5)

		actualError := uploadClient.CreateBucketIfDoesntExist(bucketName, bucketRegion)

		assert.EqualError(t, actualError, fmt.Sprintf("while creating bucket `%s` in region `%s`: %s", bucketName, bucketRegion, bucketMakingError))
		clientMock.AssertExpectations(t)
	})

	t.Run("Failing on bucket policy configuration", func(t *testing.T) {
		bucketName := "configureFailure"
		timeout, err := time.ParseDuration("10h")
		assert.Nil(t, err)
		clientMock := new(MockClient)
		bucketConfigurationError := errors.New("Bucket Configuration Failure")

		clientMock.On("BucketExists", bucketName).Return(false, nil).Once()
		clientMock.On("MakeBucket", bucketName, bucketRegion).Return(nil).Once()
		clientMock.On("SetBucketPolicy", bucketName, objectPrefix, bucketPolicy).Return(bucketConfigurationError).Once()

		uploadClient := uploader.New(clientMock, timeout, 5)

		actualError := uploadClient.CreateBucketIfDoesntExist(bucketName, bucketRegion)

		assert.EqualError(t, actualError, fmt.Sprintf("while setting bucket policy on bucket `%s`: %s", bucketName, bucketConfigurationError))
		clientMock.AssertExpectations(t)
	})
}

func TestUploadFilesSuccess(t *testing.T) {
	bucketName := "testBucket"
	timeout, err := time.ParseDuration("10h")
	assert.Nil(t, err)
	files := getTestFiles()
	clientMock := new(MockClient)
	clientMock.On("FPutObjectWithContext", nil, "testBucket", "path/first", "local/path/first", minio.PutObjectOptions{}).Return(1, nil).Once()
	clientMock.On("FPutObjectWithContext", nil, "testBucket", "path/second", "local/path/second", minio.PutObjectOptions{}).Return(1, nil).Once()
	uploadClient := uploader.New(clientMock, timeout, 5)
	uploadClient.UploadFiles(context.TODO(), files, bucketName)

	clientMock.AssertExpectations(t)
}

func TestUploadFilesFailure(t *testing.T) {
	bucketName := "testBucket"
	timeout, err := time.ParseDuration("10h")
	assert.Nil(t, err)
	files := getTestFiles()
	clientMock := new(MockClient)
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

func getTestFiles() []filereader.File {
	return []filereader.File{
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
