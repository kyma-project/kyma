package assethook

import (
	"bytes"
	"context"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
)

type mutationEngine struct {
	processor httpProcessor
}

//go:generate mockery -name=Mutator -output=automock -outpkg=automock -case=underscore
type Mutator interface {
	Mutate(ctx context.Context, basePath string, files []string, services []v1alpha2.AssetWebhookService) (Result, error)
}

func NewMutator(httpClient HttpClient, timeout time.Duration, workers int) Mutator {
	return &mutationEngine{
		processor: &processor{
			timeout:        timeout,
			workers:        workers,
			onFail:         mutationFailureHandler,
			onSuccess:      mutationSuccessHandler,
			continueOnFail: false,
			httpClient:     httpClient,
		},
	}
}

func mutationSuccessHandler(_ context.Context, basePath, filePath string, responseBody io.Reader, _ chan Message, errChan chan error) {
	path := filepath.Join(basePath, filePath)
	body, err := ioutil.ReadAll(responseBody)
	if err != nil {
		errChan <- errors.Wrap(err, "while reading response body")
		return
	}

	err = ioutil.WriteFile(path, body, os.ModePerm)
	if err != nil {
		errChan <- errors.Wrapf(err, "while writing file %s", path)
	}
}

func mutationFailureHandler(_ context.Context, _, filePath string, responseBody io.Reader, messagesChan chan Message, _ chan error) {
	buffer := new(bytes.Buffer)
	buffer.ReadFrom(responseBody)

	message := Message{Filename: filePath, Message: buffer.String()}
	messagesChan <- message
}

func (e *mutationEngine) Mutate(ctx context.Context, basePath string, files []string, services []v1alpha2.AssetWebhookService) (Result, error) {
	results, err := e.processor.Do(ctx, basePath, files, services)
	if err != nil {
		return Result{}, errors.Wrap(err, "while mutating")
	}

	return Result{
		Success:  len(results) == 0,
		Messages: results,
	}, nil
}
