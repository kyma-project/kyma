package assethook

import (
	"bytes"
	"context"
	"github.com/pkg/errors"
	"io"
	"time"

	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
)

type validationEngine struct {
	processor httpProcessor
}

//go:generate mockery -name=Validator -output=automock -outpkg=automock -case=underscore
type Validator interface {
	Validate(ctx context.Context, basePath string, files []string, services []v1alpha2.AssetWebhookService) (Result, error)
}

func NewValidator(httpClient HttpClient, timeout time.Duration, workers int) *validationEngine {
	return &validationEngine{
		processor: &processor{
			timeout:        timeout,
			workers:        workers,
			onFail:         validationFailureHandler,
			continueOnFail: true,
			httpClient:     httpClient,
		},
	}
}

func validationFailureHandler(_ context.Context, _, filePath string, responseBody io.Reader, messagesChan chan Message, _ chan error) {
	buffer := new(bytes.Buffer)
	buffer.ReadFrom(responseBody)

	message := Message{Filename: filePath, Message: buffer.String()}
	messagesChan <- message
}

func (e *validationEngine) Validate(ctx context.Context, basePath string, files []string, services []v1alpha2.AssetWebhookService) (Result, error) {
	results, err := e.processor.Do(ctx, basePath, files, services)
	if err != nil {
		return Result{}, errors.Wrap(err, "while validating")
	}

	return Result{
		Success:  len(results) == 0,
		Messages: results,
	}, nil
}
