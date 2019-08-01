package assethook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/assethook/api/v1alpha1"
	pkgPath "github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/path"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

//go:generate mockery -name=MetadataExtractor -output=automock -outpkg=automock -case=underscore
type MetadataExtractor interface {
	Extract(ctx context.Context, basePath string, files []string, services []v1alpha2.WebhookService) ([]File, error)
}

type File struct {
	Name     string
	Metadata *json.RawMessage
}

type metadataEngine struct {
	timeout    time.Duration
	fileReader func(filename string) ([]byte, error)
	httpClient HttpClient
}

func NewMetadataExtractor(httpClient HttpClient, timeout time.Duration) MetadataExtractor {
	return &metadataEngine{
		httpClient: httpClient,
		timeout:    timeout,
		fileReader: ioutil.ReadFile,
	}
}

func (e *metadataEngine) Extract(ctx context.Context, basePath string, files []string, services []v1alpha2.WebhookService) ([]File, error) {
	results := make(map[string]*json.RawMessage)
	for _, service := range services {
		filtered, err := pkgPath.Filter(files, service.Filter)
		if err != nil {
			return nil, errors.Wrapf(err, "while filtering files with regex %s", service.Filter)
		}

		body, contentType, err := e.buildQuery(basePath, filtered)
		if err != nil {
			return nil, errors.Wrap(err, "while building multipart query")
		}

		response := &v1alpha1.MetadataResponse{}
		err = e.do(ctx, contentType, service, body, response)
		if err != nil {
			return nil, errors.Wrap(err, "while sending request to metadata webhook")
		}

		results = e.replaceMetadata(results, response.Data)
	}

	return e.toFiles(results), nil
}

func (*metadataEngine) replaceMetadata(current map[string]*json.RawMessage, results []v1alpha1.MetadataResultSuccess) map[string]*json.RawMessage {
	for _, result := range results {
		current[result.FilePath] = result.Metadata
	}

	return current
}

func (*metadataEngine) toFiles(results map[string]*json.RawMessage) []File {
	files := make([]File, 0, len(results))
	for k, v := range results {
		files = append(files, File{Name: k, Metadata: v})
	}

	return files
}

func (e *metadataEngine) buildQuery(basePath string, files []string) (io.Reader, string, error) {
	b := &bytes.Buffer{}
	formWriter := multipart.NewWriter(b)
	defer formWriter.Close()

	for _, file := range files {
		path := filepath.Join(basePath, file)
		if err := e.buildQueryField(formWriter, file, path); err != nil {
			return nil, "", errors.Wrapf(err, "while building query part")
		}
	}

	return b, formWriter.FormDataContentType(), nil
}

func (e *metadataEngine) buildQueryField(writer *multipart.Writer, filename, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return errors.Wrapf(err, "while opening file %s", filename)
	}
	defer file.Close()

	part, err := writer.CreateFormFile(filename, filepath.Base(file.Name()))
	if err != nil {
		return errors.Wrapf(err, "while creating part for file %s", filename)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return errors.Wrapf(err, "while copying file %s to part", filename)
	}

	return nil
}

func (e *metadataEngine) do(ctx context.Context, contentType string, webhook v1alpha2.WebhookService, body io.Reader, response interface{}) error {
	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	req, err := http.NewRequest("POST", e.getWebhookUrl(webhook), body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", contentType)
	req.WithContext(ctx)

	rsp, err := e.httpClient.Do(req)
	if err != nil {
		return errors.Wrapf(err, "while sending request to webhook")
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid response from %s, code: %d", req.URL, rsp.StatusCode)
	}

	responseBytes, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return errors.Wrapf(err, "while reading response body")
	}

	err = json.Unmarshal(responseBytes, response)
	if err != nil {
		return errors.Wrapf(err, "while parsing response body")
	}

	return nil
}

func (*metadataEngine) getWebhookUrl(service v1alpha2.WebhookService) string {
	return fmt.Sprintf("http://%s.%s.svc.cluster.local%s", service.Name, service.Namespace, service.Endpoint)
}
