package upload

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"mime/multipart"
	"net/http"
)

const (
	PrivateFileField = "private"
	DirectoryField   = "directory"
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

func NewClient(uploadServiceUrl string, httpclient *http.Client) Client {
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
	{
		err := uc.prepareMultipartForm(body, writer, file)
		if err != nil {
			return nil, err
		}
	}

	url := fmt.Sprintf("%s/%s", uc.uploadServiceUrl, "v1/upload")

	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, apperrors.Internal("Failed to create request.")
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req, nil
}

func (uc uploadClient) prepareMultipartForm(body *bytes.Buffer, writer *multipart.Writer, file InputFile) apperrors.AppError {
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
		return apperrors.Internal("Failed to write directory: %s.", err.Error())
	}

	return nil
}

func (uc uploadClient) executeRequest(r *http.Request) (*http.Response, apperrors.AppError) {
	res, err := uc.httpClient.Do(r)
	if err != nil {
		log.Errorf("Failed to call Upload Service: %s", err.Error())
		return nil, apperrors.Internal("Failed to execute request.")
	}

	if res.StatusCode == http.StatusOK {
		return res, nil
	} else {
		log.Errorf("Failed to call Upload Service: unexpected status: %s", res.Status)
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
