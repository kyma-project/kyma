package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/assethook"
	webhookv1alpha1 "github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/assethook/api/v1alpha1"
	pkgPath "github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/path"
	"github.com/pkg/errors"
)

//go:generate mockery -name=Validator -output=automock -outpkg=automock -case=underscore
type Validator interface {
	Validate(ctx context.Context, object Accessor, basePath string, files []string, services []v1alpha2.AssetWebhookService) (ValidationResult, error)
}

type validationEngine struct {
	baseEngine
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
	return &validationEngine{
		webhook:    webhook,
		timeout:    timeout,
		fileReader: ioutil.ReadFile,
	}
}

// TODO: Validation should be executed in concurrency
func (e *validationEngine) Validate(ctx context.Context, object Accessor, basePath string, files []string, services []v1alpha2.AssetWebhookService) (ValidationResult, error) {
	assetName := object.GetName()
	assetNamespace := object.GetNamespace()

	passed := true
	var errorMessages []string
	results := make(map[string][]ValidationMessage)
	for _, service := range services {
		filtered, err := pkgPath.Filter(files, service.Filter)
		assets, err := e.readFiles(basePath, filtered, e.fileReader)
		if err != nil {
			return ValidationResult{
				Success: false,
			}, err
		}

		metadata := e.parseMetadata(service.Metadata)

		request := &webhookv1alpha1.ValidationRequest{
			Name:      assetName,
			Namespace: assetNamespace,
			Metadata:  metadata,
			Assets:    assets,
		}

		response, err := e.validate(ctx, service, request)
		if err != nil {
			errorMessages = append(errorMessages, err.Error())
			continue
		}

		messages := e.parseResponse(response)
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

func (e *validationEngine) parseResponse(response *webhookv1alpha1.ValidationResponse) []ValidationMessage {
	var messages []ValidationMessage
	for key, status := range response.Status {
		if status.Status != webhookv1alpha1.ValidationSuccess {
			messages = append(messages, ValidationMessage{Filename: key, Message: status.Message})
		}
	}

	return messages
}

func (e *validationEngine) validate(ctx context.Context, service v1alpha2.AssetWebhookService, request *webhookv1alpha1.ValidationRequest) (*webhookv1alpha1.ValidationResponse, error) {
	jsonBytes, err := json.Marshal(request)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting request to JSON")
	}

	response := new(webhookv1alpha1.ValidationResponse)
	err = e.webhook.Do(ctx, "application/json", service.WebhookService, bytes.NewBuffer(jsonBytes), response, e.timeout)
	if err != nil {
		return nil, errors.Wrap(err, "while sending validation request")
	}

	return response, nil

}
