package processor

import (
	"context"
	"sync"
	"time"

	"github.com/kyma-project/kyma/components/asset-metadata-service/pkg/fileheader"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

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

var (
	processSingleFileHistogram = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "assetstore_metadata_service_process_single_file_duration_seconds",
		Help:    "Duration distribution of processing single file",
		Buckets: prometheus.ExponentialBuckets(0.00001, 2, 16), // discuss about loading those values from envs, not only here but generally
	})
	processFilesHistogram = promauto.NewHistogram(prometheus.HistogramOpts{
		Name: "assetstore_metadata_service_process_all_files_duration_seconds",
		Help: "Duration distribution of processing all files and extracting metadata",
	})
)

// New returns a new instance of Processor
func New(workerFn func(job Job) (interface{}, error), maxWorkers int, processTimeout time.Duration) *Processor {
	return &Processor{
		ProcessTimeout: processTimeout,
		MaxWorkers:     maxWorkers,
		workerFn:       workerFn,
	}
}

// Process processes files and extracts file metadata
func (p *Processor) Do(ctx context.Context, jobCh chan Job, jobCount int) ([]ResultSuccess, []ResultError) {
	start := time.Now()
	errorsCh := make(chan *ResultError, jobCount)
	successCh := make(chan *ResultSuccess, jobCount)

	contextWithTimeout, cancel := context.WithTimeout(ctx, p.ProcessTimeout)
	defer cancel()

	workersCount := p.countNeededWorkers(jobCount, p.MaxWorkers)
	glog.Infof("Creating %d concurrent worker(s)...", workersCount)
	var waitGroup sync.WaitGroup
	waitGroup.Add(workersCount)
	for i := 0; i < workersCount; i++ {
		go func() {
			defer waitGroup.Done()
			p.work(contextWithTimeout, jobCh, errorsCh, successCh)
		}()
	}

	waitGroup.Wait()
	close(successCh)
	close(errorsCh)

	result := p.populateResults(successCh)
	errs := p.populateErrors(errorsCh)

	processFilesHistogram.Observe(time.Since(start).Seconds())
	return result, errs
}

// countNeededWorkers counts how many workers are needed
func (p *Processor) countNeededWorkers(filesCount, maxUploadWorkers int) int {
	if filesCount < maxUploadWorkers {
		return filesCount
	}
	return maxUploadWorkers
}

func (p *Processor) work(context context.Context, jobCh chan Job, errorsCh chan *ResultError, successCh chan *ResultSuccess) {
	for {
		select {
		case <-context.Done():
			glog.Error(errors.Wrapf(context.Err(), "ResultError while concurrently processing file"))
			return
		default:
		}

		select {
		case job, ok := <-jobCh:
			if !ok {
				return
			}
			res, err := p.processFile(job)
			if err != nil {
				errorsCh <- &ResultError{
					Error:    err,
					FilePath: job.FilePath,
				}
			}

			if res != nil {
				successCh <- res
			}
		default:
		}
	}
}

// processFile processes a single file
func (p *Processor) processFile(job Job) (*ResultSuccess, error) {
	start := time.Now()
	glog.Infof("Processing file `%s`...\n", job.FilePath)

	res, err := p.workerFn(job)
	if err != nil {
		err = errors.Wrapf(err, "Error while processing file `%s`", job.FilePath)
		glog.Error(err)
		return nil, err
	}

	result := &ResultSuccess{
		FilePath: job.FilePath,
		Output:   res,
	}
	
	processSingleFileHistogram.Observe(time.Since(start).Seconds())
	return result, nil
}

// populateResults populates all results from all jobs
func (p *Processor) populateResults(resultsCh chan *ResultSuccess) []ResultSuccess {
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
func (p *Processor) populateErrors(errorsCh chan *ResultError) []ResultError {
	var errs []ResultError
	for uploadErr := range errorsCh {
		if uploadErr == nil {
			continue
		}

		errs = append(errs, *uploadErr)
	}

	return errs
}
