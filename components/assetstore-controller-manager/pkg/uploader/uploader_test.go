package uploader_test

import (
	"context"
	"fmt"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/uploader"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/uploader/automock"
	"github.com/minio/minio-go"
	"github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"net/http"
	"path/filepath"
	"testing"
)

func TestUploader_ContainsAll(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		minioClient := new(automock.MinioClient)
		uploader := uploader.New(minioClient)

		bucket := "ns-test-bucket"
		asset := "test-asset"
		files := []string{"ala.md", "ma.zip", "kota.json"}
		objectsInfo := make([]minio.ObjectInfo, 0, len(files))
		for _, file := range files {
			info := minio.ObjectInfo{
				Key: fmt.Sprintf("/%s/%s", asset, file),
			}

			objectsInfo = append(objectsInfo, info)
		}

		minioClient.On("ListObjects", bucket, fmt.Sprintf("/%s", asset), true, mock.Anything).Return(objectInfoChan(objectsInfo)).Once()
		defer minioClient.AssertExpectations(t)

		// When
		contains, err := uploader.ContainsAll(bucket, asset, files)

		// Then
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(contains).To(gomega.Equal(true))
	})

	t.Run("Missing", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		minioClient := new(automock.MinioClient)
		uploader := uploader.New(minioClient)

		bucket := "ns-test-bucket"
		asset := "test-asset"
		files := []string{"ala.md", "ma.zip", "kota.json"}
		objectsInfo := make([]minio.ObjectInfo, 0, len(files))

		minioClient.On("ListObjects", bucket, fmt.Sprintf("/%s", asset), true, mock.Anything).Return(objectInfoChan(objectsInfo)).Once()
		defer minioClient.AssertExpectations(t)

		// When
		contains, err := uploader.ContainsAll(bucket, asset, files)

		// Then
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(contains).To(gomega.Equal(false))
	})

	t.Run("Error", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		minioClient := new(automock.MinioClient)
		uploader := uploader.New(minioClient)

		bucket := "ns-test-bucket"
		asset := "test-asset"
		files := []string{"ala.md", "ma.zip", "kota.json"}
		objectsInfo := make([]minio.ObjectInfo, 0, len(files))
		for _, file := range files {
			info := minio.ObjectInfo{
				Key: fmt.Sprintf("/%s/%s", asset, file),
				Err: minio.ErrorResponse{StatusCode: http.StatusBadGateway},
			}

			objectsInfo = append(objectsInfo, info)
		}

		minioClient.On("ListObjects", bucket, fmt.Sprintf("/%s", asset), true, mock.Anything).Return(objectInfoChan(objectsInfo)).Once()
		defer minioClient.AssertExpectations(t)

		// When
		contains, err := uploader.ContainsAll(bucket, asset, files)

		// Then
		g.Expect(err).To(gomega.HaveOccurred())
		g.Expect(contains).To(gomega.Equal(false))
	})
}

func TestUploader_Upload(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		minioClient := new(automock.MinioClient)
		uploader := uploader.New(minioClient)

		bucket := "ns-test-bucket"
		asset := "test-asset"

		files := []string{"ala.md", "ma.zip", "kota.json"}
		basePath := "/tmp/test-asset"
		ctx := context.Background()

		for _, path := range files {
			objectName := filepath.Join(asset, path)
			filePath := filepath.Join(basePath, path)
			minioClient.On("FPutObjectWithContext", ctx, bucket, objectName, filePath, mock.Anything).Return(int64(0), nil).Once()
		}
		defer minioClient.AssertExpectations(t)

		// When
		err := uploader.Upload(ctx, bucket, asset, basePath, files)

		// Then
		g.Expect(err).NotTo(gomega.HaveOccurred())
	})

	t.Run("Fail", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		minioClient := new(automock.MinioClient)
		uploader := uploader.New(minioClient)

		bucket := "ns-test-bucket"
		asset := "test-asset"

		files := []string{"ala.md"}
		basePath := "/tmp/test-asset"
		ctx := context.Background()

		objectName := filepath.Join(asset, files[0])
		filePath := filepath.Join(basePath, files[0])
		minioClient.On("FPutObjectWithContext", ctx, bucket, objectName, filePath, mock.Anything).Return(int64(0), fmt.Errorf("oczko")).Once()
		defer minioClient.AssertExpectations(t)

		// When
		err := uploader.Upload(ctx, bucket, asset, basePath, files)

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
