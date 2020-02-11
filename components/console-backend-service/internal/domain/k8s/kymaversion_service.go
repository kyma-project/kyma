package k8s

import (
	"fmt"
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

func (svc *kymaVersionService) FindDeployment() (*api.Deployment, error) {
	key := fmt.Sprintf("%s/%s", "kyma-installer", "kyma-installer")

	item, exists, err := svc.informer.GetStore().GetByKey(key)
	if err != nil || !exists {
		return nil, err
	}

	deploy, ok := item.(*api.Deployment)
	if !ok {
		return nil, fmt.Errorf("incorrect item type: %T, should be: *v1beta2.Deployment", item)
	}

	return deploy, nil
}