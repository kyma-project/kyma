package ui

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/ui/extractor"
	res "github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/apis/ui/v1alpha1"
	"k8s.io/client-go/tools/cache"
)

type backendModuleService struct {
	informer cache.SharedIndexInformer
	extractor extractor.BMUnstructuredExtractor
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
		backendModule, err := res.ToUnstructured(item)
		if err != nil {
			return nil, err
		}
		formattedBM, err := svc.extractor.FromUnstructured(backendModule)
		if err != nil {
			return nil, err
		}
		backendModules = append(backendModules, formattedBM)
	}

	return backendModules, nil
}
