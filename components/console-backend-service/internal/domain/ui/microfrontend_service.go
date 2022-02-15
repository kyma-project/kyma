package ui

import (
	uiV1alpha1v "github.com/kyma-project/kyma/common/microfrontend-client/pkg/apis/ui/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/ui/extractor"
	res "github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
	"k8s.io/client-go/tools/cache"
)

type microFrontendService struct {
	informer  cache.SharedIndexInformer
	extractor extractor.MFUnstructuredExtractor
}

func newMicroFrontendService(informer cache.SharedIndexInformer) *microFrontendService {
	return &microFrontendService{
		informer: informer,
	}
}

func (svc *microFrontendService) List(namespace string) ([]*uiV1alpha1v.MicroFrontend, error) {
	items, err := svc.informer.GetIndexer().ByIndex(cache.NamespaceIndex, namespace)
	if err != nil {
		return nil, err
	}

	var microFrontends []*uiV1alpha1v.MicroFrontend
	for _, item := range items {
		microFrontend, err := res.ToUnstructured(item)
		if err != nil {
			return nil, err
		}
		formattedMF, err := svc.extractor.FromUnstructured(microFrontend)
		if err != nil {
			return nil, err
		}
		microFrontends = append(microFrontends, formattedMF)
	}

	return microFrontends, nil
}
