package webhook

import (
	"context"
	"fmt"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/apis/assetstore/v1alpha1"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/assethook"
	webhookv1alpha1 "github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/assethook/api/v1alpha1"
	"github.com/pkg/errors"
	"io/ioutil"
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
	Messages map[string][]ValidationMessage
}

type ValidationMessage struct {
	Filename string
	Message  string
}

func NewValidator(webhook assethook.Webhook, timeout time.Duration) Validator {
	return &validationWebhook{
		webhook:    webhook,
		timeout:    timeout,
		fileReader: ioutil.ReadFile,
	}
}

// TODO: Validation should be executed in concurrency
func (w *validationWebhook) Validate(ctx context.Context, basePath string, files []string, asset *v1alpha1.Asset) (ValidationResult, error) {
	assets, err := w.readFiles(basePath, files, w.fileReader)
	if err != nil {
		return ValidationResult{
			Success: false,
		}, err
	}

	passed := true
	var errorMessages []string
	results := make(map[string][]ValidationMessage)
	for _, service := range asset.Spec.Source.ValidationWebhookService {
		metadata := w.parseMetadata(service.Metadata)

		request := &webhookv1alpha1.ValidationRequest{
			Name:      asset.Name,
			Namespace: asset.Namespace,
			Metadata:  metadata,
			Assets:    assets,
		}
		url := w.getWebhookUrl(service)

		response, err := w.validate(ctx, url, request)
		if err != nil {
			errorMessages = append(errorMessages, err.Error())
			continue
		}

		messages := w.parseResponse(response)
		if len(messages) > 0 {
			passed = false
			name := fmt.Sprintf("%s/%s%s", service.Namespace, service.Name, service.Endpoint)
			results[name] = messages
		}
	}

	if len(errorMessages) > 0 {
		return ValidationResult{
			Success: false,
		}, fmt.Errorf("error during validation: %+v", errorMessages)
	}

	return ValidationResult{
		Success:  passed,
		Messages: results,
	}, nil
}

func (w *validationWebhook) parseResponse(response *webhookv1alpha1.ValidationResponse) []ValidationMessage {
	var messages []ValidationMessage
	for key, status := range response.Status {
		if status.Status != webhookv1alpha1.ValidationSuccess {
			messages = append(messages, ValidationMessage{Filename: key, Message: status.Message})
		}
	}

	return messages
}

func (w *validationWebhook) validate(ctx context.Context, url string, request *webhookv1alpha1.ValidationRequest) (*webhookv1alpha1.ValidationResponse, error) {
	context, cancel := context.WithTimeout(ctx, w.timeout)
	defer cancel()

	response := new(webhookv1alpha1.ValidationResponse)
	err := w.webhook.Call(context, url, request, response)
	if err != nil {
		return nil, errors.Wrap(err, "while sending validation request")
	}

	return response, nil

}
