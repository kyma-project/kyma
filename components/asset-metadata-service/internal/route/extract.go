package route

import (
	"context"
	"encoding/json"
	"github.com/kyma-project/kyma/components/asset-metadata-service/pkg/extractor"
	"github.com/kyma-project/kyma/components/asset-metadata-service/pkg/fileheader"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type ExtractHandler struct {
	maxWorkers int
	processTimeout time.Duration
}

// ResultError stores error data
type ResultError struct {
	FileName string `json:"fileName,omitempty"`
	Message  string `json:"message,omitempty"`
}

type Response struct {
	Data   []extractor.ResultSuccess `json:"uploadedFiles,omitempty"`
	Errors []ResultError             `json:"errors,omitempty"`
}

func NewExtractHandler(maxWorkers int, processTimeout time.Duration) *ExtractHandler {
	return &ExtractHandler{
		maxWorkers:maxWorkers,
		processTimeout:processTimeout,
	}
}

func (r *ExtractHandler) ServeHTTP(w http.ResponseWriter, rq *http.Request) {
	defer func() {
		err := rq.Body.Close()
		if err != nil {
			glog.Error(errors.Wrap(err, "while closing request body"))
		}
	}()

	err := rq.ParseMultipartForm(32 << 20) // 32MB
	if err != nil {
		wrappedErr := errors.Wrap(err, "while parsing multipart request")
		r.writeInternalError(w, wrappedErr)
		return
	}

	if rq.MultipartForm == nil {
		r.writeResponse(w, http.StatusBadRequest, Response{
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

	files := rq.MultipartForm.File["files"]
	filesLen := len(files)

	if filesLen == 0 {
		r.writeResponse(w, http.StatusBadRequest, Response{
			Errors: []ResultError{
				{
					Message: "No files specified. Use `files` field to provide them for processing.",
				},
			},
		})
		return
	}

	filesToProcessCh := r.chanFromFiles(files)

	e := extractor.New(r.maxWorkers, r.processTimeout)
	result, errs := e.Process(context.Background(), filesToProcessCh, filesLen)

	glog.Infof("Finished processing request with uploading %d files.", filesLen)

	var responseErrors []ResultError
	for _, err := range errs {
		responseErrors = append(responseErrors, ResultError{
			Message:  err.Error.Error(),
			FileName: err.FileName,
		})
	}

	var status int

	if len(responseErrors) == 0 {
		status = http.StatusOK
	} else if len(result) == 0 {
		status = http.StatusBadGateway
	} else {
		status = http.StatusMultiStatus
	}

	r.writeResponse(w, status, Response{
		Data:   result,
		Errors: responseErrors,
	})
}

func (r *ExtractHandler) chanFromFiles(files []*multipart.FileHeader) chan extractor.Job {
	filesCh := make(chan extractor.Job, len(files))

	go func() {
		defer close(filesCh)
		for _, file := range files {
			filesCh <- extractor.Job{
				File: fileheader.FromMultipart(file),
			}
		}
	}()

	return filesCh
}

func (r *ExtractHandler) writeResponse(w http.ResponseWriter, statusCode int, resp Response) {
	jsonResponse, err := json.Marshal(resp)
	if err != nil {
		wrappedErr := errors.Wrapf(err, "while marshalling JSON response")
		r.writeInternalError(w, wrappedErr)
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(statusCode)
	_, err = w.Write(jsonResponse)
	if err != nil {
		wrappedErr := errors.Wrapf(err, "while writing JSON response")
		glog.Error(wrappedErr)
	}
}

func (r *ExtractHandler) writeInternalError(w http.ResponseWriter, err error) {
	r.writeResponse(w, http.StatusInternalServerError, Response{
		Errors: []ResultError{
			{Message: err.Error()},
		},
	})
}
