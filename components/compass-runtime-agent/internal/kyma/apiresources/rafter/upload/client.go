package upload

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"mime/multipart"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
	"kyma-project.io/compass-runtime-agent/internal/apperrors"
	"kyma-project.io/compass-runtime-agent/internal/httpconsts"
)

const (
	PublicFileField      = "public"
	EndpointFormat       = "%s/v1/upload"
	uploadRequestTimeout = time.Duration(5 * time.Second)

	directorField = "directory"
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

//go:generate mockery -name=Client
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
		httpClient: http.Client{
			Timeout: uploadRequestTimeout,
		},
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
		return nil, apperrors.Internal("Failed to create request, %s.", err)
	}

	req.Header.Set(httpconsts.HeaderContentType, writer.FormDataContentType())

	return req, nil
}

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func (uc uploadClient) prepareMultipartForm(body *bytes.Buffer, writer *multipart.Writer, fileName string, contents []byte) apperrors.AppError {
	defer func() {
		err := writer.Close()
		if err != nil {
			log.Error("Failed to close multipart writer.")
		}
	}()

	directory := randStringBytes(10)
	err := writer.WriteField(directorField, directory)
	if err != nil {
		return apperrors.Internal("Failed to write directory field, %s.", err.Error())
	}

	publicFilePart, err := writer.CreateFormFile(PublicFileField, fileName)
	if err != nil {
		return apperrors.Internal("Failed to create multipart content, %s.", err.Error())
	}

	_, err = publicFilePart.Write(contents)
	if err != nil {
		return apperrors.Internal("Failed to write file contents, %s.", err.Error())
	}

	return nil
}

func (uc uploadClient) executeRequest(r *http.Request) (*http.Response, apperrors.AppError) {
	res, err := uc.httpClient.Do(r)
	if err != nil {
		log.Errorf("Failed to execute request, %s.", err.Error())
		return nil, apperrors.Internal("Failed to execute request.")
	}

	return res, nil
}

func (uc uploadClient) unmarshal(r *http.Response) (Response, apperrors.AppError) {
	b, err := ioutil.ReadAll(r.Body)

	if err != nil {
		return Response{}, apperrors.Internal("Failed to read request body, %s.", err)
	}

	defer func() {
		err := r.Body.Close()
		if err != nil {
			log.Error("Failed to close response body.")
		}
	}()

	var uploadResponse Response
	err = json.Unmarshal(b, &uploadResponse)

	if err != nil {
		return Response{}, apperrors.Internal("Failed to unmarshal body, %s", err)
	}

	return uploadResponse, nil
}

func (uc uploadClient) extract(fileName string, response Response) (UploadedFile, apperrors.AppError) {
	if len(response.UploadedFiles) == 1 {
		return response.UploadedFiles[0], nil
	}

	for _, e := range response.Errors {
		log.Errorf("Failed to upload %s file with Upload Service, %s.", e.FileName, e.Message)
	}
	return UploadedFile{}, apperrors.Internal("Failed to extract %s file from response.", fileName)
}
