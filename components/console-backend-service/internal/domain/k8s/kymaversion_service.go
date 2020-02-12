package k8s

import (
	"fmt"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/pretty"
	api "k8s.io/api/apps/v1"
	"k8s.io/client-go/tools/cache"
)

type kymaVersionService struct {
	informer cache.SharedIndexInformer
}

func newKymaVersionService(informer cache.SharedIndexInformer) (*kymaVersionService, error) {
	svc := &kymaVersionService{
		informer: informer,
	}
	return svc, nil
}

func (svc *kymaVersionService) FindDeployment(name, namespace string) (*api.Deployment, error) {
	key := fmt.Sprintf("%s/%s", namespace, name)

	item, exists, err := svc.informer.GetStore().GetByKey(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("%s %s not found in namespace `%s`", pretty.Deployment, name, namespace)
	}

	deployment, ok := item.(*api.Deployment)
	if !ok {
		return nil, fmt.Errorf("incorrect item type: %T, should be: *v1beta2.Deployment", item)
	}

	return deployment, nil
}
