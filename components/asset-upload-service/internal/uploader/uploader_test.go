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
				Bucket:    "test",
				File:      mock1,
				Directory: "testDir",
			},
			{
				Bucket:    "test2",
				File:      mock2,
				Directory: "testDir",
			},
		}

		expectedResult := []uploader.UploadResult{
			{
				FileName: "test1.yaml",
				RemotePath: "https://minio.example.com/test/testDir/test1.yaml",
				Bucket: "test",
				Size: -1,
			},
			{
				FileName: "test2.yaml",
				RemotePath: "https://minio.example.com/test2/testDir/test2.yaml",
				Bucket: "test2",
				Size: -1,
			},
		}

		timeout, err := time.ParseDuration("10h")
		g.Expect(err).NotTo(gomega.HaveOccurred())
		filesCh, filesCount := testUploads(files)

		ctxArgFn := func(ctx context.Context) bool { return true }

		clientMock := new(automock.MinioClient)
		clientMock.On("PutObjectWithContext", mock.MatchedBy(ctxArgFn), "test", "testDir/test1.yaml", file, int64(-1), minio.PutObjectOptions{}).Return(int64(1), nil).Once()
		clientMock.On("PutObjectWithContext", mock.MatchedBy(ctxArgFn), "test2", "testDir/test2.yaml", file, int64(-1), minio.PutObjectOptions{}).Return(int64(1), nil).Once()
		defer clientMock.AssertExpectations(t)

		uploadClient := uploader.New(clientMock, "https://minio.example.com", timeout, 5)

		// When
		res, err := uploadClient.UploadFiles(context.TODO(), filesCh, filesCount)

		// Then
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(res).To(gomega.Equal(expectedResult))
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
				Bucket:    bucketName,
				File:      mock1,
				Directory: "testDir",
			},
			{
				Bucket:    bucketName,
				File:      mock2,
				Directory: "testDir",
			},
		}

		timeout, err := time.ParseDuration("10h")
		g.Expect(err).NotTo(gomega.HaveOccurred())
		filesCh, filesCount := testUploads(files)

		ctxArgFn := func(ctx context.Context) bool { return true }

		clientMock := new(automock.MinioClient)
		clientMock.On("PutObjectWithContext", mock.MatchedBy(ctxArgFn), bucketName, "testDir/test1.yaml", file, int64(-1), minio.PutObjectOptions{}).Return(int64(1), testErr).Once()
		clientMock.On("PutObjectWithContext", mock.MatchedBy(ctxArgFn), bucketName, "testDir/test2.yaml", file, int64(-1), minio.PutObjectOptions{}).Return(int64(1), testErr).Once()
		defer clientMock.AssertExpectations(t)

		uploadClient := uploader.New(clientMock, "https://minio.example.com", timeout, 5)

		// When
		_, err = uploadClient.UploadFiles(context.TODO(), filesCh, filesCount)

		// Then
		g.Expect(err).To(gomega.HaveOccurred())

		for _, file := range files {
			g.Expect(err.Error()).To(gomega.ContainSubstring(fmt.Sprintf("while uploading file `%s` into `%s`: %s", file.File.Filename(), bucketName, testErr)))
		}
		clientMock.AssertExpectations(t)

	})
}

func TestUploader_PopulateErrors(t *testing.T) {
	t.Run("Errors", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		errCh := make(chan error, 2)
		errCh <- errors.New("Test 1")
		errCh <- errors.New("Test 2")
		close(errCh)

		u := uploader.Uploader{}

		// When
		err := u.PopulateErrors(errCh)

		// Then
		g.Expect(err).To(gomega.HaveOccurred())
		g.Expect(err.Error()).To(gomega.Equal("Test 1;\nTest 2"))
	})

	t.Run("No Errors", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		errCh := make(chan error)
		close(errCh)

		u := uploader.Uploader{}

		// When
		err := u.PopulateErrors(errCh)

		// Then
		g.Expect(err).NotTo(gomega.HaveOccurred())
	})
}

func TestUploader_PopulateResults(t *testing.T) {
	t.Run("Results", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		res1 := uploader.UploadResult{
			FileName: "test.yaml",
		}
		res2 := uploader.UploadResult{
			FileName: "test2.yaml",
		}

		resultsCh := make(chan *uploader.UploadResult, 3)
		resultsCh <- &res1
		resultsCh <- &res2
		resultsCh <- nil
		close(resultsCh)

		u := uploader.Uploader{}

		// When
		res := u.PopulateResults(resultsCh)

		// Then
		g.Expect(res).To(gomega.HaveLen(2))
		g.Expect(res).To(gomega.ContainElement(res1))
		g.Expect(res).To(gomega.ContainElement(res2))
	})

	t.Run("No Results", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		resultsCh := make(chan *uploader.UploadResult, 3)
		close(resultsCh)

		u := uploader.Uploader{}

		// When
		res := u.PopulateResults(resultsCh)

		// Then
		g.Expect(res).To(gomega.BeEmpty())
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
