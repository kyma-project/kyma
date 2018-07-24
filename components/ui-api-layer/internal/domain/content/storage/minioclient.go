package storage

import (
	"io"
	"net/url"
	"path/filepath"

	"github.com/golang/glog"
	"github.com/minio/minio-go"
	"github.com/pkg/errors"
)

type minioClient struct {
	client Minio
}

func newMinioClient(client Minio) *minioClient {
	return &minioClient{
		client: client,
	}
}

func (mc *minioClient) Object(bucketName, objectName string) (io.Reader, error) {
	return mc.client.GetObject(bucketName, objectName, minio.GetObjectOptions{})
}

func (mc *minioClient) IsNotExistsError(err error) bool {
	switch err := err.(type) {
	case minio.ErrorResponse:
		return err.Code == "NoSuchKey"
	default:
		return false
	}
}

func (mc *minioClient) NotificationChannel(bucketName string, stop <-chan struct{}) <-chan notification {
	notificationChannel := make(chan notification)

	channel := mc.client.ListenBucketNotification(bucketName, "", "", []string{
		"s3:ObjectCreated:*",
		"s3:ObjectRemoved:*",
	}, stop)

	go func() {
		defer close(notificationChannel)
		for info := range channel {
			if info.Err != nil {
				glog.Error(errors.Wrapf(info.Err, "while listening notifications on `%s` bucket", bucketName))
			}

			for _, record := range info.Records {
				key, err := url.QueryUnescape(record.S3.Object.Key)
				if err != nil {
					glog.Warningf("Cannot parse object key: `%s`", record.S3.Object.Key)
					continue
				}

				parent := filepath.Dir(key)
				if parent == "." {
					parent = ""
				}

				notification := notification{
					parent:    parent,
					filename:  filepath.Base(key),
					eventType: record.EventName,
				}

				notificationChannel <- notification
			}
		}
	}()

	return notificationChannel
}
