package extractor

import (
	"context"
	"github.com/kyma-project/kyma/components/asset-metadata-service/pkg/fileheader"
	"github.com/kyma-project/kyma/components/asset-metadata-service/pkg/matador"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type Job struct {
	File fileheader.FileHeader
}

// ResultError stores success data
type ResultSuccess struct {
	Metadata map[string]interface{} `json:"metadata"`
	FileName string                 `json:"fileName"`
}

// ResultError stores error data
type ResultError struct {
	Error    error `json:"error"`
	FileName string `json:"omitempty,fileName"`
}

// Extractor is an abstraction layer for Minio client
type Extractor struct {
	ProcessTimeout time.Duration
	MaxWorkers     int

	matador matador.Matador
}

// New returns a new instance of Extractor
func New(maxWorkers int, processTimeout time.Duration) *Extractor {
	return &Extractor{
		ProcessTimeout: processTimeout,
		MaxWorkers:     maxWorkers,
		matador:        matador.New(),
	}
}

// Process processes files and extracts file metadata
func (e *Extractor) Process(ctx context.Context, filesChannel chan Job, filesCount int) ([]ResultSuccess, []ResultError) {
	errorsCh := make(chan *ResultError, filesCount)
	resultsCh := make(chan *ResultSuccess, filesCount)

	contextWithTimeout, cancel := context.WithTimeout(ctx, e.ProcessTimeout)
	defer cancel()

	workersCount := e.countNeededWorkers(filesCount, e.MaxWorkers)
	glog.Infof("Creating %d concurrent worker(s)...", workersCount)
	var waitGroup sync.WaitGroup
	waitGroup.Add(workersCount)
	for i := 0; i < workersCount; i++ {
		go func() {
			defer waitGroup.Done()
			for {
				select {
				case <-contextWithTimeout.Done():
					glog.Error(errors.Wrapf(contextWithTimeout.Err(), "ResultError while concurrently processing file"))
					return
				default:
				}

				select {
				case job, ok := <-filesChannel:
					if !ok {
						return
					}
					res, err := e.processFile(contextWithTimeout, job)
					if err != nil {
						errorsCh <- &ResultError{
							Error:    err,
							FileName: job.File.Filename(),
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

	result := e.populateResults(resultsCh)
	errs := e.populateErrors(errorsCh)
	return result, errs
}

func (e *Extractor) countNeededWorkers(filesCount, maxUploadWorkers int) int {
	if filesCount < maxUploadWorkers {
		return filesCount
	}
	return maxUploadWorkers
}

// processFile processes a single file from given path to particular bucket
func (e *Extractor) processFile(ctx context.Context, job Job) (*ResultSuccess, error) {
	file := job.File
	fileName := file.Filename()

	glog.Infof("Extracting metadata `%s`...\n", file.Filename())

	m, err := e.matador.ReadMetadata(file)
	if err != nil {
		return nil, errors.Wrapf(err, "while processing file `%s`", file.Filename())
	}

	result := &ResultSuccess{
		FileName: fileName,
		Metadata: m,
	}

	return result, nil
}

// populateResults populates all results from all jobs
func (e *Extractor) populateResults(resultsCh chan *ResultSuccess) []ResultSuccess {
	var result []ResultSuccess
	for i := range resultsCh {
		if i == nil {
			continue
		}

		result = append(result, *i)
	}

	return result
}

// populateErrors consolidates all error messages into one and returns it
func (e *Extractor) populateErrors(errorsCh chan *ResultError) []ResultError {
	var errs []ResultError
	for uploadErr := range errorsCh {
		if uploadErr == nil {
			continue
		}

		errs = append(errs, *uploadErr)
	}

	return errs
}
