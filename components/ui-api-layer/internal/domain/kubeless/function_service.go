package kubeless

import (
	"fmt"

	"github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"
	"k8s.io/client-go/tools/cache"
)

type functionService struct {
	informer cache.SharedIndexInformer
}

func newFunctionService(informer cache.SharedIndexInformer) *functionService {
	return &functionService{
		informer: informer,
	}

}

func (svc *functionService) List(environment string, pagingParams pager.PagingParams) ([]*v1beta1.Function, error) {
	items, err := pager.FromIndexer(svc.informer.GetIndexer(), "namespace", environment).Limit(pagingParams)
	if err != nil {
		return nil, err
	}

	var functions []*v1beta1.Function
	for _, item := range items {
		function, ok := item.(*v1beta1.Function)
		if !ok {
			return nil, fmt.Errorf("Incorrect item type: %T, should be: *Function", item)
		}

		functions = append(functions, function)
	}

	return functions, nil
}
