package processor

import (
	"context"
	"github.com/kyma-project/kyma/components/asset-metadata-service/pkg/fileheader"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type Job struct {
	FilePath string
	File     fileheader.FileHeader
}

// ResultError stores success data
type ResultSuccess struct {
	FilePath string
	Output   interface{}
}

// ResultError stores error data
type ResultError struct {
	FilePath string
	Error    error
}

// Processor processes jobs concurrently
type Processor struct {
	ProcessTimeout time.Duration
	MaxWorkers     int

	workerFn func(job Job) (interface{}, error)
}

// New returns a new instance of Processor
func New(workerFn func(job Job) (interface{}, error), maxWorkers int, processTimeout time.Duration) *Processor {
	return &Processor{
		ProcessTimeout: processTimeout,
		MaxWorkers:     maxWorkers,
		workerFn:       workerFn,
	}
}

// Process processes files and extracts file metadata
func (e *Processor) Do(ctx context.Context, jobCh chan Job, jobCount int) ([]ResultSuccess, []ResultError) {
	errorsCh := make(chan *ResultError, jobCount)
	resultsCh := make(chan *ResultSuccess, jobCount)

	contextWithTimeout, cancel := context.WithTimeout(ctx, e.ProcessTimeout)
	defer cancel()

	workersCount := e.countNeededWorkers(jobCount, e.MaxWorkers)
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
				case job, ok := <-jobCh:
					if !ok {
						return
					}
					res, err := e.processFile(job)
					if err != nil {
						errorsCh <- &ResultError{
							Error:    err,
							FilePath: job.FilePath,
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

// countNeededWorkers counts how many workers are needed
func (e *Processor) countNeededWorkers(filesCount, maxUploadWorkers int) int {
	if filesCount < maxUploadWorkers {
		return filesCount
	}
	return maxUploadWorkers
}

// processFile processes a single file
func (e *Processor) processFile(job Job) (*ResultSuccess, error) {
	glog.Infof("Processing file `%s`...\n", job.FilePath)

	res, err := e.workerFn(job)
	if err != nil {
		return nil, errors.Wrapf(err, "Error while processing file `%s`", job.FilePath)
	}

	result := &ResultSuccess{
		FilePath: job.FilePath,
		Output:   res,
	}

	return result, nil
}

// populateResults populates all results from all jobs
func (e *Processor) populateResults(resultsCh chan *ResultSuccess) []ResultSuccess {
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
func (e *Processor) populateErrors(errorsCh chan *ResultError) []ResultError {
	var errs []ResultError
	for uploadErr := range errorsCh {
		if uploadErr == nil {
			continue
		}

		errs = append(errs, *uploadErr)
	}

	return errs
}
