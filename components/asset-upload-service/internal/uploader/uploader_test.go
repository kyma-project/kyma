package uploader_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/kyma-project/kyma/components/asset-upload-service/internal/uploader"
	"github.com/kyma-project/kyma/components/asset-upload-service/internal/uploader/automock"
	"github.com/onsi/gomega"
	"mime/multipart"
	"testing"
	"time"

	"github.com/minio/minio-go"
	"github.com/stretchr/testify/assert"
)

func TestUploader_UploadFiles(t *testing.T) {

	t.Run("Success", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		timeout, err := time.ParseDuration("10h")
		g.Expect(err).NotTo(gomega.HaveOccurred())
		filesCh, filesCount := testUploads()

		clientMock := new(automock.MinioClient)
		clientMock.On("FPutObjectWithContext", nil, "testBucket", "path/first", "local/path/first", minio.PutObjectOptions{}).Return(1, nil).Once()
		clientMock.On("FPutObjectWithContext", nil, "testBucket", "path/second", "local/path/second", minio.PutObjectOptions{}).Return(1, nil).Once()
		defer clientMock.AssertExpectations(t)

		uploadClient := uploader.New(clientMock, timeout, 5)

		// When
		err = uploadClient.UploadFiles(context.TODO(), filesCh, filesCount)

		// Then
		g.Expect(err).NotTo(gomega.HaveOccurred())
	})

	t.Run("Error", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		bucketName := "testBucket"
		timeout, err := time.ParseDuration("10h")
		g.Expect(err).NotTo(gomega.HaveOccurred())
		filesCh, filesCount := testUploads()

		clientMock := new(automock.MinioClient)
		uploadError := errors.New("Test upload error")
		clientMock.On("FPutObjectWithContext", nil, "testBucket", "path/first", "local/path/first", minio.PutObjectOptions{}).Return(-1, uploadError).Once()
		clientMock.On("FPutObjectWithContext", nil, "testBucket", "path/second", "local/path/second", minio.PutObjectOptions{}).Return(-1, uploadError).Once()

		uploadClient := uploader.New(clientMock, timeout, 5)


		// When
		err = uploadClient.UploadFiles(context.TODO(), filesCh, filesCount)

		// Then
		for _, file := range files {
			assert.Contains(t, actualError.Error(), fmt.Sprintf("while uploading file `%s` into `%s`: %s", file.LocalPath, bucketName, uploadError))
		}
		clientMock.AssertExpectations(t)
	})
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

func testUploads() (chan uploader.FileUpload, int) {
	filesCount := 5

	files := []uploader.FileUpload{
		{
			Bucket: "test",
			File: *multipart.FileHeader{
				Filename: "test.yaml",
				C
			}
		}
	}

	filesChannel := make(chan uploader.FileUpload, filesCount)
	for _, file := range files {
		filesChannel <- file
	}
	close(filesChannel)

	return filesChannel, filesCount
}
