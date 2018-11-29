package minio

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/kyma-project/kyma/components/metadata-service/internal/apperrors"
	"github.com/minio/minio-go"
)

const (
	timeoutDuration = time.Duration(5)
	secureTraffic   = false
	bucketLocation  = "us-east-1"
)

type Client interface {
	PutObjectWithContext(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (n int64, err error)
	GetObjectWithContext(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (*minio.Object, error)
	RemoveObject(bucketName, objectName string) error
	BucketExists(bucketName string) (found bool, err error)
	MakeBucket(bucketName, location string) error
}

type Repository interface {
	Put(bucketName string, objectName string, resource []byte) apperrors.AppError
	Get(bucketName string, objectName string) ([]byte, apperrors.AppError)
	Remove(bucketName string, objectName string) apperrors.AppError
}

type repository struct {
	minioClient    Client
	timeout        time.Duration
	bucketLocation string
}

func NewMinioRepository(endpoint string, accessKeyID string, secretAccessKey string) (Repository, apperrors.AppError) {
	minioClient, err := minio.New(endpoint, accessKeyID, secretAccessKey, secureTraffic)
	if err != nil {
		return nil, apperrors.Internal("Failed creating Minio client, %s", err)
	}

	return &repository{minioClient: minioClient, timeout: timeoutDuration, bucketLocation: bucketLocation}, nil
}

func (r *repository) Put(bucketName string, objectName string, resource []byte) apperrors.AppError {
	contextWithTimeout, cancel := context.WithTimeout(context.Background(), r.timeout*time.Second)
	defer cancel()

	reader := bytes.NewReader(resource)

	_, appErr := r.minioClient.PutObjectWithContext(contextWithTimeout, bucketName, objectName, reader, reader.Size(), minio.PutObjectOptions{})
	if appErr != nil {
		return apperrors.Internal("Uploading file to Minio failed, %s", appErr.Error())
	}

	return nil
}

func (r *repository) Get(bucketName string, objectName string) ([]byte, apperrors.AppError) {
	object, err, cancel := r.getObject(bucketName, objectName)
	defer cancel()
	if err != nil {
		return nil, err
	}

	data, err := readBytes(object)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (r *repository) Remove(bucketName string, objectName string) apperrors.AppError {
	err := r.minioClient.RemoveObject(bucketName, objectName)
	if err != nil {
		return apperrors.Internal("Removing %s object failed, %s", objectName, err.Error())
	}

	return nil
}

func (r *repository) getObject(bucketName string, objectName string) (*minio.Object, apperrors.AppError, func()) {
	contextWithTimeout, cancel := context.WithTimeout(context.Background(), r.timeout*time.Second)

	object, err := r.minioClient.GetObjectWithContext(contextWithTimeout, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, apperrors.Internal("Getting %s object failed, %s", objectName, err.Error()), cancel
	}

	return object, nil, cancel
}

func readBytes(reader io.Reader) ([]byte, apperrors.AppError) {
	buffer := bytes.Buffer{}

	_, err := buffer.ReadFrom(reader)
	if err != nil {
		errorResponse := err.(minio.ErrorResponse)
		if errorResponse.Code == "NoSuchKey" {
			return nil, nil
		}

		return nil, apperrors.Internal("Reading bytes failed, %s", err.Error())
	}

	return buffer.Bytes(), nil
}
