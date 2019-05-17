package ui

import (
	"fmt"

	uiV1alpha1v "github.com/kyma-project/kyma/common/microfrontend-client/pkg/apis/ui/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/ui/pretty"
	"k8s.io/client-go/tools/cache"
)

type microFrontendService struct {
	informer cache.SharedIndexInformer
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
		microFrontend, ok := item.(*uiV1alpha1v.MicroFrontend)
		if !ok {
			return nil, fmt.Errorf("Incorrect item type: %T, should be: *%s", item, pretty.MicroFrontend)
		}
		microFrontends = append(microFrontends, microFrontend)
	}

	return microFrontends, nil
}
