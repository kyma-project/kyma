package upload

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-registry/internal/httpconsts"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"mime/multipart"
	"net/http"
)

const (
	PublicFileField = "public"
	EndpointFormat  = "%s/v1/upload"
)

type Response struct {
	UploadedFiles []UploadedFile
	Errors        []ResponseError
}

type ResponseError struct {
	Message  string
	FileName string
}

type UploadedFile struct {
	FileName   string
	RemotePath string
	Bucket     string
	Size       int64
}

type Client interface {
	Upload(fileName string, contents []byte) (UploadedFile, apperrors.AppError)
}

type uploadClient struct {
	httpClient       http.Client
	uploadServiceUrl string
}

func NewClient(uploadServiceUrl string) Client {
	return uploadClient{
		uploadServiceUrl: uploadServiceUrl,
		httpClient:       http.Client{},
	}
}

func (uc uploadClient) Upload(fileName string, contents []byte) (UploadedFile, apperrors.AppError) {
	req, err := uc.prepareRequest(fileName, contents)
	if err != nil {
		return UploadedFile{}, err
	}

	httpRes, err := uc.executeRequest(req)
	if err != nil {
		return UploadedFile{}, err
	}

	uploadRes, err := uc.unmarshal(httpRes)
	if err != nil {
		return UploadedFile{}, err
	}

	return uc.extract(fileName, uploadRes)
}

func (uc uploadClient) prepareRequest(fileName string, contents []byte) (*http.Request, apperrors.AppError) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	{
		err := uc.prepareMultipartForm(body, writer, fileName, contents)
		if err != nil {
			return nil, err
		}
	}

	url := fmt.Sprintf(EndpointFormat, uc.uploadServiceUrl)

	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, apperrors.Internal("Failed to create request.")
	}

	req.Header.Set(httpconsts.HeaderContentType, writer.FormDataContentType())

	return req, nil
}

func (uc uploadClient) prepareMultipartForm(body *bytes.Buffer, writer *multipart.Writer, fileName string, contents []byte) apperrors.AppError {
	defer writer.Close()

	publicFilePart, err := writer.CreateFormFile(PublicFileField, fileName)
	if err != nil {
		return apperrors.Internal("Failed to create multipart content: %s.", err.Error())
	}

	_, err = publicFilePart.Write(contents)
	if err != nil {
		return apperrors.Internal("Failed to write file contents: %s.", err.Error())
	}

	return nil
}

func (uc uploadClient) executeRequest(r *http.Request) (*http.Response, apperrors.AppError) {
	res, err := uc.httpClient.Do(r)
	if err != nil {
		log.Errorf("Failed to execute request: %s.", err.Error())
		return nil, apperrors.Internal("Failed to execute request.")
	}

	return res, nil
}

func (uc uploadClient) unmarshal(r *http.Response) (Response, apperrors.AppError) {
	b, err := ioutil.ReadAll(r.Body)

	if err != nil {
		return Response{}, apperrors.Internal("Failed to read request body.")
	}

	defer r.Body.Close()

	var uploadResponse Response
	err = json.Unmarshal(b, &uploadResponse)

	if err != nil {
		return Response{}, apperrors.Internal("Failed to unmarshal body.")
	}

	return uploadResponse, nil
}

func (uc uploadClient) extract(fileName string, response Response) (UploadedFile, apperrors.AppError) {
	if len(response.UploadedFiles) == 1 {
		return response.UploadedFiles[0], nil
	} else {
		for _, e := range response.Errors {
			log.Errorf("Failed to upload file %s with Upload Service: %s.", e.FileName, e.Message)
		}

		return UploadedFile{}, apperrors.Internal("Failed to extract file %s from response.", fileName)
	}
}
