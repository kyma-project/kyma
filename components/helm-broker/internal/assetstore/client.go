package assetstore

// Copied from https://github.com/kyma-project/kyma/tree/3640406442780b8f13c1b0c40de450256ac9be01/components/application-registry/internal/metadata/specification/assetstore/upload

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/kyma-project/kyma/components/application-gateway/pkg/httpconsts"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	publicFileField      = "public"
	endpointFormat       = "%s/v1/upload"
	uploadRequestTimeout = time.Duration(5 * time.Second)
)

// Response represents response from upload service
type Response struct {
	UploadedFiles []UploadedFile
	Errors        []ResponseError
}

// ResponseError represents error response from upload service
type ResponseError struct {
	Message  string
	FileName string
}

// UploadedFile contains data about uploaded file
type UploadedFile struct {
	FileName   string
	RemotePath string
	Bucket     string
	Size       int64
}

// Client defines upload service actions
//go:generate mockery -name=Client -output=automock -outpkg=automock -case=underscore
type Client interface {
	Upload(fileName string, contents []byte) (UploadedFile, error)
}

type uploadClient struct {
	httpClient       http.Client
	uploadServiceURL string

	log logrus.FieldLogger
}

// NewClient creates a new client for upload service
func NewClient(uploadServiceURL string, log logrus.FieldLogger) Client {
	return uploadClient{
		uploadServiceURL: uploadServiceURL,
		httpClient: http.Client{
			Timeout: uploadRequestTimeout,
		},
		log: log,
	}
}

// Upload uploads a file to upload service
func (uc uploadClient) Upload(fileName string, contents []byte) (UploadedFile, error) {
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

func (uc uploadClient) prepareRequest(fileName string, contents []byte) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	{
		err := uc.prepareMultipartForm(body, writer, fileName, contents)
		if err != nil {
			return nil, err
		}
	}

	url := fmt.Sprintf(endpointFormat, uc.uploadServiceURL)

	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create request")
	}

	req.Header.Set(httpconsts.HeaderContentType, writer.FormDataContentType())

	return req, nil
}

func (uc uploadClient) prepareMultipartForm(body *bytes.Buffer, writer *multipart.Writer, fileName string, contents []byte) error {
	defer writer.Close()

	publicFilePart, err := writer.CreateFormFile(publicFileField, fileName)
	if err != nil {
		return errors.Wrap(err, "Failed to create multipart content")
	}

	_, err = publicFilePart.Write(contents)
	if err != nil {
		return errors.Wrap(err, "Failed to write file contents")
	}

	return nil
}

func (uc uploadClient) executeRequest(r *http.Request) (*http.Response, error) {
	res, err := uc.httpClient.Do(r)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to execute request.")
	}

	return res, nil
}

func (uc uploadClient) unmarshal(r *http.Response) (Response, error) {
	b, err := ioutil.ReadAll(r.Body)

	if err != nil {
		return Response{}, errors.Wrap(err, "Failed to read request body")
	}

	defer r.Body.Close()

	var uploadResponse Response
	err = json.Unmarshal(b, &uploadResponse)

	if err != nil {
		return Response{}, errors.Wrap(err, "Failed to unmarshal body")
	}

	return uploadResponse, nil
}

func (uc uploadClient) extract(fileName string, response Response) (UploadedFile, error) {
	if len(response.UploadedFiles) == 1 {
		return response.UploadedFiles[0], nil
	}

	for _, e := range response.Errors {
		uc.log.Errorf("Failed to upload %s file with Upload Service, %s.", e.FileName, e.Message)
	}
	return UploadedFile{}, errors.Errorf("Failed to extract %s file from response.", fileName)
}
