package uploader

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/minio/minio-go"
)

type uploader struct {
	minioClient MinioClient
}

//go:generate mockery -name=MinioClient -output=automock -outpkg=automock -case=underscore
type MinioClient interface {
	FPutObjectWithContext(ctx context.Context, bucketName, objectName, filePath string, opts minio.PutObjectOptions) (n int64, err error)
	ListObjects(bucketName, objectPrefix string, recursive bool, doneCh <-chan struct{}) <-chan minio.ObjectInfo
}

//go:generate mockery -name=Uploader -output=automock -outpkg=automock -case=underscore
type Uploader interface {
	Upload(ctx context.Context, bucketName, assetName, basePath string, files []string) error
	ContainsAll(bucketName, assetName string, files []string) (bool, error)
}

func New(minioClient MinioClient) Uploader {
	return &uploader{
		minioClient: minioClient,
	}
}

func (u *uploader) Upload(ctx context.Context, bucketName, assetName, basePath string, files []string) error {
	for _, f := range files {
		bucketPath := filepath.Join(assetName, f)
		path := filepath.Join(basePath, f)

		_, err := u.minioClient.FPutObjectWithContext(ctx, bucketName, bucketPath, path, minio.PutObjectOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func (u *uploader) ContainsAll(bucketName, assetName string, files []string) (bool, error) {
	objects := u.listObjects(bucketName, fmt.Sprintf("/%s", assetName))
	for _, f := range files {
		key := fmt.Sprintf("/%s/%s", assetName, f)

		info, ok := objects[key]
		if !ok {
			return false, nil
		}

		if info.Err != nil {
			return false, info.Err
		}
	}

	return true, nil
}

func (c *uploader) listObjects(bucket, objectPrefix string) map[string]minio.ObjectInfo {
	result := make(map[string]minio.ObjectInfo)

	doneCh := make(chan struct{})
	defer close(doneCh)

	for message := range c.minioClient.ListObjects(bucket, objectPrefix, true, doneCh) {
		result[message.Key] = message
	}

	return result
}
