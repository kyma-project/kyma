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

type FileToUpload struct {
	Name string
	Size int64
	FileHeader *multipart.FileHeader
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
func (u *Uploader) UploadFiles(ctx context.Context, files []FileToUpload, bucketName string) error {
	filesCount := len(files)
	uploadErrorsChannel := make(chan error, filesCount)
	filesChannel := u.makeClosedChannelWithFiles(files, filesCount)

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
				case file, ok := <-filesChannel:
					if !ok {
						return
					}
					uploadErrorsChannel <- u.uploadFile(contextWithTimeout, file, bucketName)
				default:
				}
			}
		}()
	}

	waitGroup.Wait()
	close(uploadErrorsChannel)
	return ConsumeUploadErrors(uploadErrorsChannel)
}

func (u *Uploader) makeClosedChannelWithFiles(files []filereader.File, filesCount int) chan filereader.File {
	filesChannel := make(chan filereader.File, filesCount)
	for _, file := range files {
		filesChannel <- file
	}
	close(filesChannel)
	return filesChannel
}

func (u *Uploader) countNeededWorkers(filesCount, maxUploadWorkers int) int {
	if filesCount < maxUploadWorkers {
		return filesCount
	}
	return maxUploadWorkers
}

// UploadFile uploads single file from given path to particular bucket
func (u *Uploader) uploadFile(ctx context.Context, file FileToUpload, bucketName string) error {
	f, err := file.FileHeader.Open()
	if err != nil {
		return errors.Wrapf(err,"while opening file %s", file.Name)
	}
	defer f.Close()

	glog.Infof("Uploading `%s`...\n", file.Name)

	_, err = u.client.PutObjectWithContext(ctx, bucketName, file.Name, f, file.Size, minio.PutObjectOptions{})
	if err != nil {
		return errors.Wrapf(err, "Error while uploading file `%s` into `%s`", file.name, bucketName)
	}

	return nil
}

// ConsumeUploadErrors consolidates all error messages into one and returns it
func ConsumeUploadErrors(channel chan error) error {
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
