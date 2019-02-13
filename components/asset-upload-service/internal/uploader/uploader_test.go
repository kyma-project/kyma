package uploader_test

import (
	"context"
	"errors"
	"fmt"
	fautomock "github.com/kyma-project/kyma/components/asset-upload-service/internal/fileheader/automock"
	"github.com/kyma-project/kyma/components/asset-upload-service/internal/uploader"
	"github.com/kyma-project/kyma/components/asset-upload-service/internal/uploader/automock"
	"github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"

	"github.com/minio/minio-go"
	"github.com/stretchr/testify/assert"
)

func TestUploader_UploadFiles(t *testing.T) {

	t.Run("Success", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		file := &fautomock.File{}
		file.On("Close").Return(nil)

		mock1 := &fautomock.FileHeader{}
		mock1.On("Filename").Return("test1.yaml")
		mock1.On("Size", ).Return(int64(-1)).Once()
		mock1.On("Open").Return(file, nil).Once()

		mock2 := &fautomock.FileHeader{}
		mock2.On("Filename").Return("test2.yaml")
		mock2.On("Size", ).Return(int64(-1)).Once()
		mock2.On("Open").Return(file, nil).Once()

		files := []uploader.FileUpload{
			{
				Bucket: "test",
				File:   mock1,
			},
			{
				Bucket: "test2",
				File:   mock2,
			},
		}

		timeout, err := time.ParseDuration("10h")
		g.Expect(err).NotTo(gomega.HaveOccurred())
		filesCh, filesCount := testUploads(files)

		ctxArgFn := func(ctx context.Context) bool { return true }

		clientMock := new(automock.MinioClient)
		clientMock.On("PutObjectWithContext", mock.MatchedBy(ctxArgFn), "test", "test1.yaml", file, int64(-1), minio.PutObjectOptions{}).Return(int64(1), nil).Once()
		clientMock.On("PutObjectWithContext", mock.MatchedBy(ctxArgFn), "test2", "test2.yaml", file, int64(-1), minio.PutObjectOptions{}).Return(int64(1), nil).Once()
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

		file := &fautomock.File{}
		file.On("Close").Return(nil)

		mock1 := &fautomock.FileHeader{}
		mock1.On("Filename").Return("test1.yaml")
		mock1.On("Size", ).Return(int64(-1)).Once()
		mock1.On("Open").Return(file, nil).Once()

		mock2 := &fautomock.FileHeader{}
		mock2.On("Filename").Return("test2.yaml")
		mock2.On("Size", ).Return(int64(-1)).Once()
		mock2.On("Open").Return(file, nil).Once()

		testErr := errors.New("Test error")
		bucketName := "test"
		files := []uploader.FileUpload{
			{
				Bucket: bucketName,
				File:   mock1,
			},
			{
				Bucket: bucketName,
				File:   mock2,
			},
		}

		timeout, err := time.ParseDuration("10h")
		g.Expect(err).NotTo(gomega.HaveOccurred())
		filesCh, filesCount := testUploads(files)

		ctxArgFn := func(ctx context.Context) bool { return true }

		clientMock := new(automock.MinioClient)
		clientMock.On("PutObjectWithContext", mock.MatchedBy(ctxArgFn), bucketName, "test1.yaml", file, int64(-1), minio.PutObjectOptions{}).Return(int64(1), testErr).Once()
		clientMock.On("PutObjectWithContext", mock.MatchedBy(ctxArgFn), bucketName, "test2.yaml", file, int64(-1), minio.PutObjectOptions{}).Return(int64(1), testErr).Once()
		defer clientMock.AssertExpectations(t)

		uploadClient := uploader.New(clientMock, timeout, 5)

		// When
		err = uploadClient.UploadFiles(context.TODO(), filesCh, filesCount)

		// Then
		g.Expect(err).To(gomega.HaveOccurred())

		for _, file := range files {
			g.Expect(err.Error()).To(gomega.ContainSubstring(fmt.Sprintf("while uploading file `%s` into `%s`: %s", file.File.Filename(), bucketName, testErr)))
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

func testUploads(files []uploader.FileUpload) (chan uploader.FileUpload, int) {
	filesCount := len(files)

	filesChannel := make(chan uploader.FileUpload, filesCount)
	for _, file := range files {
		filesChannel <- file
	}
	close(filesChannel)

	return filesChannel, filesCount
}
