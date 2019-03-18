package engine

import (
	"context"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/assethook"
	assethookv1alpha1 "github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/assethook/api/v1alpha1"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"time"
)

//go:generate mockery -name=Mutator -output=automock -outpkg=automock -case=underscore
type Mutator interface {
	Mutate(ctx context.Context, object Accessor, basePath string, files []string, services []v1alpha2.AssetWebhookService) error
}

type mutationEngine struct {
	baseEngine
	webhook    assethook.Webhook
	timeout    time.Duration
	fileReader func(filename string) ([]byte, error)
	fileWriter func(filename string, data []byte, perm os.FileMode) error
}

func NewMutator(webhook assethook.Webhook, timeout time.Duration) Mutator {
	return &mutationEngine{
		webhook:    webhook,
		timeout:    timeout,
		fileReader: ioutil.ReadFile,
		fileWriter: ioutil.WriteFile,
	}
}

func (e *mutationEngine) Mutate(ctx context.Context, object Accessor, basePath string, files []string, services []v1alpha2.AssetWebhookService) error {
	assetName := object.GetName()
	assetNamespace := object.GetNamespace()

	for _, service := range services {
		metadata := e.parseMetadata(service.Metadata)
		assets, err := e.readFiles(basePath, files, e.fileReader)
		if err != nil {
			return err
		}

		request := &assethookv1alpha1.MutationRequest{
			Name:      assetName,
			Namespace: assetNamespace,
			Assets:    assets,
			Metadata:  metadata,
		}
		url := e.getWebhookUrl(service)

		response, err := e.mutate(ctx, url, request)
		if err != nil {
			return err
		}

		if err := e.writeFiles(basePath, response.Assets, e.fileWriter); err != nil {
			return err
		}
	}

	return nil
}

func (e *mutationEngine) mutate(ctx context.Context, url string, request *assethookv1alpha1.MutationRequest) (*assethookv1alpha1.MutationResponse, error) {
	context, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	response := new(assethookv1alpha1.MutationResponse)
	err := e.webhook.Call(context, url, request, response)
	if err != nil {
		return nil, errors.Wrap(err, "while sending mutation request")
	}

	return response, nil
}
