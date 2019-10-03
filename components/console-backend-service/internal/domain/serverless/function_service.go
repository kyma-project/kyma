package serverless

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/serverless/convert"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type functionService struct {
	*resource.Service
}

func newFunctionService(service *resource.Service) *functionService {
	return &functionService{
		Service: service,
	}
}

func (svc *functionService) List(namespace string)([]*v1alpha1.Function, error) {
	items, err := svc.Informer.GetIndexer().ByIndex("namespace", namespace)
	if err != nil {
		return nil, err
	}

	var functions []*v1alpha1.Function
	for _, item := range items {
		function, err := convert.UnstructuredToFunction(item.(*unstructured.Unstructured))
		if err != nil {
			return nil, errors.Wrapf(err, "Incorrect item type: %T, should be: *Function", item)
		}
		functions = append(functions, function)
	}

	return functions, nil
}

func (svc *functionService) Delete(name string, namespace string) error {
	return svc.Client.Namespace(namespace).Delete(name, &metav1.DeleteOptions{})
}