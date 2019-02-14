package requesthandler

import (
	"context"
	"encoding/json"
	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/asset-upload-service/internal/bucket"
	"github.com/kyma-project/kyma/components/asset-upload-service/internal/fileheader"
	"github.com/kyma-project/kyma/components/asset-upload-service/internal/uploader"
	"github.com/pkg/errors"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"
)

type RequestHandler struct {
	client           uploader.MinioClient
	uploadTimeout    time.Duration
	maxUploadWorkers int
	buckets          bucket.SystemBucketNames
	uploadOrigin     string
}

type response struct {
	UploadedFiles []uploader.UploadResult `json:"uploadedFiles"`
	Errors        []string                `json:"errors"`
}

func New(client uploader.MinioClient, buckets bucket.SystemBucketNames, uploadOrigin string, uploadTimeout time.Duration, maxUploadWorkers int) *RequestHandler {
	return &RequestHandler{
		client:           client,
		uploadTimeout:    uploadTimeout,
		maxUploadWorkers: maxUploadWorkers,
		buckets:          buckets,
		uploadOrigin:     uploadOrigin,
	}
}

func (r *RequestHandler) ServeHTTP(w http.ResponseWriter, rq *http.Request) {
	err := rq.ParseMultipartForm(32 << 20)
	if err != nil {
		wrappedErr := errors.Wrap(err, "while parsing multipart request")
		r.writeInternalError(w, wrappedErr)
	}
	defer rq.MultipartForm.RemoveAll()

	directoryValues := rq.MultipartForm.Value["directory"]

	var directory string
	if directoryValues == nil {
		directory = r.generateDirectoryName()
	} else {
		directory = directoryValues[0]
	}

	privateFiles := rq.MultipartForm.File["private"]
	publicFiles := rq.MultipartForm.File["public"]
	filesCount := len(publicFiles) + len(privateFiles)

	u := uploader.New(r.client, r.uploadOrigin, r.uploadTimeout, r.maxUploadWorkers)
	fileToUploadCh := r.populateFilesChannel(publicFiles, privateFiles, filesCount, directory)
	uploadedFiles, errs := u.UploadFiles(context.Background(), fileToUploadCh, filesCount)

	glog.Infof("Finished processing request with uploading %d files.", filesCount)

	var errMessages []string
	for _, err := range errs {
		errMessages = append(errMessages, err.Error())
	}
	r.writeResponse(w, response{
		UploadedFiles: uploadedFiles,
		Errors:        errMessages,
	})
}

func (r *RequestHandler) generateDirectoryName() string {
	unixTime := time.Now().Unix()
	return strconv.FormatInt(unixTime, 32)
}

func (r *RequestHandler) populateFilesChannel(publicFiles, privateFiles []*multipart.FileHeader, filesCount int, directory string) chan uploader.FileUpload {
	filesCh := make(chan uploader.FileUpload, filesCount)

	go func() {
		for _, file := range publicFiles {
			filesCh <- uploader.FileUpload{
				Bucket:    r.buckets.Public,
				File:      fileheader.FromMultipart(file),
				Directory: directory,
			}
		}
		for _, file := range privateFiles {
			filesCh <- uploader.FileUpload{
				Bucket:    r.buckets.Private,
				File:      fileheader.FromMultipart(file),
				Directory: directory,
			}
		}
		close(filesCh)
	}()

	return filesCh
}

func (r *RequestHandler) writeResponse(w http.ResponseWriter, resp response) {
	jsonResponse, err := json.Marshal(resp)
	if err != nil {
		wrappedErr := errors.Wrapf(err, "while marshalling JSON response")
		r.writeInternalError(w, wrappedErr)
	}

	w.Header().Set("Content-Type", "application/json")

	var status int
	if len(resp.Errors) > 0 {
		status = http.StatusInternalServerError
	} else {
		status = http.StatusCreated
	}

	w.WriteHeader(status)
	_, err = w.Write(jsonResponse)
	if err != nil {
		wrappedErr := errors.Wrapf(err, "while writing JSON response")
		r.writeInternalError(w, wrappedErr)
	}
}

func (r *RequestHandler) writeInternalError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
