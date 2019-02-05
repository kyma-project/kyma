package webhook

import (
	"context"
	"fmt"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/apis/assetstore/v1alpha1"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/assethook"
	webhookv1alpha1 "github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/assethook/api/v1alpha1"
	"io/ioutil"
	"sync"
	"time"
)

//go:generate mockery -name=Validator -output=automock -outpkg=automock -case=underscore
type Validator interface {
	Validate(ctx context.Context, basePath string, files []string, asset *v1alpha1.Asset) (ValidationResult, error)
}

type validationWebhook struct {
	baseWebhook
	webhook    assethook.Webhook
	timeout    time.Duration
	fileReader func(filename string) ([]byte, error)
}

type ValidationResult struct {
	Success  bool
	Messages []string
}

func NewValidator(webhook assethook.Webhook, timeout time.Duration) Validator {
	return &validationWebhook{
		webhook:    webhook,
		timeout:    timeout,
		fileReader: ioutil.ReadFile,
	}
}

func (w *validationWebhook) Validate(ctx context.Context, basePath string, files []string, asset *v1alpha1.Asset) (ValidationResult, error) {
	errCh := make(chan error, 1)
	rspCh := make(chan webhookv1alpha1.ValidationResponse, 1)

	go func() {
		defer close(errCh)
		defer close(rspCh)

		validatorsCount := len(asset.Spec.Source.ValidationWebhookService)
		var wg sync.WaitGroup
		wg.Add(validatorsCount)

		for _, validation := range asset.Spec.Source.ValidationWebhookService {
			go w.validate(ctx, basePath, files, asset, validation, errCh, rspCh, &wg)
		}

		wg.Wait()
	}()

	var err error
	for er := range errCh {
		err = er
	}

	valid := true
	var messages []string
	for response := range rspCh {
		for key, status := range response.Status {
			if status.Status != webhookv1alpha1.ValidationSuccess {
				valid = false
				message := fmt.Sprintf("%s: %s", key, status.Message)
				messages = append(messages, message)
			}
		}
	}

	return ValidationResult{
		Success:  valid,
		Messages: messages,
	}, err
}

func (w *validationWebhook) validate(ctx context.Context, basePath string, files []string, asset *v1alpha1.Asset, service v1alpha1.AssetWebhookService, errCh chan<- error, rspCh chan<- webhookv1alpha1.ValidationResponse, wg *sync.WaitGroup) {
	defer wg.Done()

	context, cancel := context.WithTimeout(ctx, w.timeout)
	defer cancel()

	metadata := w.parseMetadata(service.Metadata)
	assets, err := w.readFiles(basePath, files, w.fileReader)
	if err != nil {
		errCh <- err
		return
	}

	request := &webhookv1alpha1.ValidationRequest{
		Name:      asset.Name,
		Namespace: asset.Namespace,
		Metadata:  metadata,
		Assets:    assets,
	}
	response := new(webhookv1alpha1.ValidationResponse)
	url := w.getWebhookUrl(service)

	err = w.webhook.Call(context, url, request, response)
	if err != nil {
		errCh <- err
		return
	}

	rspCh <- *response
}
