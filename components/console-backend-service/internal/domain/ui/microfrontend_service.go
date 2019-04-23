package ui

import (
	"fmt"

	uiV1alpha1v "github.com/kyma-project/kyma/common/microfrontend-client/pkg/apis/ui/v1alpha1"
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

func (svc *microfrontendService) List() ([]*uiV1alpha1v.MicroFrontend, error) {
	items := svc.informer.GetStore().List()

	var microfrontends []*uiV1alpha1v.MicroFrontend
	for _, item := range items {
		microfrontend, ok := item.(*uiV1alpha1v.MicroFrontend)
		if !ok {
			return nil, fmt.Errorf("Incorrect item type: %T, should be: *Microfrontend", item)
		}
		microfrontends = append(microfrontends, microfrontend)
	}

	return microfrontends, nil
}
