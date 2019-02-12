package uploader

import (
	"context"
	"io"
	"mime/multipart"
	"strings"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/minio/minio-go"
	"github.com/pkg/errors"
)

//go:generate mockery -name=MinioClient -output=automock -outpkg=automock -case=underscore
type MinioClient interface {
	PutObjectWithContext(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64,
		opts minio.PutObjectOptions) (n int64, err error)
}

type FileUpload struct {
	Bucket string
	File   *multipart.FileHeader
}

// Uploader is an abstraction layer for Minio client
type Uploader struct {
	client           MinioClient
	UploadTimeout    time.Duration
	MaxUploadWorkers int
}

// New returns a new instance of Uploader
func New(client MinioClient, uploadTimeout time.Duration, maxUploadWorkers int) *Uploader {
	return &Uploader{
		client:           client,
		UploadTimeout:    uploadTimeout,
		MaxUploadWorkers: maxUploadWorkers,
	}
}

// UploadFiles uploads multiple files (Files struct) to particular bucket
func (u *Uploader) UploadFiles(ctx context.Context, filesChannel chan FileUpload, filesCount int) error {
	uploadErrorsChannel := make(chan error, filesCount)
	contextWithTimeout, cancel := context.WithTimeout(ctx, u.UploadTimeout)
	defer cancel()

	workersCount := u.countNeededWorkers(filesCount, u.MaxUploadWorkers)
	glog.Infof("Creating %d concurrent upload worker(s)...", workersCount)
	var waitGroup sync.WaitGroup
	waitGroup.Add(workersCount)
	for i := 0; i < workersCount; i++ {
		go func() {
			defer waitGroup.Done()
			for {
				select {
				case <-contextWithTimeout.Done():
					glog.Error(errors.Wrapf(contextWithTimeout.Err(), "Error while concurrently uploading file"))
					return
				default:
				}

				select {
				case upload, ok := <-filesChannel:
					if !ok {
						return
					}
					uploadErrorsChannel <- u.uploadFile(contextWithTimeout, upload)
				default:
				}
			}
		}()
	}


	waitGroup.Wait()
	close(uploadErrorsChannel)
	return consumeUploadErrors(uploadErrorsChannel)
}

func (u *Uploader) countNeededWorkers(filesCount, maxUploadWorkers int) int {
	if filesCount < maxUploadWorkers {
		return filesCount
	}
	return maxUploadWorkers
}

// UploadFile uploads single file from given path to particular bucket
func (u *Uploader) uploadFile(ctx context.Context, fileUpload FileUpload) error {
	file := fileUpload.File
	f, err := fileUpload.File.Open()
	if err != nil {
		return errors.Wrapf(err, "while opening file %s", file.Filename)
	}
	defer f.Close()

	glog.Infof("Uploading `%s`...\n", file.Filename)

	_, err = u.client.PutObjectWithContext(ctx, fileUpload.Bucket, file.Filename, f, file.Size, minio.PutObjectOptions{})
	if err != nil {
		return errors.Wrapf(err, "Error while uploading file `%s` into `%s`", file.Filename, fileUpload.Bucket)
	}

	glog.Infof("Upload succeeded for `%s`.\n", file.Filename)

	return nil
}

// consumeUploadErrors consolidates all error messages into one and returns it
func consumeUploadErrors(channel chan error) error {
	var messages []string
	for err := range channel {
		if err != nil {
			messages = append(messages, err.Error())
		}
	}

	if len(messages) > 0 {
		message := strings.Join(messages, ";\n")
		return errors.New(message)
	}

	return nil
}
