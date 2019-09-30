package route

import (
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"

	"github.com/kyma-project/kyma/components/asset-metadata-service/pkg/extractor"
	"github.com/kyma-project/kyma/components/asset-metadata-service/pkg/fileheader"
	"github.com/kyma-project/kyma/components/asset-metadata-service/pkg/processor"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/golang/glog"
	"github.com/pkg/errors"
)

var (
	httpServeHistogram = promauto.NewHistogram(prometheus.HistogramOpts{
		Name: "assetstore_metadata_service_http_request_duration_seconds",
		Help: "Request's duration distribution",
	})
	statusCodesCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "assetstore_metadata_service_http_request_returned_status_code",
		Help: "Service's HTTP response status code",
	}, []string{"status_code"})
)

func incrementStatusCounter(status int) {
	statusCodesCounter.WithLabelValues(strconv.Itoa(status)).Inc()
}

type ExtractHandler struct {
	maxWorkers     int
	processTimeout time.Duration

	metadataExtractor extractor.Extractor
}

// ResultError stores error data
type ResultError struct {
	FilePath string `json:"filePath,omitempty"`
	Message  string `json:"message,omitempty"`
}

// ResultSuccess stores success data
type ResultSuccess struct {
	FilePath string                 `json:"filePath,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type Response struct {
	Data   []ResultSuccess `json:"data,omitempty"`
	Errors []ResultError   `json:"errors,omitempty"`
}

func NewExtractHandler(maxWorkers int, processTimeout time.Duration) *ExtractHandler {
	return &ExtractHandler{
		maxWorkers:        maxWorkers,
		processTimeout:    processTimeout,
		metadataExtractor: extractor.New(),
	}
}

func (h *ExtractHandler) ServeHTTP(w http.ResponseWriter, rq *http.Request) {
	start := time.Now()

	defer func() {
		err := rq.Body.Close()
		if err != nil {
			glog.Error(errors.Wrap(err, "while closing request body"))
		}
	}()

	err := rq.ParseMultipartForm(32 << 20) // 32MB
	if err != nil {
		wrappedErr := errors.Wrap(err, "while parsing multipart request")
		h.writeInternalError(w, wrappedErr)
		return
	}

	if rq.MultipartForm == nil {
		status := http.StatusBadRequest
		incrementStatusCounter(status)
		h.writeResponse(w, status, Response{
			Errors: []ResultError{
				{
					Message: "No multipart/form-data form received.",
				},
			},
		})
		return
	}

	defer func() {
		err := rq.MultipartForm.RemoveAll()
		if err != nil {
			glog.Error(errors.Wrap(err, "while removing files loaded from multipart form"))
		}
	}()

	jobCh, jobsCount, err := h.chanFromFormFiles(rq.MultipartForm.File)
	if err != nil {
		status := http.StatusBadRequest
		incrementStatusCounter(status)
		h.writeResponse(w, status, Response{
			Errors: []ResultError{
				{
					Message: err.Error(),
				},
			},
		})
		return
	}

	processFn := func(job processor.Job) (interface{}, error) {
		return h.metadataExtractor.ReadMetadata(job.File)
	}

	e := processor.New(processFn, h.maxWorkers, h.processTimeout)
	succ, errs := e.Do(context.Background(), jobCh, jobsCount)

	glog.Infof("Finished processing request with %d files attached.", jobsCount)

	response := h.convertToResponse(succ, errs)

	var status int

	if len(response.Errors) == 0 {
		status = http.StatusOK
		incrementStatusCounter(status)
	} else if len(response.Data) == 0 {
		status = http.StatusUnprocessableEntity
		incrementStatusCounter(status)
	} else {
		status = http.StatusMultiStatus
		incrementStatusCounter(status)
	}

	h.writeResponse(w, status, response)

	httpServeHistogram.Observe(time.Since(start).Seconds())
}

func (h *ExtractHandler) chanFromFormFiles(fileFields map[string][]*multipart.FileHeader) (chan processor.Job, int, error) {
	var jobs []processor.Job

	for key, files := range fileFields {
		if len(files) > 1 {
			return nil, 0, fmt.Errorf("Multiple files assigned to a single field %s .", key)
		}

		if len(files) == 0 || files[0] == nil {
			continue
		}

		if files[0].Size == 0 {
			continue
		}

		jobs = append(jobs, processor.Job{
			FilePath: key,
			File:     fileheader.FromMultipart(files[0]),
		})
	}

	jobsCount := len(jobs)
	if jobsCount == 0 {
		return nil, jobsCount, errors.New("No files sent with form.")
	}

	jobsCh := make(chan processor.Job, jobsCount)
	go func() {
		defer close(jobsCh)
		for _, job := range jobs {
			jobsCh <- job
		}
	}()

	return jobsCh, jobsCount, nil
}

func (h *ExtractHandler) convertToResponse(successes []processor.ResultSuccess, errs []processor.ResultError) Response {
	var responseData []ResultSuccess
	for _, succ := range successes {
		metadata, ok := succ.Output.(map[string]interface{})
		if !ok {
			glog.Errorf("Invalid conversion for extracted metadata from file %s: %+v", succ.FilePath, succ.Output)
			continue
		}

		responseData = append(responseData, ResultSuccess{
			FilePath: succ.FilePath,
			Metadata: metadata,
		})
	}

	var responseErrors []ResultError
	for _, err := range errs {
		responseErrors = append(responseErrors, ResultError{
			FilePath: err.FilePath,
			Message:  err.Error.Error(),
		})
	}

	return Response{
		Data:   responseData,
		Errors: responseErrors,
	}
}

func (h *ExtractHandler) writeResponse(w http.ResponseWriter, statusCode int, resp Response) {
	jsonResponse, err := json.Marshal(resp)
	if err != nil {
		wrappedErr := errors.Wrapf(err, "while marshalling JSON response")
		h.writeInternalError(w, wrappedErr)
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(statusCode)
	_, err = w.Write(jsonResponse)
	if err != nil {
		wrappedErr := errors.Wrapf(err, "while writing JSON response")
		glog.Error(wrappedErr)
	}
}

func (h *ExtractHandler) writeInternalError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	incrementStatusCounter(status)
	h.writeResponse(w, status, Response{
		Errors: []ResultError{
			{Message: err.Error()},
		},
	})
}
