package cleaner

import (
	"context"
	"fmt"
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

type CleanError struct {
	message string
	errors  []error
}

func (e *CleanError) Error() string {
	return e.message
}

func (e *CleanError) Errors() []error {
	return e.errors
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
	keys, errors := c.listObjectsKeys(bucket, objectPrefix)
	if len(errors) > 0 {
		messages := c.extractErrorMessages(errors)
		return fmt.Errorf("cannot list objects in bucket: %+v", messages)
	}
	if len(keys) == 0 {
		return nil
	}

	objectsCh := make(chan string)

	go func() {
		defer close(objectsCh)

		for _, key := range keys {
			objectsCh <- key
		}
	}()

	for err := range c.minioClient.RemoveObjectsWithContext(ctx, bucket, objectsCh) {
		errors = append(errors, err.Err)
	}

	if len(errors) > 0 {
		messages := c.extractErrorMessages(errors)
		return fmt.Errorf("cannot delete objects from bucke: %+v", messages)
	}

	return nil
}

func (c *cleaner) extractErrorMessages(errors []error) []string {
	messages := make([]string, 0, len(errors))
	for _, err := range errors {
		messages = append(messages, err.Error())
	}

	return messages
}

func (c *cleaner) listObjectsKeys(bucket, objectPrefix string) ([]string, []error) {
	var result []string

	doneCh := make(chan struct{})
	defer close(doneCh)

	var errors []error
	for message := range c.minioClient.ListObjects(bucket, objectPrefix, true, doneCh) {
		if message.Err != nil {
			errors = append(errors, message.Err)
		}

		result = append(result, message.Key)
	}

	return result, errors
}
