package uploader

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/kyma-project/kyma/components/asset-upload-service/internal/fileheader"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

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
	FileName   string `json:"fileName"`
	RemotePath string `json:"remotePath"`
	Bucket     string `json:"bucket"`
	Size       int64  `json:"size"`
}

type UploadError struct {
	FileName string
	Error    error
}

// Uploader is an abstraction layer for Minio client
type Uploader struct {
	client               MinioClient
	externalUploadOrigin string
	UploadTimeout        time.Duration
	MaxUploadWorkers     int
}

var (
	uploadFilesDurationHistogram = promauto.NewHistogram(prometheus.HistogramOpts{
		Name: "assetstore_upload_service_upload_files_duration_seconds",
		Help: "All files upload duration distribution",
	})
	uploadSingleFileDurationHistogram = promauto.NewHistogram(prometheus.HistogramOpts{
		Name: "assetstore_upload_service_upload_single_file_duration_seconds",
		Help: "Single file upload duration distribution",
	})
)

// New returns a new instance of Uploader
func New(client MinioClient, uploadOrigin string, uploadTimeout time.Duration, maxUploadWorkers int) *Uploader {
	return &Uploader{
		client:               client,
		UploadTimeout:        uploadTimeout,
		MaxUploadWorkers:     maxUploadWorkers,
		externalUploadOrigin: uploadOrigin,
	}
}

// UploadFiles uploads multiple files (Files struct) to particular bucket
func (u *Uploader) UploadFiles(ctx context.Context, filesChannel chan FileUpload, filesCount int) ([]UploadResult, []UploadError) {
	start := time.Now()

	errorsCh := make(chan *UploadError, filesCount)
	resultsCh := make(chan *UploadResult, filesCount)

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
					res, err := u.uploadFile(contextWithTimeout, upload)
					if err != nil {
						errorsCh <- &UploadError{
							Error:    err,
							FileName: upload.File.Filename(),
						}
					}

					if res != nil {
						resultsCh <- res
					}
				default:
				}
			}
		}()
	}

	waitGroup.Wait()
	close(resultsCh)
	close(errorsCh)

	result := u.populateResults(resultsCh)
	errs := u.populateErrors(errorsCh)

	uploadFilesDurationHistogram.Observe(time.Since(start).Seconds())

	return result, errs
}

func (u *Uploader) countNeededWorkers(filesCount, maxUploadWorkers int) int {
	if filesCount < maxUploadWorkers {
		return filesCount
	}
	return maxUploadWorkers
}

// UploadFile uploads single file from given path to particular bucket
func (u *Uploader) uploadFile(ctx context.Context, fileUpload FileUpload) (*UploadResult, error) {
	start := time.Now()

	file := fileUpload.File
	f, err := fileUpload.File.Open()
	if err != nil {
		return nil, errors.Wrapf(err, "while opening file %s", file.Filename())
	}
	defer func() {
		err := f.Close()
		if err != nil {
			glog.Error(errors.Wrapf(err, "while closing file %s", file.Filename()))
		}
	}()

	fileName := file.Filename()
	fileSize := file.Size()
	objectName := fmt.Sprintf("%s/%s", fileUpload.Directory, fileName)

	glog.Infof("Uploading `%s` into bucket `%s`...\n", objectName, fileUpload.Bucket)

	_, err = u.client.PutObjectWithContext(ctx, fileUpload.Bucket, objectName, f, fileSize, minio.PutObjectOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "Error while uploading file `%s` into `%s`", objectName, fileUpload.Bucket)
	}

	glog.Infof("Upload succeeded for `%s`.\n", objectName)

	result := &UploadResult{
		FileName:   fileName,
		Size:       fileSize,
		Bucket:     fileUpload.Bucket,
		RemotePath: fmt.Sprintf("%s/%s/%s", u.externalUploadOrigin, fileUpload.Bucket, objectName),
	}

	uploadSingleFileDurationHistogram.Observe(time.Since(start).Seconds())
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
func (u *Uploader) populateErrors(errorsCh chan *UploadError) []UploadError {
	var errs []UploadError
	for uploadErr := range errorsCh {
		if uploadErr == nil {
			continue
		}

		errs = append(errs, *uploadErr)
	}

	return errs
}

func Origin(uploadEndpoint string, secure bool) string {
	uploadProtocol := fmt.Sprint("http")
	if secure {
		uploadProtocol = uploadProtocol + "s"
	}

	return fmt.Sprintf("%s://%s", uploadProtocol, uploadEndpoint)
}
