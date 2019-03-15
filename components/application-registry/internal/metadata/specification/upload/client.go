package upload

import (
	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"net/http"
	"net/url"
)

type InputFile struct {
	Directory string
	Name string
	Contents []byte
}

type OutputFile struct {
	FileName string
	RemotePath string
	Bucket string
	Size int
}

type UploadClient interface {
	Upload(file InputFile) (OutputFile, apperrors.AppError)
}

type uploadClient struct {

}

func NewUploadClient(uploadServiceUrl *url.URL, httpclient http.Client) UploadClient {
	return uploadClient{}
}

func (uc uploadClient) Upload(file InputFile) (OutputFile, apperrors.AppError){
	return OutputFile{},nil
}
