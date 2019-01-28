package cleaner

import (
	"context"
	"github.com/minio/minio-go"
)

//go:generate mockery -name=MinioClient -output=automock -outpkg=automock -case=underscore
type MinioClient interface {
	ListObjects(bucketName, objectPrefix string, recursive bool, doneCh <-chan struct{}) <-chan minio.ObjectInfo
	RemoveObjectsWithContext(ctx context.Context, bucketName string, objectsCh <-chan string) <-chan minio.RemoveObjectError
}

//go:generate mockery -name=Cleaner -output=automock -outpkg=automock -case=underscore
type Cleaner interface {
	Clean(ctx context.Context, bucket, objectPrefix string) error
}

type cleaner struct {
	minioClient MinioClient
}

func New(minioClient MinioClient) Cleaner {
	return &cleaner{
		minioClient: minioClient,
	}
}

func (c *cleaner) Clean(ctx context.Context, bucket, objectPrefix string) error {
	keys, err := c.listObjectsKeys(bucket, objectPrefix)
	if err != nil || len(keys) == 0 {
		return err
	}
	objectsCh := make(chan string)

	go func() {
		defer close(objectsCh)

		for _, key := range keys {
			objectsCh <- key
		}
	}()

	for rErr := range c.minioClient.RemoveObjectsWithContext(ctx, bucket, objectsCh) {
		err = rErr.Err
	}

	return err
}

func (c *cleaner) listObjectsKeys(bucket, objectPrefix string) ([]string, error) {
	var result []string

	doneCh := make(chan struct{})
	defer close(doneCh)

	for message := range c.minioClient.ListObjects(bucket, objectPrefix, true, doneCh) {
		if message.Err != nil {
			return nil, message.Err
		}
		result = append(result, message.Key)
	}

	return result, nil
}
