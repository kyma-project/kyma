package cleaner_test

import (
	"context"
	"errors"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/cleaner"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/cleaner/automock"
	"github.com/minio/minio-go"
	"github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestCleaner_Clean(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		minioClient := &automock.MinioClient{}
		cleaner := cleaner.New(minioClient)

		bucket := "ns-test-bucket"
		prefix := "test-asset/"
		ctx := context.Background()
		objectsInfo := []minio.ObjectInfo{
			{
				Key: "test-asset/test-one",
			},
			{
				Key: "test-asset/sub/test-two",
			},
		}

		minioClient.On("ListObjects", bucket, prefix, true, mock.Anything).Return(objectInfoChan(objectsInfo)).Once()
		minioClient.On("RemoveObjectsWithContext", ctx, bucket, mock.Anything).Return(objectErrorsChan(nil)).Once()
		defer minioClient.AssertExpectations(t)

		// When
		err := cleaner.Clean(ctx, bucket, prefix)

		// Then
		g.Expect(err).NotTo(gomega.HaveOccurred())
	})

	t.Run("Empty", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		minioClient := &automock.MinioClient{}
		cleaner := cleaner.New(minioClient)

		bucket := "ns-test-bucket"
		prefix := "test-asset/"
		ctx := context.Background()

		minioClient.On("ListObjects", bucket, prefix, true, mock.Anything).Return(objectInfoChan(nil)).Once()
		defer minioClient.AssertExpectations(t)

		// When
		err := cleaner.Clean(ctx, bucket, prefix)

		// Then
		g.Expect(err).NotTo(gomega.HaveOccurred())
	})

	t.Run("CollectingError", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		minioClient := &automock.MinioClient{}
		cleaner := cleaner.New(minioClient)

		bucket := "ns-test-bucket"
		prefix := "test-asset/"
		ctx := context.Background()
		objectsInfo := []minio.ObjectInfo{
			{
				Key: "test-asset/test-one",
				Err: errors.New("test error"),
			},
			{
				Key: "test-asset/sub/test-two",
			},
		}

		minioClient.On("ListObjects", bucket, prefix, true, mock.Anything).Return(objectInfoChan(objectsInfo)).Once()
		defer minioClient.AssertExpectations(t)

		// When
		err := cleaner.Clean(ctx, bucket, prefix)

		// Then
		g.Expect(err).To(gomega.HaveOccurred())
	})

	t.Run("RemovingError", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		minioClient := &automock.MinioClient{}
		cleaner := cleaner.New(minioClient)

		bucket := "ns-test-bucket"
		prefix := "test-asset/"
		ctx := context.Background()
		objectsInfo := []minio.ObjectInfo{
			{
				Key: "test-asset/test-one",
			},
			{
				Key: "test-asset/sub/test-two",
			},
		}
		errors := []minio.RemoveObjectError{
			{
				Err: errors.New("oczko"),
			},
		}

		minioClient.On("ListObjects", bucket, prefix, true, mock.Anything).Return(objectInfoChan(objectsInfo)).Once()
		minioClient.On("RemoveObjectsWithContext", ctx, bucket, mock.Anything).Return(objectErrorsChan(errors)).Once()
		defer minioClient.AssertExpectations(t)

		// When
		err := cleaner.Clean(ctx, bucket, prefix)

		// Then
		g.Expect(err).To(gomega.HaveOccurred())
	})
}

func objectInfoChan(objects []minio.ObjectInfo) <-chan minio.ObjectInfo {
	objectsCh := make(chan minio.ObjectInfo)

	go func() {
		defer close(objectsCh)

		for _, obj := range objects {
			objectsCh <- obj
		}
	}()

	return objectsCh
}

func objectErrorsChan(errors []minio.RemoveObjectError) <-chan minio.RemoveObjectError {
	errorsCh := make(chan minio.RemoveObjectError)

	go func() {
		defer close(errorsCh)

		for _, obj := range errors {
			errorsCh <- obj
		}
	}()

	return errorsCh
}
