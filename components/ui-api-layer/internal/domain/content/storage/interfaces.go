package storage

import (
	"io"

	"github.com/minio/minio-go"
)

//go:generate mockery -name=Cache -output=automock -outpkg=automock -case=underscore
type Cache interface {
	Delete(key string) error
	Get(key string) ([]byte, error)
	Set(key string, entry []byte) error
	Reset() error
}

//go:generate mockery -name=Minio -output=automock -outpkg=automock -case=underscore
type Minio interface {
	GetObject(bucketName, objectName string, opts minio.GetObjectOptions) (*minio.Object, error)
	ListenBucketNotification(bucketName, prefix, suffix string, events []string, doneCh <-chan struct{}) <-chan minio.NotificationInfo
}

//go:generate mockery -name=client -inpkg -case=underscore
type client interface {
	Object(bucketName, objectName string) (io.Reader, error)
	NotificationChannel(bucketName string, stop <-chan struct{}) <-chan notification
	IsNotExistsError(err error) bool
}
