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
	PrivateFileField = "private"
	DirectoryField   = "directory"
	EndpointFormat   = "%s/v1/upload"
)

type File struct {
	Directory string
	Name      string
	Contents  []byte
}

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
	Upload(file File) (UploadedFile, apperrors.AppError)
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

func (uc uploadClient) Upload(file File) (UploadedFile, apperrors.AppError) {
	req, err := uc.prepareRequest(file)
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

	return uc.extract(file, uploadRes)
}

func (uc uploadClient) prepareRequest(file File) (*http.Request, apperrors.AppError) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	{
		err := uc.prepareMultipartForm(body, writer, file)
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

func (uc uploadClient) prepareMultipartForm(body *bytes.Buffer, writer *multipart.Writer, file File) apperrors.AppError {
	defer writer.Close()

	privateFilePart, err := writer.CreateFormFile(PrivateFileField, file.Name)
	if err != nil {
		return apperrors.Internal("Failed to create multipart content: %s.", err.Error())
	}

	_, err = privateFilePart.Write(file.Contents)
	if err != nil {
		return apperrors.Internal("Failed to write file contents: %s.", err.Error())
	}

	directoryPart, err := writer.CreateFormField(DirectoryField)
	if err != nil {
		return apperrors.Internal("Failed to create multipart content: %s.", err.Error())
	}

	_, err = directoryPart.Write([]byte(file.Directory))
	if err != nil {
		return apperrors.Internal("Failed to write directory name: %s.", err.Error())
	}

	return nil
}

func (uc uploadClient) executeRequest(r *http.Request) (*http.Response, apperrors.AppError) {
	res, err := uc.httpClient.Do(r)
	if err != nil {
		log.Errorf("Failed to execute request: %s.", err.Error())
		return nil, apperrors.Internal("Failed to execute request.")
	}

	if res.StatusCode == http.StatusOK {
		return res, nil
	} else {
		log.Errorf("Upload Service returned unexpected status: %s.", res.Status)
		return nil, apperrors.Internal("Failed to call Upload Service.")
	}
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

func (uc uploadClient) extract(inputFile File, response Response) (UploadedFile, apperrors.AppError) {
	if len(response.UploadedFiles) == 1 {
		return response.UploadedFiles[0], nil
	} else {
		for _, e := range response.Errors {
			log.Errorf("Failed to upload file %s with Upload Service: %s.", e.FileName, e.Message)
		}

		return UploadedFile{}, apperrors.Internal("Failed to extract file %s from response.", inputFile.Name)
	}
}
