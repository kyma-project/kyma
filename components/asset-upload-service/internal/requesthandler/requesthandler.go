package requesthandler

import (
	"context"
	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/asset-upload-service/internal/bucket"
	"github.com/kyma-project/kyma/components/asset-upload-service/internal/uploader"
	"github.com/pkg/errors"
	"net/http"
	"time"
)

type RequestHandler struct {
	client           uploader.MinioClient
	uploadTimeout    time.Duration
	maxUploadWorkers int
	buckets          bucket.SystemBucketNames,
}

func New(client uploader.MinioClient, buckets bucket.SystemBucketNames, uploadTimeout time.Duration, maxUploadWorkers int) *RequestHandler {
	return &RequestHandler{
		client:           client,
		uploadTimeout:    uploadTimeout,
		maxUploadWorkers: maxUploadWorkers,
		buckets:          buckets,
	}
}

func (r RequestHandler) ServeHTTP(wr http.ResponseWriter, rq *http.Request) {
	err := rq.ParseMultipartForm(32 << 20)
	if err != nil {
		return
	}
	defer rq.MultipartForm.RemoveAll()

	//TODO: Handle directory param
	//directory := rq.MultipartForm.Value["directory"]

	privateFiles := rq.MultipartForm.File["private"]
	publicFiles := rq.MultipartForm.File["public"]
	filesCount := len(publicFiles) + len(privateFiles)

	u := uploader.New(r.client, r.uploadTimeout, r.maxUploadWorkers)

	fileToUploadCh := make(chan uploader.FileUpload, filesCount)

	go func() {
		for _, file := range publicFiles {
			fileToUploadCh <- uploader.FileUpload{
				Bucket: r.buckets.Public,
				File:   file,
			}
		}
		for _, file := range privateFiles {
			fileToUploadCh <- uploader.FileUpload{
				Bucket: r.buckets.Private,
				File:   file,
			}
		}
		close(fileToUploadCh)
	}()

	err = u.UploadFiles(context.Background(), fileToUploadCh, filesCount)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while uploading files"))
	}

	//TODO: Return results
	glog.Infof("Finished processing request with uploading %d files.", filesCount)
	wr.WriteHeader(http.StatusCreated)


}
