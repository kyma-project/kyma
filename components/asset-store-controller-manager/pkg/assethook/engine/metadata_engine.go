package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/assethook"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/assethook/api/v1alpha1"
	pkgPath "github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/path"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"
)

//go:generate mockery -name=MetadataExtractor -output=automock -outpkg=automock -case=underscore
type MetadataExtractor interface {
	Extract(ctx context.Context, object Accessor, basePath string, files []string, services []v1alpha2.WebhookService) ([]File, error)
}

type File struct {
	Name     string
	Metadata *json.RawMessage
}

type metadataEngine struct {
	webhook    assethook.Webhook
	timeout    time.Duration
	fileReader func(filename string) ([]byte, error)
}

func NewMetadataExtractor(webhook assethook.Webhook, timeout time.Duration) MetadataExtractor {
	return &metadataEngine{
		webhook:    webhook,
		timeout:    timeout,
		fileReader: ioutil.ReadFile,
	}
}

func (e *metadataEngine) Extract(ctx context.Context, object Accessor, basePath string, files []string, services []v1alpha2.WebhookService) ([]File, error) {
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
		err = e.webhook.Do(ctx, contentType, service, body, response, e.timeout)
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
