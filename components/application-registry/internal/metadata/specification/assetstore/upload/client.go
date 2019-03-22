package upload

import (
	"bytes"
	"encoding/json"
	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"io/ioutil"
	"mime/multipart"
	"net/http"
)

const (
	AssetFieldName = "assetFile"
)

type InputFile struct {
	Directory string
	Name      string
	Contents  []byte
}

type OutputFile struct {
	FileName   string
	RemotePath string
	Bucket     string
	Size       int
}

type Client interface {
	Upload(file InputFile) (OutputFile, apperrors.AppError)
}

type uploadClient struct {
	httpClient       *http.Client
	uploadServiceUrl string
}

func NewUploadClient(uploadServiceUrl string, httpclient *http.Client) Client {
	return uploadClient{
		uploadServiceUrl: uploadServiceUrl,
		httpClient:       httpclient,
	}
}

func (uc uploadClient) Upload(file InputFile) (OutputFile, apperrors.AppError) {

	req, err := uc.prepareRequest(file)
	if err != nil {
		return OutputFile{}, err
	}

	res, err := uc.executeRequest(req)
	if err != nil {
		return OutputFile{}, err
	}

	return uc.unmarshal(res)
}

func (uc uploadClient) prepareRequest(file InputFile) (*http.Request, apperrors.AppError) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile(AssetFieldName, file.Name)
	if err != nil {
		return nil, apperrors.Internal("Failed to create multipart content.")
	}

	_, err = part.Write(file.Contents)
	if err != nil {
		return nil, apperrors.Internal("Failed to write file contents.")
	}

	writer.Close()

	req, err := http.NewRequest(http.MethodPost, uc.uploadServiceUrl, body)
	if err != nil {
		return nil, apperrors.Internal("Failed to create request.")
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req, nil
}

func (uc uploadClient) executeRequest(r *http.Request) (*http.Response, apperrors.AppError) {
	res, err := uc.httpClient.Do(r)
	if err != nil {
		return nil, apperrors.Internal("Failed to execute request.")
	}

	switch res.StatusCode {
	case http.StatusOK:
		return res, nil
	case http.StatusNotFound:
		return nil, apperrors.NotFound("Upload service call failed.")
	default:
		return nil, apperrors.Internal("Failed to call Upload Service.")
	}
}

func (uc uploadClient) unmarshal(r *http.Response) (OutputFile, apperrors.AppError) {

	b, err := ioutil.ReadAll(r.Body)

	if err != nil {
		return OutputFile{}, apperrors.Internal("Failed to read request body.")
	}

	defer r.Body.Close()

	var outputFile OutputFile
	err = json.Unmarshal(b, &outputFile)

	if err != nil {
		return OutputFile{}, apperrors.Internal("Failed to unmarshal body.")
	}

	return outputFile, nil
}
