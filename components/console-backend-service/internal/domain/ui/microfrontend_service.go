package ui

import (
	"fmt"

	uiV1alpha1v "github.com/kyma-project/kyma/common/microfrontend-client/pkg/apis/ui/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/ui/pretty"
	"k8s.io/client-go/tools/cache"
)

type microfrontendService struct {
	informer cache.SharedIndexInformer
}

func newMicrofrontendService(informer cache.SharedIndexInformer) *microfrontendService {
	return &microfrontendService{
		informer: informer,
	}
}

func (svc *microfrontendService) List(namespace string) ([]*uiV1alpha1v.MicroFrontend, error) {
	items, err := svc.informer.GetIndexer().ByIndex(cache.NamespaceIndex, namespace)
	if err != nil {
		return nil, err
	}

	var microfrontends []*uiV1alpha1v.MicroFrontend
	for _, item := range items {
		microfrontend, ok := item.(*uiV1alpha1v.MicroFrontend)
		if !ok {
			return nil, fmt.Errorf("Incorrect item type: %T, should be: *%s", item, pretty.MicroFrontend)
		}
		microfrontends = append(microfrontends, microfrontend)
	}

	return microfrontends, nil
}
