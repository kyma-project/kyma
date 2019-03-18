package ui

import (
	"fmt"

	"github.com/kyma-project/kyma/components/console-backend-service/pkg/apis/ui/v1alpha1"
	"k8s.io/client-go/tools/cache"
)

type backendModuleService struct {
	informer cache.SharedIndexInformer
}

func newBackendModuleService(informer cache.SharedIndexInformer) *backendModuleService {
	return &backendModuleService{
		informer: informer,
	}
}

func (svc *backendModuleService) List() ([]*v1alpha1.BackendModule, error) {
	items := svc.informer.GetStore().List()

	var backendModules []*v1alpha1.BackendModule
	for _, item := range items {
		backendModule, ok := item.(*v1alpha1.BackendModule)
		if !ok {
			return nil, fmt.Errorf("Incorrect item type: %T, should be: *BackendModule", item)
		}
		backendModules = append(backendModules, backendModule)
	}

	return backendModules, nil
}
