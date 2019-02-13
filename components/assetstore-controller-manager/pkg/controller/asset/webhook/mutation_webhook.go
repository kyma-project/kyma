package webhook

import (
	"context"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/apis/assetstore/v1alpha1"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/assethook"
	assethookv1alpha1 "github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/assethook/api/v1alpha1"
	"github.com/pkg/errors"
	"io/ioutil"
	"time"
)

//go:generate mockery -name=Mutator -output=automock -outpkg=automock -case=underscore
type Mutator interface {
	Mutate(ctx context.Context, basePath string, files []string, asset *v1alpha1.Asset) error
}

type mutationWebhook struct {
	baseWebhook
	webhook    assethook.Webhook
	timeout    time.Duration
	fileReader func(filename string) ([]byte, error)
}

func NewMutator(webhook assethook.Webhook, timeout time.Duration) Mutator {
	return &mutationWebhook{
		webhook:    webhook,
		timeout:    timeout,
		fileReader: ioutil.ReadFile,
	}
}

func (w *mutationWebhook) Mutate(ctx context.Context, basePath string, files []string, asset *v1alpha1.Asset) error {
	for _, service := range asset.Spec.Source.MutationWebhookService {
		metadata := w.parseMetadata(service.Metadata)
		assets, err := w.readFiles(basePath, files, w.fileReader)
		if err != nil {
			return err
		}

		request := &assethookv1alpha1.MutationRequest{
			Name:      asset.Name,
			Namespace: asset.Namespace,
			Assets:    assets,
			Metadata:  metadata,
		}
		url := w.getWebhookUrl(service)

		response, err := w.mutate(ctx, url, request)
		if err != nil {
			return err
		}

		if err := w.writeFiles(basePath, response.Assets); err != nil {
			return err
		}
	}

	return nil
}

func (w *mutationWebhook) mutate(ctx context.Context, url string, request *assethookv1alpha1.MutationRequest) (*assethookv1alpha1.MutationResponse, error) {
	context, cancel := context.WithTimeout(ctx, w.timeout)
	defer cancel()

	response := new(assethookv1alpha1.MutationResponse)
	err := w.webhook.Call(context, url, request, response)
	if err != nil {
		return nil, errors.Wrap(err, "while sending mutation request")
	}

	return response, nil
}
