package uploader

import (
	"context"
	"fmt"
	"github.com/kyma-project/kyma/components/asset-upload-service/internal/fileheader"
	"io"
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
	Bucket    string
	File      fileheader.FileHeader
	Directory string
}

type UploadResult struct {
	FileName   string
	RemotePath string
	Bucket     string
	Size       int64
}

// Uploader is an abstraction layer for Minio client
type Uploader struct {
	client           MinioClient
	storeEndpoint    string
	UploadTimeout    time.Duration
	MaxUploadWorkers int
}

// New returns a new instance of Uploader
func New(client MinioClient, storeEndpoint string, uploadTimeout time.Duration, maxUploadWorkers int) *Uploader {
	return &Uploader{
		client:           client,
		UploadTimeout:    uploadTimeout,
		MaxUploadWorkers: maxUploadWorkers,
		storeEndpoint:    storeEndpoint,
	}
}

// UploadFiles uploads multiple files (Files struct) to particular bucket
func (u *Uploader) UploadFiles(ctx context.Context, filesChannel chan FileUpload, filesCount int) ([]UploadResult, error) {
	errorsCh := make(chan error, filesCount)
	resultsCh := make(chan *UploadResult, filesCount)

	contextWithTimeout, cancel := context.WithTimeout(ctx, u.UploadTimeout)
	defer cancel()

	go func() {

	}()

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
					res, err := u.uploadFile(contextWithTimeout, upload)
					resultsCh <- res
					errorsCh <- err
				default:
				}
			}
		}()
	}

	waitGroup.Wait()
	close(resultsCh)
	close(errorsCh)

	result := u.populateResults(resultsCh)
	errors := u.populateErrors(errorsCh)
	return result, errors
}

func (u *Uploader) countNeededWorkers(filesCount, maxUploadWorkers int) int {
	if filesCount < maxUploadWorkers {
		return filesCount
	}
	return maxUploadWorkers
}

// UploadFile uploads single file from given path to particular bucket
func (u *Uploader) uploadFile(ctx context.Context, fileUpload FileUpload) (*UploadResult, error) {
	file := fileUpload.File
	f, err := fileUpload.File.Open()
	if err != nil {
		return nil, errors.Wrapf(err, "while opening file %s", file.Filename())
	}
	defer f.Close()

	glog.Infof("Uploading `%s`...\n", file.Filename())

	fileName := file.Filename()
	fileSize := file.Size()
	objectName := fmt.Sprintf("%s/%s", fileUpload.Directory, fileName)
	_, err = u.client.PutObjectWithContext(ctx, fileUpload.Bucket, objectName, f, fileSize, minio.PutObjectOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "Error while uploading file `%s` into `%s`", file.Filename(), fileUpload.Bucket)
	}

	glog.Infof("Upload succeeded for `%s`.\n", file.Filename())

	result := &UploadResult{
		FileName:   fileName,
		Size:       fileSize,
		Bucket:     fileUpload.Bucket,
		RemotePath: fmt.Sprintf("%s/%s/%s", u.storeEndpoint, fileUpload.Bucket, objectName),
	}

	return result, nil
}

func (u *Uploader) populateResults(resultsCh chan *UploadResult) []UploadResult {
	var result []UploadResult
	for i := range resultsCh {
		if i == nil {
			continue
		}

		result = append(result, *i)
	}

	return result
}

// consumeUploadErrors consolidates all error messages into one and returns it
func (u *Uploader) populateErrors(errorsCh chan error) error {
	var messages []string
	for err := range errorsCh {
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
