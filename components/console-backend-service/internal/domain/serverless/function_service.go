package serverless

import (
	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"k8s.io/client-go/tools/cache"
)

type functionService struct {
	informer cache.SharedIndexInformer
}

func newFunctionService(informer cache.SharedIndexInformer) *functionService {
	return &functionService{
		informer:informer,
	}
}

func (svc *functionService) List(namespace string)([]*v1alpha1.Function, error){
	items, err := svc.informer.GetIndexer().ByIndex("namespace", namespace)
	if err != nil {
		return nil, err
	}
	return items, nil
}